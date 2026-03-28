package server

import (
	// "encoding/json"
	// "log"
	// "net/http"

	// "github.com/firebreather-heart/kyle/internal/llm"
	// "github.com/firebreather-heart/kyle/internal/models"
	"github.com/firebreather-heart/kyle/internal/config"
)

type Server struct {
	cfg *config.AppConfig
}

func NewServer(cfg *config.AppConfig) *Server{
	return &Server{cfg: cfg}
}

