package checker

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEmptyIsHealthy(t *testing.T) {
	checker := &Checker{
		ClusterID:        "abc",
		Interval:         1 * time.Second,
		FailureThreshold: 1,
		SuccessThreshold: 1,
		StateThreshold:   100,
	}
	if err := checker.Run(); err != nil {
		t.Errorf("got error %v", err)
		return
	}
	defer checker.Stop()

	if !checker.Healthy() {
		t.Error("empty checker must report healthy status")
	}
}

func TestHealthyEndpoint(t *testing.T) {
	checker := &Checker{
		ClusterID:        "abc",
		Interval:         1 * time.Second,
		FailureThreshold: 1,
		SuccessThreshold: 1,
		StateThreshold:   100,
	}
	checker.updates = make(chan struct{}, 1)
	if err := checker.Run(); err != nil {
		t.Errorf("got error %v", err)
		return
	}
	defer checker.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "OK\n")
	}))
	defer server.Close()

	checker.Add("test", server.URL)
	<-checker.updates

	if !checker.Healthy() {
		t.Error("checker must be healthy")
	}
}

func TestUnhealthyEndpoint(t *testing.T) {
	checker := &Checker{
		ClusterID:        "abc",
		Interval:         1 * time.Second,
		FailureThreshold: 1,
		SuccessThreshold: 1,
		StateThreshold:   100,
	}
	checker.updates = make(chan struct{}, 1)
	if err := checker.Run(); err != nil {
		t.Errorf("got error %v", err)
		return
	}
	defer checker.Stop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker.Add("test", server.URL)
	<-checker.updates

	if checker.Healthy() {
		t.Error("checker must be unhealthy")
	}
}

func TestCalcTimeout(t *testing.T) {
	interval := 10 * time.Second

	if want, got := 8*time.Second, calcTimeout(interval); got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestClusterID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"", false},
		{strings.Repeat("a", 64), false},
		{"123", true},
		{"abc", true},
		{"www.mydomain.com", true},
		{"wiley-dev-us-east-1", true},
		{"us-east-1.dev.edpub.wiley.com", true},
		{" abc", false},
		{"abc ", false},
		{"Abc", false},
		{"a b", false},
		{"12", false},
		{"😄", false},
		{"-www.mydomain.com", false},
		{"www.mydomain.com-", false},
		{".www.mydomain.com", false},
		{"www.mydomain.com.", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			c := &Checker{
				ClusterID: tt.id,
			}
			err := c.Run()
			c.Stop()

			if tt.valid != (err == nil) {
				if err == nil {
					t.Errorf("want error")
				} else {
					t.Errorf("got error %v", err)
				}
			}
		})
	}
}

func TestHealthStatus(t *testing.T) {
	cases := []struct {
		name      string
		total     int
		healthy   int
		threshold int
		status    bool
	}{
		{"Empty", 0, 0, 100, true},
		{"AllHealthy", 10, 10, 100, true},
		{"NoneHealthy", 10, 0, 100, false},
		{"HealthyEqualToThreshold", 10, 5, 50, true},
		{"HealthyBelowThreshold", 10, 4, 50, false},
		{"HealthyAboveThreshold", 10, 6, 50, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if want, got := c.status, calcHealthStatus(c.total, c.healthy, c.threshold); want != got {
				t.Errorf("Want status %t, got %t", want, got)
			}
		})
	}
}
