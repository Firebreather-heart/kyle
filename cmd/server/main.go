package main

import (
	"log"

	"github.com/firebreather-heart/kyle/internal/api"
	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/server"
	"github.com/firebreather-heart/kyle/internal/store"
	"net/http"
)

// @title Kyle AI Research API
// @version 1.0
// @description High-performance AI document generation and identity resolution API.
// @host localhost:8080
// @BasePath /api/v1

func main() {
	cfg := config.Load()

	log.Printf("DEBUG: Final Config - RedisAddr: %s, ServerPort: %s", cfg.RedisAddr, cfg.ServerPort)

	idStore, err := store.NewRedisStore(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	log.Println("Redis connected at", cfg.RedisAddr)

	cldStore, err := store.NewCloudinaryStore(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	log.Println("Cloudinary storage ready")

	router := api.NewRouter(idStore, cfg, cldStore)
	mux := http.NewServeMux()
	handler := router.SetupRoutes(mux)
	srv := server.NewServer(cfg.ServerPort, handler)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
}
