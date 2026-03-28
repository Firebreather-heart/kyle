package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/llm"
	"github.com/firebreather-heart/kyle/internal/models"
	"github.com/firebreather-heart/kyle/internal/orchestrator"
)

type APIServer struct {
	Config *config.AppConfig
}

func NewServer(cfg *config.AppConfig) *APIServer {
	return &APIServer{
		Config: cfg,
	}
}

func (s *APIServer) HandleResearch(w http.ResponseWriter, r *http.Request) {
	// CORS Headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Topic == "" {
		http.Error(w, "Topic is required", http.StatusBadRequest)
		return
	}

	log.Printf("Received research request for topic: %s", req.Topic)

	if s.Config.GEMINI_API_KEY == "" {
		http.Error(w, "Server configuration error: Missing API Key", http.StatusInternalServerError)
		return
	}

	engine := llm.NewGeminiClient(s.Config.GEMINI_API_KEY)
	agent := orchestrator.NewAgent(engine)
	
	result := agent.Run(req.Topic)

	w.Header().Set("Content-Type", "application/json")
	if result.Status == "substandard" {
		w.WriteHeader(http.StatusPartialContent)
	} else if result.Status == "error" {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(result)
}