package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"wiley.com/do-k8s-cluster-health-check/checker"
)

type testReporter bool

func (r testReporter) State() checker.ClusterState {
	return checker.ClusterState{
		Healthy: bool(r),
	}
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

			reporter := testReporter(c.healthy)

			server := &router{reporter: reporter}
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

	server := &router{}
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
