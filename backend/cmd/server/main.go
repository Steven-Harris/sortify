package main

import (
	"log/slog"
	"os"

	"github.com/Steven-harris/sortify/backend/internal/api"
	"github.com/Steven-harris/sortify/backend/internal/config"
)

func main() {
	cfg := config.Load()

	server := api.NewServer(cfg)

	if err := server.Initialize(); err != nil {
		slog.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}

	if err := server.Start(); err != nil {
		slog.Error("Server stopped with error", "error", err)
		os.Exit(1)
	}
}
