package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/worker"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "run worker healthcheck and exit")
	flag.Parse()

	if *healthcheck {
		worker.Healthcheck("scraper")
		return
	}

	cfg := config.Load()
	appLogger := logger.New(cfg.Environment)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := worker.Run(ctx, appLogger, "scraper", cfg.WorkerTick); err != nil {
		appLogger.Error("scraper worker failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
