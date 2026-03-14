package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
)

func TestHTTPRun_GracefulShutdownOnCancel(t *testing.T) {
	cfg := config.Config{
		HTTPPort:        "0",
		ShutdownTimeout: 2 * time.Second,
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	httpServer := NewHTTP(cfg, handler, logger)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	if err := httpServer.Run(ctx); err != nil {
		t.Fatalf("expected graceful shutdown without error, got %v", err)
	}
}
