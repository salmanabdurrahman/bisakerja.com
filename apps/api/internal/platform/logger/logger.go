package logger

import (
	"log/slog"
	"os"
)

// New creates a new instance.
func New(environment string) *slog.Logger {
	level := slog.LevelDebug
	if environment == "production" {
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
