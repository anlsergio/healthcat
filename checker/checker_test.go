package checker

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEmptyIsHealthy(t *testing.T) {
	checker := New(1*time.Second, 1, 1, 100)

	if !checker.State().Healthy {
		t.Error("empty checker must report healthy status")
	}
}

func TestHealthyEndpoint(t *testing.T) {
	checker := New(1*time.Second, 1, 1, 100)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "OK\n")
	}))
	defer server.Close()

	checker.Add(server.URL)

	time.Sleep(2 * time.Second)

	if !checker.State().Healthy {
		t.Error("checker must be healthy")
	}
}

func TestUnhealthyEndpoint(t *testing.T) {
	checker := New(1*time.Second, 1, 1, 100)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker.Add(server.URL)

	time.Sleep(2 * time.Second)

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
