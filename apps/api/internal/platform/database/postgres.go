package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
)

// OpenPostgres creates a PostgreSQL connection pool and verifies connectivity.
func OpenPostgres(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	databaseURL := strings.TrimSpace(cfg.DatabaseURL)
	if databaseURL == "" {
		return nil, fmt.Errorf("database url is required")
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DatabaseMaxOpenConns)
	poolConfig.MinConns = int32(cfg.DatabaseMinOpenConns)
	poolConfig.MaxConnLifetime = cfg.DatabaseMaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.DatabaseMaxConnIdleTime

	connectTimeout := cfg.DatabaseConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 5 * time.Second
	}

	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("open postgres pool: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, connectTimeout)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
