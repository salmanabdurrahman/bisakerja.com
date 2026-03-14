package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/billing/mayar"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/postgres"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
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
		worker.Healthcheck("billing-worker")
		return
	}

	if err := envloader.LoadAPIEnv(); err != nil {
		slog.Error("failed to load api environment", slog.String("error", err.Error()))
		os.Exit(1)
	}

	cfg := config.Load()
	appLogger := logger.New(cfg.Environment)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := database.OpenPostgres(ctx, cfg)
	if err != nil {
		appLogger.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	identityRepository := postgres.NewIdentityRepository(dbPool)
	billingRepository := postgres.NewBillingRepository(dbPool)
	mayarClient := mayar.NewClient(mayar.ClientConfig{
		BaseURL:    cfg.MayarBaseURL,
		APIKey:     cfg.MayarAPIKey,
		Timeout:    cfg.MayarRequestTimeout,
		MaxRetries: cfg.MayarMaxRetries,
	})
	billingService := billingapp.NewService(identityRepository, billingRepository, mayarClient, billingapp.Config{
		RedirectAllowlist: cfg.BillingRedirectAllowlist,
		IdempotencyWindow: cfg.BillingIdempotencyWindow,
		RateLimitWindow:   cfg.BillingUserRateLimitWindow,
	})

	if err = worker.RunWithTask(ctx, appLogger, "billing-worker", cfg.WorkerTick, func(taskCtx context.Context) error {
		summary, reconcileErr := billingService.ReconcileWithMayar(taskCtx)
		if reconcileErr != nil {
			return reconcileErr
		}

		if summary.AnomalyCount > 0 {
			appLogger.Warn(
				"billing anomaly detected",
				"anomaly_count", summary.AnomalyCount,
				"pending_or_reminder_scanned", summary.ScannedTransactions,
			)
		}

		appLogger.Info(
			"billing reconciliation tick finished",
			"scanned_transactions", summary.ScannedTransactions,
			"reconciled", summary.ReconciledCount,
			"retryable_failures", summary.RetryableFailures,
			"anomaly_count", summary.AnomalyCount,
		)

		return nil
	}); err != nil {
		appLogger.Error("billing worker failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
