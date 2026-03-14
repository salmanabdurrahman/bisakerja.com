package source

import (
	"context"
	"errors"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
)

func TestJobstreetAdapter_Fetch_RequiresToken(t *testing.T) {
	adapter := NewJobstreetAdapter(nil)

	_, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
		Keyword: "backend",
		Page:    1,
		Limit:   10,
		Token:   "",
	})
	if !errors.Is(err, scraper.ErrTokenUnavailable) {
		t.Fatalf("expected ErrTokenUnavailable, got %v", err)
	}
}
