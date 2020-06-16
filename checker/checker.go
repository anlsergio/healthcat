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

type Cluster struct {
	Name    string `json:"name"`    // The cluster name (ID)
	Healthy bool   `json:"healthy"` // Health status
	Total   int    `json:"total"`   // Total monitored services
	Failed  int    `json:"failed"`  // Failed services
}

type Service struct {
	Name    string `json:"name"`    // The cluster name (ID)
	Healthy bool   `json:"healthy"` // Cluster healthy state
}

// ClusterState describes the current cluster state
type ClusterState struct {
	Cluster  Cluster   `json:"cluster"`
	Services []Service `json:"services"`
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
	healthy      bool
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
	c.healthy = true
	c.ready = true

	go c.run()
	return nil
}

func (c *Checker) validate() error {
	pattern := "^[a-zA-Z0-9\\-\\.]{3,63}$"
	if ok, err := regexp.MatchString(pattern, c.ClusterID); !ok {
		if err != nil {
			panic(err)
		}
		return errors.New("Cluster ID must be alphanumeric string of length between 3 an 63 inclusive ('.' and '-' are the only special characters allowed)")
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
		cs := &ClusterState{
			Cluster: Cluster{
				Name:    c.ClusterID,
				Healthy: c.healthy,
				Total:   c.activeCount,
				Failed:  c.activeCount - c.healthyCount,
			},
			Services: make([]Service, 0, c.activeCount),
		}
		for k, v := range c.targets {
			if v.state != 0 {
				cs.Services = append(cs.Services, Service{
					Name:    k,
					Healthy: v.healthy,
				})
			}
		}
		result <- cs
	}
	return *<-result
}

// Healthy returns current cluster health state
func (c *Checker) Healthy() bool {
	result := make(chan bool, 1)
	c.accessors <- func(c *Checker) {
		result <- c.healthy
	}
	return <-result
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
		c.updateHealthStatus()
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

	c.updateHealthStatus()
	log.Printf("Report from %s: s:%d, h:%t, err:%v\n", t.url, t.state, t.healthy, r.err)
	if c.updates != nil {
		select {
		case c.updates <- struct{}{}:
		default:
		}
	}
}

func (c *Checker) updateHealthStatus() {
	c.healthy = calcHealthStatus(c.activeCount, c.healthyCount, c.StateThreshold)
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

func calcHealthStatus(total, healthy, threshold int) bool {
	status := true
	if total > 0 {
		status = healthy*100/total >= threshold
	}
	return status
}
