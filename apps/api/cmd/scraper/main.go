package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/scraper/source"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/scraper/token"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
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

	repository := memory.NewJobsRepository()
	orchestrator := scraper.NewOrchestrator(
		appLogger,
		repository,
		token.NewEnvProvider(),
		[]scraper.SourceAdapter{
			source.NewGlintsAdapter(nil),
			source.NewKalibrrAdapter(nil),
			source.NewJobstreetAdapter(nil),
		},
		scraper.Config{
			Keywords: parseKeywords(os.Getenv("SCRAPER_KEYWORDS")),
			PageSize: cfg.ScraperPageSize,
			MaxPages: cfg.ScraperMaxPages,
		},
	)

	if err := worker.RunWithTask(ctx, appLogger, "scraper", cfg.WorkerTick, func(taskCtx context.Context) error {
		summary, runErr := orchestrator.RunOnce(taskCtx)
		if runErr != nil {
			return runErr
		}

		appLogger.Info(
			"scrape tick finished",
			"sources", summary.Sources,
			"success_sources", summary.SuccessSources,
			"partial_sources", summary.PartialSources,
			"failed_sources", summary.FailedSources,
			"inserted_count", summary.InsertedCount,
			"duplicate_count", summary.DuplicateCount,
		)
		return nil
	}); err != nil {
		appLogger.Error("scraper worker failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func parseKeywords(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	chunks := strings.Split(raw, ",")
	result := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		keyword := strings.TrimSpace(chunk)
		if keyword == "" {
			continue
		}
		result = append(result, keyword)
	}

	return result
}
