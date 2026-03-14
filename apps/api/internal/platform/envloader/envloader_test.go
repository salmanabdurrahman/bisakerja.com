package envloader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_SetsMissingEnvOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("DATABASE_URL=postgres://localhost/test\nAPP_ENV=development\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	t.Setenv("APP_ENV", "production")

	if err := Load(path); err != nil {
		t.Fatalf("load env: %v", err)
	}

	if value := os.Getenv("DATABASE_URL"); value != "postgres://localhost/test" {
		t.Fatalf("unexpected database url: %q", value)
	}
	if value := os.Getenv("APP_ENV"); value != "production" {
		t.Fatalf("expected existing env to win, got %q", value)
	}
}

func TestParseLine_SupportsExportAndQuotedValue(t *testing.T) {
	key, value, ok := parseLine(`export AUTH_JWT_SECRET="super-secret"`)
	if !ok {
		t.Fatal("expected parse success")
	}
	if key != "AUTH_JWT_SECRET" {
		t.Fatalf("unexpected key: %q", key)
	}
	if value != "super-secret" {
		t.Fatalf("unexpected value: %q", value)
	}
}

func TestParseLine_RejectsInvalidLine(t *testing.T) {
	if _, _, ok := parseLine("INVALID LINE"); ok {
		t.Fatal("expected invalid line to be rejected")
	}
}
