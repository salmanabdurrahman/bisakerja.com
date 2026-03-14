package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/migration"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	path := flag.String("path", "./migrations", "migration directory path")
	flag.Parse()

	if *direction != "up" && *direction != "down" {
		fmt.Fprintf(os.Stderr, "invalid direction: %s\n", *direction)
		os.Exit(1)
	}

	count, err := migration.ValidateDirectory(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migration validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("migration validation passed: direction=%s pairs=%d path=%s\n", *direction, count, *path)
}
