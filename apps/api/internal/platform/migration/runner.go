package migration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Runner struct {
	pool *pgxpool.Pool
}

func NewRunner(pool *pgxpool.Pool) *Runner {
	return &Runner{pool: pool}
}

func (r *Runner) MigrateUp(ctx context.Context, path string) ([]string, error) {
	files, err := CollectDirectory(path)
	if err != nil {
		return nil, err
	}
	if err := r.ensureSchemaMigrationsTable(ctx); err != nil {
		return nil, err
	}

	appliedVersions, err := r.loadAppliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	applied := make([]string, 0)
	for _, file := range files {
		if _, exists := appliedVersions[file.Key]; exists {
			continue
		}
		if err := r.applyFile(ctx, file.Key, file.UpPath, true); err != nil {
			return nil, err
		}
		applied = append(applied, file.Key)
	}

	return applied, nil
}

func (r *Runner) MigrateDown(ctx context.Context, path string) (string, bool, error) {
	files, err := CollectDirectory(path)
	if err != nil {
		return "", false, err
	}
	if err := r.ensureSchemaMigrationsTable(ctx); err != nil {
		return "", false, err
	}

	latestVersion, err := r.latestAppliedVersion(ctx)
	if err != nil {
		return "", false, err
	}
	if latestVersion == "" {
		return "", false, nil
	}

	for index := len(files) - 1; index >= 0; index-- {
		file := files[index]
		if file.Key != latestVersion {
			continue
		}

		if err := r.applyFile(ctx, file.Key, file.DownPath, false); err != nil {
			return "", false, err
		}
		return file.Key, true, nil
	}

	return "", false, fmt.Errorf("latest applied migration %s not found in %s", latestVersion, path)
}

func (r *Runner) ensureSchemaMigrationsTable(ctx context.Context) error {
	if r.pool == nil {
		return errors.New("migration runner requires database pool")
	}

	query := `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
)
`

	if _, err := r.pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	return nil
}

func (r *Runner) loadAppliedVersions(ctx context.Context) (map[string]struct{}, error) {
	rows, err := r.pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	defer rows.Close()

	result := make(map[string]struct{})
	for rows.Next() {
		var version string
		if scanErr := rows.Scan(&version); scanErr != nil {
			return nil, fmt.Errorf("scan applied migration: %w", scanErr)
		}
		result[strings.TrimSpace(version)] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return result, nil
}

func (r *Runner) latestAppliedVersion(ctx context.Context) (string, error) {
	var version string
	err := r.pool.QueryRow(
		ctx,
		`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`,
	).Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("load latest applied migration: %w", err)
	}
	return strings.TrimSpace(version), nil
}

func (r *Runner) applyFile(ctx context.Context, version, path string, isUp bool) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", path, err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration transaction for %s: %w", version, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("execute migration %s: %w", version, err)
	}

	if isUp {
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			return fmt.Errorf("record applied migration %s: %w", version, err)
		}
	} else {
		if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, version); err != nil {
			return fmt.Errorf("remove applied migration %s: %w", version, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}

	return nil
}
