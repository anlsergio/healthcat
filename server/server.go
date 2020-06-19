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
	"wiley.com/do-k8s-cluster-health-check/checker"
	chczap "wiley.com/do-k8s-cluster-health-check/logger"
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
	Add(url string)
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

	sugaredLogger := s.Logger.Sugar()

	httpServer := http.Server{
		Addr:         s.Address,
		Handler:      router(s.Checker, s.Logger),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		<-interrupted
		sugaredLogger.Info("Stopping server")
		s.Checker.Stop()
		httpServer.Shutdown(context.Background())
	}()

	sugaredLogger.Infof("Starting server on %s", s.Address)
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
		sr.Add(target)
	})

	r.Delete("/services", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		service := string(body)
		sr.Delete(service)
	})
	return r
}

//
// Clones the logger with new HTTP request fields
//
// TODO: Consider addding additional fields
func getHTTPLogger(log *zap.SugaredLogger) func(*http.Request) *zap.SugaredLogger {
	return func(r *http.Request) *zap.SugaredLogger {
		return log.With(
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
	}
}
