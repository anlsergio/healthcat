package checker

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEmptyIsHealthy(t *testing.T) {
	checker := New("test", 1*time.Second, 1, 1, 100)

	if !checker.Healthy() {
		t.Error("empty checker must report healthy status")
	}
}

func TestHealthyEndpoint(t *testing.T) {
	checker := New("test", 1*time.Second, 1, 1, 100)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "OK\n")
	}))
	defer server.Close()

	checker.Add(server.URL)

	time.Sleep(2 * time.Second)

	if !checker.Healthy() {
		t.Error("checker must be healthy")
	}
}

func TestUnhealthyEndpoint(t *testing.T) {
	checker := New("test", 1*time.Second, 1, 1, 100)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker.Add(server.URL)

	time.Sleep(2 * time.Second)

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
