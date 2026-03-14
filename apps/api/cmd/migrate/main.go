package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/config"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/database"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/envloader"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/migration"
)

func main() {
	mode := flag.String("mode", "validate", "migration mode: validate, up, or down")
	direction := flag.String("direction", "up", "migration direction: up or down")
	path := flag.String("path", "./migrations", "migration directory path")
	flag.Parse()

	switch *mode {
	case "validate":
		runValidate(*path)
	case "up", "down":
		runMigrate(*mode, *direction, *path)
	default:
		fmt.Fprintf(os.Stderr, "invalid mode: %s\n", *mode)
		os.Exit(1)
	}
}

func runValidate(path string) {
	count, err := migration.ValidateDirectory(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migration validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("migration validation passed: pairs=%d path=%s\n", count, path)
}

func runMigrate(mode, direction, path string) {
	if direction != mode {
		fmt.Fprintf(os.Stderr, "mode %s must be used with -direction %s\n", mode, mode)
		os.Exit(1)
	}
	if err := envloader.LoadAPIEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load environment: %v\n", err)
		os.Exit(1)
	}

	cfg := config.Load()
	dbPool, err := database.OpenPostgres(context.Background(), cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect database: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	runner := migration.NewRunner(dbPool)
	switch mode {
	case "up":
		applied, err := runner.MigrateUp(context.Background(), path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "migration up failed: %v\n", err)
			os.Exit(1)
		}
		if len(applied) == 0 {
			fmt.Printf("migration up complete: no pending migrations path=%s\n", path)
			return
		}
		fmt.Printf("migration up complete: applied=%d versions=%v path=%s\n", len(applied), applied, path)
	case "down":
		version, reverted, err := runner.MigrateDown(context.Background(), path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "migration down failed: %v\n", err)
			os.Exit(1)
		}
		if !reverted {
			fmt.Printf("migration down complete: no applied migrations path=%s\n", path)
			return
		}
		fmt.Printf("migration down complete: reverted=%s path=%s\n", version, path)
	}
}
