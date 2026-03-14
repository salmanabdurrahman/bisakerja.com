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
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("APP_NAME", "custom-api")
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", ":9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "25s")
	t.Setenv("WORKER_TICK_INTERVAL", "9")

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
}
