package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStatus(t *testing.T) {
	cases := []struct {
		healthy bool
		label   string
	}{
		{true, "ok\n"},
		{false, "failed\n"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("HealthEnpoint:%v", c.healthy), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/status", nil)
			response := httptest.NewRecorder()

			server := &HealthCheckServer{MockHealthChecker(c.healthy)}
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

	server := &HealthCheckServer{nil}
	server.ServeHTTP(response, request)

	statusGot := response.Result().StatusCode
	statusWant := http.StatusOK

	if statusGot != statusWant {
		t.Errorf("got status %d, want %d", statusGot, statusWant)
	}

	responseGot := response.Body.String()
	responseWant := "OK"

	if responseGot != responseWant {
		t.Errorf("Got response string %q, want %q", responseGot, responseWant)
	}
}

type MockHealthChecker bool

func (m MockHealthChecker) Healthy() bool {
	return bool(m)
}
