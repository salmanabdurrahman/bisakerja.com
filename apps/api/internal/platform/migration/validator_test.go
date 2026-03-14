package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDirectory_OK(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "000001_init.up.sql")
	mustWriteFile(t, dir, "000001_init.down.sql")

	count, err := ValidateDirectory(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 1 {
		t.Fatalf("expected migration count = 1, got %d", count)
	}
}

func TestValidateDirectory_IncompletePair(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "000001_init.up.sql")

	if _, err := ValidateDirectory(dir); err == nil {
		t.Fatal("expected incomplete pair error, got nil")
	}
}

func mustWriteFile(t *testing.T, dir, name string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("-- migration"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
