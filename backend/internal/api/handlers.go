package api

import (
	"net/http"
	"time"

	"github.com/Steven-harris/sortify/backend/pkg/response"
)

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	healthData := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"service":   "sortify-api",
	}

	response.Success(w, healthData)
}

func (s *Server) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	response.NotFound(w, "Endpoint not found")
}

func (s *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	apiInfo := map[string]any{
		"name":        "Sortify API",
		"description": "Photo and video management API",
		"version":     "1.0.0",
		"endpoints": map[string]string{
			"health": "/api/health",
			"upload": "/api/upload/*",
			"media":  "/api/media/*",
		},
	}

	response.Success(w, apiInfo)
}
