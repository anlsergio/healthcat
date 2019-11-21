package server

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
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("HealthEnpoint:%v", c.healthy), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/status", nil)
			response := httptest.NewRecorder()

			server := &ServerRouter{}
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

	server := &ServerRouter{}
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
