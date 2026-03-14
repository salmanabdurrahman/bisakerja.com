package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	appLogger := logger.New(cfg.Environment)

	handler := router.New(appLogger)
	httpServer := server.NewHTTP(cfg, handler, appLogger)
	if err := httpServer.Run(ctx); err != nil {
		appLogger.Error("api server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
