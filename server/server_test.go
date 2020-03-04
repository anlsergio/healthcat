package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"wiley.com/do-k8s-cluster-health-check/checker"
)

type testReporter struct {
	healthy bool
	ready   bool
	state   checker.ClusterState
}

func (r testReporter) State() checker.ClusterState {
	return checker.ClusterState{}
}

func (r testReporter) Healthy() bool {
	return r.healthy
}

func (r testReporter) Ready() bool {
	return r.ready
}

func (r testReporter) Add(url string)    {}
func (r testReporter) Delete(url string) {}

func TestGetStatus(t *testing.T) {
	cases := []struct {
		healthy bool
		label   string
	}{
		{true, "OK\n"},
		{false, "Failure\n"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("HealthEnpoint:%v", c.healthy), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/status", nil)
			response := httptest.NewRecorder()

			reporter := testReporter{healthy: c.healthy}

			server := router(reporter)
			server.ServeHTTP(response, request)

			got := response.Body.String()
			want := c.label

			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

func TestHealthz(t *testing.T) {
	request := httptest.NewRequest("", "/healthz", nil)
	response := httptest.NewRecorder()

	server := router(testReporter{})
	server.ServeHTTP(response, request)

	statusGot := response.Result().StatusCode
	statusWant := http.StatusOK

	if statusGot != statusWant {
		t.Errorf("got status %d, want %d", statusGot, statusWant)
	}

	responseGot := response.Body.String()
	responseWant := "OK\n"

	if responseGot != responseWant {
		t.Errorf("Got response string %q, want %q", responseGot, responseWant)
	}
}

func TestReadiness(t *testing.T) {
	cases := []struct {
		name    string
		ready   bool
		status  int
		message string
	}{
		{"Ready", true, 200, "OK\n"},
		{"NotReady", false, 552, "Not ready\n"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/healthz/ready", nil)
			resp := httptest.NewRecorder()

			server := router(testReporter{ready: c.ready})
			server.ServeHTTP(resp, req)

			if want, got := c.status, resp.Result().StatusCode; want != got {
				t.Errorf("Want status %d, got %d", want, got)
			}

			if want, got := c.message, resp.Body.String(); want != got {
				t.Errorf("Want message %q, got %q", want, got)
			}
		})
	}
}

func TestServices(t *testing.T) {
	req := httptest.NewRequest("", "/healthz/services", nil)
	resp := httptest.NewRecorder()

	state := checker.ClusterState{
		Cluster: checker.Cluster{
			Name:    "c1",
			Healthy: false,
			Total:   2,
			Failed:  1,
		},
		Services: []checker.Service{
			{Name: "s1", Healthy: true},
			{Name: "s2", Healthy: false},
		},
	}
	server := router(testReporter{state: state})
	server.ServeHTTP(resp, req)

	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&state); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}
}
