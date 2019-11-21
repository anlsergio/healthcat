package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Server struct {
	Address string
}

func (s *Server) Run() {
	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt)

	httpServer := http.Server{
		Addr:         s.Address,
		Handler:      &ServerRouter{},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		<-interrupted
		log.Printf("Stopping server")
		httpServer.Shutdown(context.Background())
	}()

	log.Printf("Starting server on %s", s.Address)
	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
