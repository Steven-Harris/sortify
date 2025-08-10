package api

import (
	"net/http"
	"os"
	"path/filepath"
)

func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Apply middleware
	var handler http.Handler = mux
	handler = CORS(s.config.CORSOrigins)(handler)
	handler = Logging(handler)
	handler = Recovery(handler)

	// API routes
	mux.HandleFunc("/api/health", s.HealthHandler)

	// Upload routes
	mux.HandleFunc("/api/upload/start", s.uploadHandler.StartUploadHandler)
	mux.HandleFunc("/api/upload/chunk", s.uploadHandler.UploadChunkHandler)
	mux.HandleFunc("/api/upload/complete", s.uploadHandler.CompleteUploadHandler)
	mux.HandleFunc("/api/upload/progress", s.uploadHandler.GetProgressHandler)
	mux.HandleFunc("/api/upload/pause", s.uploadHandler.PauseUploadHandler)
	mux.HandleFunc("/api/upload/resume", s.uploadHandler.ResumeUploadHandler)
	mux.HandleFunc("/api/upload/cancel", s.uploadHandler.CancelUploadHandler)

	// Media browsing routes
	mux.HandleFunc("/api/media/browse", s.mediaHandler.BrowseHandler)
	mux.HandleFunc("/api/media/files", s.mediaHandler.ListFilesHandler)
	mux.HandleFunc("/api/media/metadata", s.mediaHandler.MetadataHandler)
	mux.HandleFunc("/api/media/user-date", s.mediaHandler.UserDateHandler)

	// Static file serving for media files
	mediaFileServer := http.FileServer(http.Dir(s.config.MediaPath))
	mux.Handle("/media/", http.StripPrefix("/media/", mediaFileServer))

	// Catch-all for undefined API routes
	mux.HandleFunc("/api/", s.NotFoundHandler)

	// Serve frontend static files
	mux.HandleFunc("/", s.serveFrontend)

	return handler
}

func (s *Server) serveFrontend(w http.ResponseWriter, r *http.Request) {
	// Try embedded files first (available when built with embedded frontend)
	distPath := filepath.Join("./web")
	if _, err := os.Stat(distPath); err == nil {
		if r.URL.Path != "/" {
			filePath := filepath.Join(distPath, r.URL.Path)
			if _, err := os.Stat(filePath); err == nil {
				http.ServeFile(w, r, filePath)
				return
			}
		}

		// Serve index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
		return
	}

	// Fallback to local frontend/dist directory for development
	distPath = filepath.Join("../frontend/dist")
	if _, err := os.Stat(distPath); err == nil {
		if r.URL.Path != "/" {
			filePath := filepath.Join(distPath, r.URL.Path)
			if _, err := os.Stat(filePath); err == nil {
				http.ServeFile(w, r, filePath)
				return
			}
		}

		// Serve index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
		return
	}

	// No frontend available, serve a simple message
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head><title>Sortify API</title></head>
		<body>
			<h1>Sortify API Server</h1>
			<p>The frontend is not yet built. Build the frontend or access the API at <a href="/api/health">/api/health</a></p>
			<p>Available endpoints:</p>
			<ul>
				<li><a href="/api/health">Health Check</a></li>
				<li>/api/upload/* - Upload endpoints</li>
				<li>/api/media/* - Media browsing endpoints</li>
			</ul>
		</body>
		</html>
	`))
	if err != nil {
		// Error writing response - not much we can do at this point
		// since headers are already sent, but we should log it
		return
	}
}
