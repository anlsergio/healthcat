package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt)

	var address string
	flag.StringVar(&address, "l", ":8080", "server listen address")
	flag.Parse()

	s := http.Server{
		Addr: address,
		Handler: &HealthCheckServer{
			checker: &ClusterHealthChecker{},
		},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		<-interrupted
		log.Printf("Stopping server")
		s.Shutdown(context.Background())
	}()

	log.Printf("Starting server on %s", address)
	err := s.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
