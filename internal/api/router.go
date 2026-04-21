package api

import (
	"encoding/json"
	"net/http"

	"github.com/firebreather-heart/kyle/internal/identity"
	"github.com/firebreather-heart/kyle/internal/api/middleware"
)

type Router struct{
	idService identity.Service
}

func NewRouter(idService identity.Service) *Router {
	return &Router{idService: idService}
}

func (r *Router) SetupRoutes(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("/api/v1/triangulate", r.handleTriangulate)
	mux.HandleFunc("/api/v1/generate", r.handleGenerate)
	mux.HandleFunc("/api/v1/tasks/", r.handleGetTaskStatus)
	mux.HandleFunc("/api/v1/download/", r.handleDownload)

	var handler http.Handler = middleware.Logger(mux)
	handler = middleware.CORS(handler)
	handler = middleware.Authenticator(r.idService)(handler)
	return handler
}

func (r *Router) handleTriangulate(w http.ResponseWriter, req *http.Request) {
	// Logic for user resolution
}

func (r *Router) handleGenerate(w http.ResponseWriter, req *http.Request) {
	// Logic for starting AI generation
}

func (r *Router) handleGetTaskStatus(w http.ResponseWriter, req *http.Request) {
	// Logic for polling task status
}

func (r *Router) handleDownload(w http.ResponseWriter, req *http.Request) {
	// Logic for serving the .docx file
}

func (r *Router) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}