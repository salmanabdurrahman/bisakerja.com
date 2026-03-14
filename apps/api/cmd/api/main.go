package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/jobs"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	appLogger := logger.New(cfg.Environment)

	jobsRepository := memory.NewJobsRepository()
	jobsService := jobs.NewService(jobsRepository)
	jobsHandler := handler.NewJobsHandler(jobsService)

	httpHandler := router.New(appLogger, router.Dependencies{JobsHandler: jobsHandler})
	httpServer := server.NewHTTP(cfg, httpHandler, appLogger)
	if err := httpServer.Run(ctx); err != nil {
		appLogger.Error("api server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
