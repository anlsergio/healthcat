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
			request, _ := http.NewRequest(http.MethodGet, "/status", nil)
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

type MockHealthChecker bool

func (m MockHealthChecker) Healthy() bool {
	return bool(m)
}
