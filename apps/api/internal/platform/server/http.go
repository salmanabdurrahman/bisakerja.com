package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
)

type HTTP struct {
	server          *http.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

func NewHTTP(cfg config.Config, handler http.Handler, logger *slog.Logger) *HTTP {
	return &HTTP{
		server: &http.Server{
			Addr:              cfg.HTTPAddress(),
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
		logger:          logger,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

func (s *HTTP) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("http server started", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		s.logger.Info("http server shutting down", "timeout", s.shutdownTimeout.String())
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}

		return nil
	case err := <-errCh:
		return err
	}
}
