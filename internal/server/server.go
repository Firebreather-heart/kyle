package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type APIServer struct {
	addr    string
	handler http.Handler
}

func NewServer(port string, handler http.Handler) *APIServer {
	return &APIServer{
		addr:    ":" + port,
		handler: handler,
	}
}

func (s *APIServer) Run() error {
	server := &http.Server{
		Addr:    s.addr,
		Handler: s.handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Kyle AI server starting on %s", s.addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}