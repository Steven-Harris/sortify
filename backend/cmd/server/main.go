package main

import (
	"log/slog"
	"os"

	"github.com/Steven-harris/sortify/backend/internal/api"
	"github.com/Steven-harris/sortify/backend/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create server instance
	server := api.NewServer(cfg)

	// Initialize server (create directories, etc.)
	if err := server.Initialize(); err != nil {
		slog.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}

	// Start server with graceful shutdown
	if err := server.Start(); err != nil {
		slog.Error("Server stopped with error", "error", err)
		os.Exit(1)
	}
}
