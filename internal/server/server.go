package server

import (
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	"strings"

	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/llm"
	"github.com/firebreather-heart/kyle/internal/models"
	"github.com/firebreather-heart/kyle/internal/orchestrator"
	"github.com/firebreather-heart/kyle/internal/docxgen"
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

	if s.Config.KIMI_API_KEY == "" {
		http.Error(w, "Server configuration error: Missing Kimi API Key", http.StatusInternalServerError)
		return
	}

	engine := llm.NewKIMIClient(s.Config.KIMI_API_KEY)
	agent := orchestrator.NewAgent(engine)
	
	result := agent.Run(req.Topic)

	if result.Status == "success" && result.Document != nil {
		wordFilename := fmt.Sprintf("research_%s.docx", strings.ReplaceAll(strings.ToLower(req.Topic), " ", "_"))
		if err := docxgen.GenerateWordDoc(wordFilename, result.Document); err != nil {
			log.Printf("Error generating Word doc: %v", err)
			http.Error(w, fmt.Sprintf("Pipeline succeeded but Word doc generation failed: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Word doc generated: %s", wordFilename)
	}

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