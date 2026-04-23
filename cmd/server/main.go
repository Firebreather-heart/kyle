package main

import (
	"log"
	"os"

	"github.com/firebreather-heart/kyle/internal/api"
	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/server"
	"github.com/firebreather-heart/kyle/internal/store"
	"net/http"
)

func main() {
	cfg := config.Load()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	idStore, err := store.NewRedisStore(redisAddr)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	log.Println("Redis connected at", redisAddr)

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
