package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/firebreather-heart/kyle/internal/api/middleware"
	"github.com/firebreather-heart/kyle/internal/config"
	"github.com/firebreather-heart/kyle/internal/docxgen"
	"github.com/firebreather-heart/kyle/internal/identity"
	"github.com/firebreather-heart/kyle/internal/llm"
	"github.com/firebreather-heart/kyle/internal/models"
	"github.com/firebreather-heart/kyle/internal/orchestrator"
	_ "github.com/firebreather-heart/kyle/docs" // Swagger docs
	httpSwagger "github.com/swaggo/http-swagger/v2"
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
	mux.HandleFunc("GET /api/v1/download/{id}/pdf", r.handleDownloadPDF)
	mux.HandleFunc("GET /api/v1/download/{id}", r.handleDownload)
	mux.Handle("GET /docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))

	var handler http.Handler = middleware.Logger(mux)
	handler = middleware.CORS(handler)
	handler = middleware.Authenticator(r.idService)(handler)
	handler = middleware.RateLimiter(r.idService)(handler)
	return handler
}

// handleObtainCookieToken godoc
// @Summary Obtain a session token
// @Description Creates or retrieves a user identity based on fingerprint and sets a secure cookie.
// @Tags Identity
// @Accept json
// @Produce json
// @Param body body models.TokenRequest true "Fingerprint"
// @Success 200 {object} identity.User
// @Router /obtain-token [post]
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
	status, _ := r.idService.GetSystemStatus(req.Context())
	r.sendJSON(w, http.StatusOK, map[string]interface{}{
		"user":          user,
		"system_status": status,
	})

}

func (r *Router) handleTriangulate(w http.ResponseWriter, req *http.Request) {
	fp := req.Header.Get("X-Fingerprint")
	var cookieVal string
	if c, err := req.Cookie("kyle_id"); err == nil {
		cookieVal = c.Value
	}

	user, _ := r.idService.Triangulate(req.Context(), fp, cookieVal)
	status, _ := r.idService.GetSystemStatus(req.Context())

	r.sendJSON(w, http.StatusOK, map[string]interface{}{
		"user":          user,
		"system_status": status,
	})
}

// handleGenerate godoc
// @Summary Start AI Research Task
// @Description Initiates an asynchronous research task and returns a task ID.
// @Tags AI
// @Accept json
// @Produce json
// @Param request body models.ClientRequest true "Research Topic & Provider"
// @Success 202 {object} map[string]string "accepted, task_id"
// @Router /generate [post]
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

	// Initialize with active key if multi-key rotation is supported for this provider
	providerKeys := r.cfg.GetProviderKeys(modelRequest.Provider)
	if len(providerKeys) > 0 {
		activeKey, _ := r.idService.GetActiveKey(req.Context(), modelRequest.Provider, providerKeys)
		if activeKey != "" {
			engine.UpdateAPIKey(activeKey)
		}
	}

	agent := orchestrator.NewAgent(engine, r.idService, r.cfg, modelRequest.Provider)
	taskID := r.generateSecureToken(8)

	r.idService.RecordLLMUsage(req.Context(), user.Fingerprint, user.ShadowCookie)

	go func() {
		ctx := context.Background()
		report := func(msg string) {
			update := models.SSEUpdate{
				Status:   "processing",
				Progress: msg,
				NewLogs:  []string{msg},
			}
			b, _ := json.Marshal(update)
			r.idService.PublishTaskUpdate(ctx, taskID, string(b))
		}

		result := agent.Run(modelRequest.Topic, report)

		if strings.Contains(result.Message, "rate limit") {
			r.idService.SetSystemStatus(ctx, modelRequest.Provider, "rate_limited", 2*time.Minute)
		}

		if result.Status == "success" && result.Document != nil {
			// Generate requested assets
			var primaryFile string
			if modelRequest.Format == "pdf" {
				primaryFile = fmt.Sprintf("research_%s.pdf", taskID)
				if err := docxgen.GeneratePDF(primaryFile, result.Document); err != nil {
					log.Printf("PDF gen failed: %v", err)
				}
			} else {
				primaryFile = fmt.Sprintf("research_%s.docx", taskID)
				if err := docxgen.GenerateWordDoc(primaryFile, result.Document); err != nil {
					log.Printf("DOCX gen failed: %v", err)
				}
			}

			log.Printf("Target asset generated: %s", primaryFile)

			report(fmt.Sprintf("Finalizing research cloud-matrix (%s)...", strings.ToUpper(modelRequest.Format)))
			cloudURL, err := r.fileStore.UploadFile(ctx, primaryFile, taskID)
			if err != nil {
				log.Printf("Cloud upload failed: %v", err)
				update := models.SSEUpdate{Status: "error", Progress: "Cloud upload failed"}
				b, _ := json.Marshal(update)
				r.idService.PublishTaskUpdate(ctx, taskID, string(b))
				return
			}

			r.idService.SetTaskFile(ctx, taskID, cloudURL)
			update := models.SSEUpdate{
				Status:   "complete",
				Progress: fmt.Sprintf("Synthesis complete. %s finalized.", strings.ToUpper(modelRequest.Format)),
				Complete: true,
				Result:   cloudURL,
			}
			b, _ := json.Marshal(update)
			r.idService.PublishTaskUpdate(ctx, taskID, string(b))
		} else {
			update := models.SSEUpdate{
				Status:   "error",
				Progress: result.Message,
			}
			b, _ := json.Marshal(update)
			r.idService.PublishTaskUpdate(ctx, taskID, string(b))
		}
	}()

	r.sendJSON(w, http.StatusAccepted, map[string]string{
		"status":  "accepted",
		"task_id": taskID,
	})

	r.idService.RecordLLMUsage(req.Context(), user.Fingerprint, user.ShadowCookie)
}

// handleGetTaskStatus godoc
// @Summary Stream task progress
// @Description Opens an SSE stream to track the research pipeline.
// @Tags AI
// @Param id path string true "Task ID"
// @Success 200 {string} string "data: progress..."
// @Router /tasks/{id} [get]
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

// handleDownload godoc
// @Summary Download generated document
// @Description Serves the completed Word document.
// @Tags AI
// @Param id path string true "Task ID"
// @Produce application/vnd.openxmlformats-officedocument.wordprocessingml.document
// @Success 200 {file} file
// @Router /download/{id} [get]
func (r *Router) handleDownload(w http.ResponseWriter, req *http.Request) {
	taskID := req.PathValue("id")
	filePath, err := r.idService.GetTaskFile(req.Context(), taskID)
	if err != nil || filePath == "" {
		r.sendJSON(w, http.StatusNotFound, map[string]string{"error": "document not found"})
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filePath)))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	http.ServeFile(w, req, filePath)
}

func (r *Router) handleDownloadPDF(w http.ResponseWriter, req *http.Request) {
	taskID := req.PathValue("id")
	filePath, err := r.idService.GetTaskFile(req.Context(), taskID)
	if err != nil || filePath == "" {
		r.sendJSON(w, http.StatusNotFound, map[string]string{"error": "document not found"})
		return
	}
	// If filePath is a URL (Cloudinary), we should either redirect or proxy.
	// For simplicity, we'll try to serve local file first if it exists, otherwise error for now.
	// But actually, the frontend will likely use the result URL from SSE.
	if strings.HasPrefix(filePath, "http") {
		http.Redirect(w, req, filePath, http.StatusFound)
		return
	}
	
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="kyle_research_%s.pdf"`, taskID))
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, req, filePath)
}

func (r *Router) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
