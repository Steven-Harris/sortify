package api

import (
	"net/http"
	"time"
)

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	healthData := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"service":   "sortify-api",
	}

	Success(w, healthData)
}

func (s *Server) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	NotFound(w, "Endpoint not found")
}
