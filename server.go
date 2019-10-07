package main

import (
	"fmt"
	"net/http"
)

type HealthCheckServer struct {
	checker HealthChecker
}

func (s HealthCheckServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		var label string
		if s.checker.Healthy() {
			label = "ok"
		} else {
			label = "failed"
		}
		fmt.Fprintf(w, "%s\n", label)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.ServeHTTP(w, r)
}
