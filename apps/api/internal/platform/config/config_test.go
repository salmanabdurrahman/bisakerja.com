package config

import (
	"testing"
	"time"
)

func TestLoad_DefaultValues(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	t.Setenv("WORKER_TICK_INTERVAL", "")
	t.Setenv("SCRAPER_PAGE_SIZE", "")
	t.Setenv("SCRAPER_MAX_PAGES", "")
	t.Setenv("AUTH_JWT_SECRET", "")
	t.Setenv("AUTH_ACCESS_TOKEN_TTL", "")
	t.Setenv("AUTH_REFRESH_TOKEN_TTL", "")

	cfg := Load()

	if cfg.AppName != "bisakerja-api" {
		t.Fatalf("expected default app name, got %q", cfg.AppName)
	}

	if cfg.Environment != "development" {
		t.Fatalf("expected default environment, got %q", cfg.Environment)
	}

	if cfg.HTTPAddress() != ":8080" {
		t.Fatalf("expected default address :8080, got %q", cfg.HTTPAddress())
	}

	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("expected default shutdown timeout 10s, got %s", cfg.ShutdownTimeout)
	}

	if cfg.WorkerTick != 15*time.Second {
		t.Fatalf("expected default worker tick 15s, got %s", cfg.WorkerTick)
	}

	if cfg.ScraperPageSize != 30 {
		t.Fatalf("expected default scraper page size 30, got %d", cfg.ScraperPageSize)
	}

	if cfg.ScraperMaxPages != 1 {
		t.Fatalf("expected default scraper max pages 1, got %d", cfg.ScraperMaxPages)
	}

	if cfg.AuthJWTSecret != "bisakerja-dev-secret" {
		t.Fatalf("expected default auth jwt secret, got %q", cfg.AuthJWTSecret)
	}

	if cfg.AuthAccessTokenTTL != 15*time.Minute {
		t.Fatalf("expected default access ttl 15m, got %s", cfg.AuthAccessTokenTTL)
	}

	if cfg.AuthRefreshTokenTTL != 168*time.Hour {
		t.Fatalf("expected default refresh ttl 168h, got %s", cfg.AuthRefreshTokenTTL)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("APP_NAME", "custom-api")
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", ":9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "25s")
	t.Setenv("WORKER_TICK_INTERVAL", "9")
	t.Setenv("SCRAPER_PAGE_SIZE", "45")
	t.Setenv("SCRAPER_MAX_PAGES", "3")
	t.Setenv("AUTH_JWT_SECRET", "super-secret")
	t.Setenv("AUTH_ACCESS_TOKEN_TTL", "25m")
	t.Setenv("AUTH_REFRESH_TOKEN_TTL", "336h")

	cfg := Load()

	if cfg.AppName != "custom-api" {
		t.Fatalf("expected env app name, got %q", cfg.AppName)
	}

	if cfg.Environment != "production" {
		t.Fatalf("expected env environment, got %q", cfg.Environment)
	}

	if cfg.HTTPAddress() != ":9090" {
		t.Fatalf("expected env address :9090, got %q", cfg.HTTPAddress())
	}

	if cfg.ShutdownTimeout != 25*time.Second {
		t.Fatalf("expected shutdown timeout 25s, got %s", cfg.ShutdownTimeout)
	}

	if cfg.WorkerTick != 9*time.Second {
		t.Fatalf("expected worker tick 9s, got %s", cfg.WorkerTick)
	}

	if cfg.ScraperPageSize != 45 {
		t.Fatalf("expected scraper page size 45, got %d", cfg.ScraperPageSize)
	}

	if cfg.ScraperMaxPages != 3 {
		t.Fatalf("expected scraper max pages 3, got %d", cfg.ScraperMaxPages)
	}

	if cfg.AuthJWTSecret != "super-secret" {
		t.Fatalf("expected auth jwt secret super-secret, got %q", cfg.AuthJWTSecret)
	}

	if cfg.AuthAccessTokenTTL != 25*time.Minute {
		t.Fatalf("expected access ttl 25m, got %s", cfg.AuthAccessTokenTTL)
	}

	if cfg.AuthRefreshTokenTTL != 336*time.Hour {
		t.Fatalf("expected refresh ttl 336h, got %s", cfg.AuthRefreshTokenTTL)
	}
}
