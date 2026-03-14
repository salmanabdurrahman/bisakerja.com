package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	notificationemail "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/notifier/email"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	queuememory "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/queue/memory"
	notificationapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/notification"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/worker"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "run worker healthcheck and exit")
	flag.Parse()

	if *healthcheck {
		worker.Healthcheck("notifier")
		return
	}

	cfg := config.Load()
	appLogger := logger.New(cfg.Environment)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	jobsRepository := memory.NewJobsRepository()
	identityRepository := memory.NewIdentityRepository()
	notificationRepository := memory.NewNotificationRepository()
	queue := queuememory.NewQueue()

	matcher := notificationapp.NewMatcher(
		appLogger,
		jobsRepository,
		identityRepository,
		notificationRepository,
		queue,
		100,
	)
	emailSender := notificationemail.NewLoggerSender(appLogger)
	notifier := notificationapp.NewNotifier(appLogger, notificationRepository, queue, emailSender, 100)

	if err := worker.RunWithTask(ctx, appLogger, "notifier", cfg.WorkerTick, func(taskCtx context.Context) error {
		matchSummary, matchErr := matcher.RunOnce(taskCtx)
		if matchErr != nil {
			return matchErr
		}

		notifySummary, notifyErr := notifier.RunOnce(taskCtx)
		if notifyErr != nil {
			return notifyErr
		}

		appLogger.Info(
			"notifier tick finished",
			"job_events_processed", matchSummary.ProcessedEvents,
			"matched_users", matchSummary.MatchedUsers,
			"deliveries_enqueued", matchSummary.EnqueuedDeliveries,
			"duplicate_notifications", matchSummary.DuplicateCount,
			"delivery_tasks_processed", notifySummary.ProcessedTasks,
			"notifications_sent", notifySummary.SentCount,
			"notifications_failed", notifySummary.FailedCount,
		)
		return nil
	}); err != nil {
		appLogger.Error("notifier worker failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
