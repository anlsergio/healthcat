package checker

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
)

// ClusterState describes the current cluster state
type ClusterState struct {
	ActiveCount  int  // Number of active services
	HealthyCount int  // Number of healthy services
	Healthy      bool // Cluster healthy state
}

// Checker periodically checks availability of targets in the list
type Checker struct {
	ClusterID        string
	Interval         time.Duration
	FailureThreshold int
	SuccessThreshold int
	StateThreshold   int

	done chan struct{}
	mux  sync.Mutex

	client       *http.Client
	targets      map[string]*target
	activeCount  int
	healthyCount int
	added        chan string
	deleted      chan string
	reports      chan *report
	accessors    chan accessor
	ready        bool
	updates      chan struct{}
}

type accessor func(c *Checker)

type target struct {
	url        string
	healthy    bool
	done       chan struct{}
	lastReport *report

	// state represents the current state of the target
	// if positive, it contains the number of consecutive successful checks;
	// if negative, it contains the (negative) number of consecutive failed checks
	state int64
}

// Starts the checker
func (c *Checker) Run() error {
	if err := c.validate(); err != nil {
		return err
	}

	c.mux.Lock()
	c.done = make(chan struct{})
	c.mux.Unlock()

	c.targets = make(map[string]*target)
	c.reports = make(chan *report)
	c.added = make(chan string)
	c.deleted = make(chan string)
	c.accessors = make(chan accessor)
	c.client = &http.Client{
		Timeout: calcTimeout(c.Interval),
	}
	c.ready = true

	go c.run()
	return nil
}

func (c *Checker) validate() error {
	if ok, err := regexp.MatchString("^[[:alnum:]]{3,30}$", c.ClusterID); !ok {
		if err != nil {
			panic(err)
		}
		return errors.New("Cluster ID must be alphanumeric string of length between 3 an 30 inclusive")
	}
	return nil
}

func calcTimeout(interval time.Duration) time.Duration {
	return time.Duration(float64(interval) * 0.8)
}

// Stop stops the checker
func (c *Checker) Stop() {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.done != nil {
		close(c.done)
	}
}

// State reports about the current cluster state
func (c *Checker) State() ClusterState {
	result := make(chan *ClusterState, 1)
	c.accessors <- func(c *Checker) {
		healthy := true
		if c.activeCount > 0 {
			healthy = c.healthyCount*100/c.activeCount >= c.StateThreshold
		}
		result <- &ClusterState{
			ActiveCount:  c.activeCount,
			HealthyCount: c.healthyCount,
			Healthy:      healthy,
		}
	}
	return *<-result
}

// Ready gets the current readiness status
func (c *Checker) Ready() bool {
	result := make(chan bool, 1)
	c.accessors <- func(c *Checker) {
		result <- c.ready
	}
	return <-result
}

// Add adds the given target to the check list
func (c *Checker) Add(url string) {
	if url == "" {
		log.Println("Attempt to add empty target")
		return
	}
	c.added <- url
}

// Delete removes the given target from the check list
func (c *Checker) Delete(url string) {
	if url == "" {
		log.Println("Attempt to remove empty target")
		return
	}
	c.deleted <- url
}

func (c *Checker) addTarget(url string) {
	if _, ok := c.targets[url]; ok {
		log.Printf("Attempt to add already added target %s\n", url)
		return
	}
	log.Printf("Adding target %s", url)
	t := &target{
		url:  url,
		done: make(chan struct{}),
	}
	c.targets[url] = t
	go c.newTargetLoop(url, t.done)
}

func (c *Checker) deleteTarget(url string) {
	t, ok := c.targets[url]
	if !ok {
		log.Printf("Attempt to delete unregistered target %s\n", url)
		return
	}

	close(t.done)
	delete(c.targets, url)

	if t.state != 0 {
		c.activeCount--
		if t.healthy {
			c.healthyCount--
		}
	}
}

func (c *Checker) update(r *report) {
	t, ok := c.targets[r.url]
	if !ok {
		log.Printf("Received report from unregistered target %s\n", r.url)
		return
	}
	if t.state == 0 {
		c.activeCount++
	}
	t.lastReport = r

	if r.err == nil {
		if t.state < 0 {
			t.state = 0
		}
		t.state++
		if t.state >= int64(c.SuccessThreshold) && !t.healthy {
			t.healthy = true
			c.healthyCount++
		}
	} else {
		if t.state > 0 {
			t.state = 0
		}
		t.state--
		if t.state <= int64(-c.FailureThreshold) && t.healthy {
			t.healthy = false
			c.healthyCount--
		}
	}

	log.Printf("Report from %s: s:%d, h:%t, err:%v\n", t.url, t.state, t.healthy, r.err)
	if c.updates != nil {
		select {
		case c.updates <- struct{}{}:
		default:
		}
	}
}

func (c *Checker) reportState(results chan<- *ClusterState) {

}

func (c *Checker) run() {
Loop:
	for {
		select {
		case url := <-c.added:
			c.addTarget(url)
		case url := <-c.deleted:
			c.deleteTarget(url)
		case r := <-c.reports:
			c.update(r)
		case a := <-c.accessors:
			a(c)
		case <-c.done:
			log.Println("Stopping all target loops")
			for _, c := range c.targets {
				close(c.done)
			}
			break Loop
		}
	}
}

func (c *Checker) newTargetLoop(url string, done <-chan struct{}) {
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	checkURL := fmt.Sprintf("%s%s", url, "/healthz")
Loop:
	for {
		ts := time.Now()
		resp, err := c.client.Get(checkURL)
		if err == nil {
			resp.Body.Close() // TODO: Do we need to drain the body before closing?
			if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("Status %d", resp.StatusCode)
			}
		}

		c.reports <- &report{
			url: url,
			ts:  ts,
			err: err,
		}

		select {
		case <-done:
			break Loop
		case <-ticker.C:
		}
	}
}

type report struct {
	url string
	ts  time.Time
	err error
}
