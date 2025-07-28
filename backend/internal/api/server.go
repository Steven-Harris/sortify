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

type Server struct {
	config        *config.Config
	server        *http.Server
	uploadHandler *UploadHandlers
	mediaHandler  *MediaHandlers
}

func NewServer(cfg *config.Config) *Server {
	// Create temporary directory for uploads
	tempDir := filepath.Join(cfg.MediaPath, "temp")

	return &Server{
		config:        cfg,
		uploadHandler: NewUploadHandlers(tempDir, cfg.MediaPath),
		mediaHandler:  NewMediaHandlers(cfg.MediaPath),
	}
}

func (s *Server) Start() error {
	router := s.setupRoutes()

	s.server = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Starting server", "port", s.config.Port, "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		return err
	}

	slog.Info("Server stopped gracefully")
	return nil
}

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

func (s *Server) Initialize() error {
	slog.Info("Initializing server...")

	if err := s.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	slog.Info("Server initialization completed")
	return nil
}
