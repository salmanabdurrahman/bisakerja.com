package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

func Healthcheck(name string) {
	fmt.Printf("worker=%s status=ok\n", name)
}

func Run(ctx context.Context, logger *slog.Logger, name string, tickInterval time.Duration) error {
	return RunWithTask(ctx, logger, name, tickInterval, nil)
}

func RunWithTask(
	ctx context.Context,
	logger *slog.Logger,
	name string,
	tickInterval time.Duration,
	task func(context.Context) error,
) error {
	logger.Info("worker started", "worker", name, "tick_interval", tickInterval.String())
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker shutting down", "worker", name)
			return nil
		case <-ticker.C:
			if task != nil {
				if err := task(ctx); err != nil {
					logger.Error("worker task failed", "worker", name, "error", err.Error())
				}
			}
			logger.Debug("worker heartbeat", "worker", name)
		}
	}
}
