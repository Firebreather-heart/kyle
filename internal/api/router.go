package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/firebreather-heart/kyle/internal/api/middleware"
	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/docxgen"
	"github.com/firebreather-heart/kyle/internal/identity"
	"github.com/firebreather-heart/kyle/internal/llm"
	"github.com/firebreather-heart/kyle/internal/models"
	"github.com/firebreather-heart/kyle/internal/orchestrator"
)

type FileStore interface {
	UploadFile(ctx context.Context, localPath string, remoteName string) (string, error)
}

type Router struct {
	idService identity.Service
	cfg       *config.AppConfig
	fileStore FileStore
}

func NewRouter(idService identity.Service, cfg *config.AppConfig, fileStore FileStore) *Router {
	return &Router{
		idService: idService,
		cfg:       cfg,
		fileStore: fileStore,
	}
}

func (r *Router) generateSecureToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("failed to generate secure token: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (r *Router) SetupRoutes(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("POST /api/v1/obtain-token", r.handleObtainCookieToken)
	mux.HandleFunc("GET /api/v1/triangulate", r.handleTriangulate)
	mux.HandleFunc("POST /api/v1/generate", r.handleGenerate)
	mux.HandleFunc("GET /api/v1/tasks/{id}", r.handleGetTaskStatus)
	mux.HandleFunc("GET /api/v1/download/{id}", r.handleDownload)

	var handler http.Handler = middleware.Logger(mux)
	handler = middleware.CORS(handler)
	handler = middleware.Authenticator(r.idService)(handler)
	handler = middleware.RateLimiter(r.idService)(handler)
	return handler
}

func (r *Router) handleObtainCookieToken(w http.ResponseWriter, req *http.Request) {
	var body struct {
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.sendJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	token := r.generateSecureToken(16)
	user, _ := identity.NewUser(body.Fingerprint, token)
	if err := r.idService.CreateOrUpdateUser(req.Context(), *user, 24*time.Hour); err != nil {
		log.Printf("Error creating user: %v", err)
		r.sendJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "kyle_id",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   31536000,
	})
	r.sendJSON(w, http.StatusOK, user)

}

func (r *Router) handleTriangulate(w http.ResponseWriter, req *http.Request) {
	user := middleware.GetUser(req)
	if user == nil {
		r.sendJSON(w, http.StatusUnauthorized, map[string]string{"error": "unknown user"})
		return
	}
	r.sendJSON(w, http.StatusOK, user)
}

func (r *Router) handleGenerate(w http.ResponseWriter, req *http.Request) {
	user := middleware.GetUser(req)
	allowed, err := r.idService.AllowLLMCall(req.Context(), user.Fingerprint, user.ShadowCookie)
	if err != nil {
		log.Printf("Error checking LLM call allowance: %v", err)
		r.sendJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
		return
	}
	if !allowed {
		r.sendJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
		return
	}
	var modelRequest models.ClientRequest
	if err := json.NewDecoder(req.Body).Decode(&modelRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if modelRequest.Topic == "" {
		http.Error(w, "Topic is required", http.StatusBadRequest)
		return
	}

	log.Printf("Received research request for %s topic: %s", modelRequest.Provider, modelRequest.Topic)

	engine, err := llm.NewProvider(r.cfg, modelRequest.Provider)
	if err != nil {
		log.Printf("Error creating LLM provider: %v", err)
		r.sendJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid provider: %s", modelRequest.Provider)})
		return
	}

	agent := orchestrator.NewAgent(engine)
	taskID := r.generateSecureToken(8)

	go func() {
		ctx := context.Background()
		report := func(msg string) {
			r.idService.PublishTaskUpdate(ctx, taskID, msg)
		}

		result := agent.Run(modelRequest.Topic, report)

		if result.Status == "success" && result.Document != nil {
			localFilename := fmt.Sprintf("research_%s.docx", strings.ReplaceAll(strings.ToLower(modelRequest.Topic), " ", "_"))
			if err := docxgen.GenerateWordDoc(localFilename, result.Document); err != nil {
				log.Printf("Error generating Word doc: %v", err)
				report("Error: Pipeline succeeded but Word doc generation failed")
				return
			}
			log.Printf("Local Word doc generated: %s", localFilename)

			report("Uploading to cloud storage...")
			cloudURL, err := r.fileStore.UploadFile(ctx, localFilename, taskID)
			if err != nil {
				log.Printf("Cloud upload failed: %v", err)
				report("Error: Cloud storage upload failed")
				return
			}

			log.Printf("Document uploaded to cloud: %s", cloudURL)
			report("COMPLETE:" + cloudURL)
		} else {
			report("ERROR:" + result.Message)
		}
	}()

	r.sendJSON(w, http.StatusAccepted, map[string]string{
		"status":  "accepted",
		"task_id": taskID,
	})

	r.idService.RecordLLMUsage(req.Context(), user.Fingerprint, user.ShadowCookie)
}

func (r *Router) handleGetTaskStatus(w http.ResponseWriter, req *http.Request) {
	taskID := req.PathValue("id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)
	err := rc.Flush()
	if err != nil {
		log.Printf("Error flushing response: %v", err)
		return
	}

	pubsub := r.idService.SubscribeTask(req.Context(), taskID)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-req.Context().Done():
			return
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
			err := rc.Flush()
			if err != nil {
				log.Printf("Error flushing response: %v", err)
				return
			}
			if strings.HasPrefix(msg.Payload, "COMPLETE:") || strings.HasPrefix(msg.Payload, "ERROR:") {
				return
			}
		}
	}
}

func (r *Router) handleDownload(w http.ResponseWriter, req *http.Request) {
	// Logic for serving the .docx file
}

func (r *Router) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
