package token

import (
	"context"
	"errors"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

func TestEnvProvider_ResolveJobstreetToken(t *testing.T) {
	t.Setenv("JOBSTREET_BEARER_TOKEN", "token-123")
	provider := NewEnvProvider()

	token, err := provider.Resolve(context.Background(), job.SourceJobstreet)
	if err != nil {
		t.Fatalf("resolve token: %v", err)
	}
	if token != "token-123" {
		t.Fatalf("expected token token-123, got %s", token)
	}
}

func TestEnvProvider_ResolveTokenMissing(t *testing.T) {
	t.Setenv("JOBSTREET_BEARER_TOKEN", "")
	provider := NewEnvProvider()

	_, err := provider.Resolve(context.Background(), job.SourceJobstreet)
	if !errors.Is(err, scraper.ErrTokenUnavailable) {
		t.Fatalf("expected ErrTokenUnavailable, got %v", err)
	}
}
