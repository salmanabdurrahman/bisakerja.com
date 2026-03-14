package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/billing/mayar"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/jobs"
	platformauth "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	appLogger := logger.New(cfg.Environment)

	tokenManager, err := platformauth.NewManager(cfg.AuthJWTSecret, cfg.AuthAccessTokenTTL, cfg.AuthRefreshTokenTTL)
	if err != nil {
		appLogger.Error("invalid auth configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	jobsRepository := memory.NewJobsRepository()
	jobsService := jobs.NewService(jobsRepository)
	jobsHandler := handler.NewJobsHandler(jobsService)

	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)
	authHandler := handler.NewAuthHandler(identityService)
	preferencesHandler := handler.NewPreferencesHandler(identityService)
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	billingRepository := memory.NewBillingRepository()
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
	billingHandler := handler.NewBillingHandler(billingService)

	httpHandler := router.New(appLogger, router.Dependencies{
		JobsHandler:        jobsHandler,
		AuthHandler:        authHandler,
		PreferencesHandler: preferencesHandler,
		BillingHandler:     billingHandler,
		AuthMiddleware:     authMiddleware,
	})
	httpServer := server.NewHTTP(cfg, httpHandler, appLogger)
	if err = httpServer.Run(ctx); err != nil {
		appLogger.Error("api server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
