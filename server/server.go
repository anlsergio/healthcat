package server

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"wiley.com/healthcat/checker"
	chczap "wiley.com/healthcat/logger"
	"wiley.com/healthcat/version"
)

// Server properties
// TODO: add better descriptipn
type Server struct {
	Address string
	Checker *checker.Checker
	Logger  *zap.Logger
}

// StateReporter methods
type StateReporter interface {
	Add(name, url string)
	Delete(url string)
	State() checker.ClusterState
	Healthy() bool
	Ready() bool
}

// Run HTTP server
// TODO: add better descriptipn
func (s *Server) Run() {
	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt)

	logger := s.Logger.Sugar()

	httpServer := http.Server{
		Addr:         s.Address,
		Handler:      router(s.Checker, s.Logger),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		<-interrupted
		logger.Info("Stopping CHC")
		s.Checker.Stop()
		httpServer.Shutdown(context.Background())
	}()

	logger.Infof("Starting CHC %s on %s", version.Version, s.Address)

	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

//
// HTTP router
// TODO: add better descriptipn
func router(sr StateReporter, log *zap.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(chczap.Chczap(log, time.RFC3339, true))
	r.Use(chczap.RecoveryWithZap(log, false))

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
			return
		}
		target := string(body)
		sr.Add(target, target)
	})

	r.Delete("/services", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		service := string(body)
		sr.Delete(service)
	})

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		err := encoder.Encode(version.ChcVer)
		if err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
		}
	})

	return r
}
