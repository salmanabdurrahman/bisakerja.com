package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/ai/openai"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/billing/mayar"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/postgres"
	aiapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/ai"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
	growthapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/growth"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/jobs"
	notificationapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/notification"
	platformauth "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/database"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/envloader"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := envloader.LoadAPIEnv(); err != nil {
		slog.Error("failed to load api environment", slog.String("error", err.Error()))
		os.Exit(1)
	}

	cfg := config.Load()
	appLogger := logger.New(cfg.Environment)

	tokenManager, err := platformauth.NewManager(cfg.AuthJWTSecret, cfg.AuthAccessTokenTTL, cfg.AuthRefreshTokenTTL)
	if err != nil {
		appLogger.Error("invalid auth configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	dbPool, err := database.OpenPostgres(ctx, cfg)
	if err != nil {
		appLogger.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	jobsRepository := postgres.NewJobsRepository(dbPool)
	jobsService := jobs.NewService(jobsRepository)
	jobsHandler := handler.NewJobsHandler(jobsService)

	identityRepository := postgres.NewIdentityRepository(dbPool)
	identityService := identityapp.NewService(identityRepository, tokenManager)
	authHandler := handler.NewAuthHandler(identityService)
	preferencesHandler := handler.NewPreferencesHandler(identityService)
	authMiddleware := middleware.NewAuthenticator(tokenManager)
	growthRepository := postgres.NewGrowthRepository(dbPool)
	growthService := growthapp.NewService(identityRepository, growthRepository)
	growthHandler := handler.NewGrowthHandler(growthService)

	notificationRepository := postgres.NewNotificationRepository(dbPool)
	notificationCenterService := notificationapp.NewCenterService(identityRepository, notificationRepository)
	notificationHandler := handler.NewNotificationHandler(notificationCenterService)

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
	billingHandler := handler.NewBillingHandler(billingService, cfg.BillingWebhookToken)

	aiRepository := postgres.NewAIRepository(dbPool)
	aiProvider := openai.NewClient(openai.ClientConfig{
		BaseURL: cfg.AIProviderBaseURL,
		APIKey:  cfg.AIProviderAPIKey,
		Model:   cfg.AIProviderModelDefault,
		Timeout: cfg.AIProviderTimeout,
	})
	aiService := aiapp.NewService(identityRepository, jobsRepository, aiRepository, aiProvider, aiapp.Config{
		DailyQuotaFree:    cfg.AIDailyQuotaFree,
		DailyQuotaPremium: cfg.AIDailyQuotaPremium,
	})
	aiHandler := handler.NewAIHandler(aiService)

	httpHandler := router.New(appLogger, router.Dependencies{
		JobsHandler:         jobsHandler,
		AuthHandler:         authHandler,
		PreferencesHandler:  preferencesHandler,
		BillingHandler:      billingHandler,
		AIHandler:           aiHandler,
		GrowthHandler:       growthHandler,
		NotificationHandler: notificationHandler,
		AuthMiddleware:      authMiddleware,
	})
	httpServer := server.NewHTTP(cfg, httpHandler, appLogger)
	if err = httpServer.Run(ctx); err != nil {
		appLogger.Error("api server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
