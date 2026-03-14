package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/postgres"
	queuepostgres "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/queue/postgres"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/scraper/source"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/scraper/token"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/database"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/envloader"
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

	if err := envloader.LoadAPIEnv(); err != nil {
		slog.Error("failed to load api environment", slog.String("error", err.Error()))
		os.Exit(1)
	}

	cfg := config.Load()
	appLogger := logger.New(cfg.Environment).With("service", "scraper-worker")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := database.OpenPostgres(ctx, cfg)
	if err != nil {
		appLogger.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	repository := postgres.NewJobsRepository(dbPool)
	queue := queuepostgres.NewQueue(dbPool)
	glintsAdapter := source.NewGlintsAdapter(nil)
	glintsAdapter.Cookie = strings.TrimSpace(os.Getenv("GLINTS_COOKIE"))
	jobstreetAdapter := source.NewJobstreetAdapter(nil)
	jobstreetAdapter.Cookie = strings.TrimSpace(os.Getenv("JOBSTREET_COOKIE"))
	jobstreetAdapter.SeekSessionID = strings.TrimSpace(os.Getenv("JOBSTREET_EC_SESSION_ID"))
	jobstreetAdapter.SeekVisitorID = strings.TrimSpace(os.Getenv("JOBSTREET_EC_VISITOR_ID"))
	orchestrator := scraper.NewOrchestrator(
		appLogger,
		repository,
		token.NewEnvProvider(),
		[]scraper.SourceAdapter{
			glintsAdapter,
			source.NewKalibrrAdapter(nil),
			jobstreetAdapter,
		},
		scraper.Config{
			Keywords: parseKeywords(os.Getenv("SCRAPER_KEYWORDS")),
			PageSize: cfg.ScraperPageSize,
			MaxPages: cfg.ScraperMaxPages,
		},
	)
	orchestrator.SetOnJobInserted(func(callbackCtx context.Context, insertedJob job.Job) error {
		return queue.EnqueueJobEvent(callbackCtx, notification.JobEvent{JobID: insertedJob.ID})
	})

	if err = worker.RunWithTask(ctx, appLogger, "scraper", cfg.WorkerTick, func(taskCtx context.Context) error {
		tickStartedAt := time.Now().UTC()
		summary, runErr := orchestrator.RunOnce(taskCtx)
		if runErr != nil {
			return runErr
		}

		appLogger.Info(
			"scrape tick finished",
			"operation", "run_tick",
			"sources", summary.Sources,
			"success_sources", summary.SuccessSources,
			"partial_sources", summary.PartialSources,
			"failed_sources", summary.FailedSources,
			"inserted_count", summary.InsertedCount,
			"duplicate_count", summary.DuplicateCount,
			"processed_at", summary.ProcessedAt.Format(time.RFC3339),
			"duration_ms", time.Since(tickStartedAt).Milliseconds(),
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
