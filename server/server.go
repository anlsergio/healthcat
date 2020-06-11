package server

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"wiley.com/do-k8s-cluster-health-check/checker"
)

type Server struct {
	Address string
	Checker *checker.Checker
}

type StateReporter interface {
	Add(url string)
	Delete(url string)
	State() checker.ClusterState
	Healthy() bool
	Ready() bool
}

func (s *Server) Run() {
	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt)

	httpServer := http.Server{
		Addr:         s.Address,
		Handler:      router(s.Checker),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		<-interrupted
		log.Printf("Stopping server")
		s.Checker.Stop()
		httpServer.Shutdown(context.Background())
	}()

	log.Printf("Starting server on %s", s.Address)
	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func router(sr StateReporter) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		if sr.Healthy() {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "OK\n")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failure\n")
		}
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "OK\n")
	})

	r.Get("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		if sr.Ready() {
			io.WriteString(w, "OK\n")
		} else {
			w.WriteHeader(552)
			io.WriteString(w, "Not ready\n")
		}
	})

	r.Get("/services", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		err := encoder.Encode(sr.State())
		if err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
		}
	})

	r.Post("/services", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			return
		}
		target := string(body)
		sr.Add(target)
	})

	r.Delete("/services", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			return
		}
		service := string(body)
		sr.Delete(service)
	})
	return r
}
