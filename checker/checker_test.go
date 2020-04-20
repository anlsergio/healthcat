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

	if !checker.State().Healthy {
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

	checker.Add(server.URL)
	<-checker.updates

	if !checker.State().Healthy {
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

	checker.Add(server.URL)
	<-checker.updates

	if checker.State().Healthy {
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
		{strings.Repeat("a", 31), false},
		{"12", false},
		{"123", true},
		{"abc", true},
		{"a b", false},
		{"😄",  false},
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
