package logger

import (
	"log/slog"
	"os"
)

// New creates a new structured logger
func New() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, opts)

	// Use text handler for development
	if os.Getenv("ENV") == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
