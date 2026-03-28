package main

import (
	"log"
	"net/http"

	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/server"
)

func main() {
	cfg := config.Load()

	api := server.NewServer(cfg)

	http.HandleFunc("/api/research", api.HandleResearch)

	log.Printf("Kyle AI Engine booting sequence initialized...")
	log.Printf("Listening on port :%s", cfg.ServerPort)

	err := http.ListenAndServe(":"+cfg.ServerPort, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}