package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Healthcheck reports worker health and exits with status code.
func Healthcheck(name string) {
	fmt.Printf("worker=%s status=ok\n", name)
}

// Run runs the main execution flow.
func Run(ctx context.Context, logger *slog.Logger, name string, tickInterval time.Duration) error {
	return RunWithTask(ctx, logger, name, tickInterval, nil)
}

// RunWithTask runs with task.
func RunWithTask(
	ctx context.Context,
	logger *slog.Logger,
	name string,
	tickInterval time.Duration,
	task func(context.Context) error,
) error {
	logger.Info("worker started", "operation", "worker_lifecycle", "worker", name, "tick_interval", tickInterval.String())
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker shutting down", "operation", "worker_lifecycle", "worker", name)
			return nil
		case <-ticker.C:
			if task != nil {
				if err := task(ctx); err != nil {
					logger.Error(
						"worker task failed",
						"operation", "run_worker_task",
						"error_class", "worker_task_error",
						"worker", name,
						"error", err.Error(),
					)
				}
			}
			logger.Debug("worker heartbeat", "operation", "worker_heartbeat", "worker", name)
		}
	}
}
