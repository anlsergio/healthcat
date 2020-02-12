package server

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
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
}

func (s *Server) Run() {
	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt)

	httpServer := http.Server{
		Addr: s.Address,
		Handler: &router{
			reporter: s.Checker,
		},
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

type router struct {
	reporter StateReporter
}

func (s router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		state := s.reporter.State()
		if state.Healthy {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "OK\n")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failure\n")
		}
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK\n")
	})

	mux.HandleFunc("/targets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
				return
			}
			target := string(body)
			s.reporter.Add(target)
		} else if r.Method == http.MethodDelete {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
				return
			}
			service := string(body)
			s.reporter.Delete(service)
		}
	})

	mux.ServeHTTP(w, r)
}
