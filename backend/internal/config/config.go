package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	MediaPath   string
	LogLevel    string
	CORSOrigins string
}

func Load() *Config {
	config := &Config{
		Port:        getEnv("PORT", "8080"),
		MediaPath:   getEnv("MEDIA_PATH", "./media"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		CORSOrigins: getEnv("CORS_ORIGINS", "*"),
	}

	var logLevel slog.Level
	switch config.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Configuration loaded",
		"port", config.Port,
		"media_path", config.MediaPath,
		"log_level", config.LogLevel,
	)

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
