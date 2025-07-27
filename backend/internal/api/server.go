package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Steven-harris/sortify/backend/internal/config"
)

// Server represents the HTTP server
type Server struct {
	config        *config.Config
	server        *http.Server
	uploadHandler *UploadHandlers
	mediaHandler  *MediaHandlers
}

// NewServer creates a new HTTP server instance
func NewServer(cfg *config.Config) *Server {
	// Create temporary directory for uploads
	tempDir := filepath.Join(cfg.MediaPath, "temp")
	
	return &Server{
		config:        cfg,
		uploadHandler: NewUploadHandlers(tempDir, cfg.MediaPath),
		mediaHandler:  NewMediaHandlers(cfg.MediaPath),
	}
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	// Create router
	router := s.setupRoutes()

	// Configure HTTP server
	s.server = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		slog.Info("Starting server", "port", s.config.Port, "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-quit
	slog.Info("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		return err
	}

	slog.Info("Server stopped gracefully")
	return nil
}

// setupRoutes configures all the HTTP routes
func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Apply middleware
	var handler http.Handler = mux
	handler = CORS(s.config.CORSOrigins)(handler)
	handler = Logging(handler)
	handler = Recovery(handler)

	// Register routes
	mux.HandleFunc("/", s.RootHandler)
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
	mux.HandleFunc("/api/media/metadata", s.mediaHandler.MetadataHandler)
	mux.HandleFunc("/api/media/user-date", s.mediaHandler.UserDateHandler)
	
	// Catch-all for undefined routes
	mux.HandleFunc("/api/", s.NotFoundHandler)

	return handler
}

// ensureDirectories creates necessary directories if they don't exist
func (s *Server) ensureDirectories() error {
	directories := []string{
		s.config.MediaPath,
		filepath.Join(s.config.MediaPath, "temp"), // Temporary upload directory
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		slog.Info("Directory ensured", "path", dir)
	}

	return nil
}

// Initialize performs server initialization tasks
func (s *Server) Initialize() error {
	slog.Info("Initializing server...")
	
	// Ensure required directories exist
	if err := s.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	slog.Info("Server initialization completed")
	return nil
}
