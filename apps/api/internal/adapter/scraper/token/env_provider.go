package token

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

// EnvProvider represents env provider.
type EnvProvider struct{}

// NewEnvProvider creates a new env provider instance.
func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

// Resolve resolves.
func (p *EnvProvider) Resolve(_ context.Context, source job.Source) (string, error) {
	switch source {
	case job.SourceJobstreet:
		token := strings.TrimSpace(os.Getenv("JOBSTREET_BEARER_TOKEN"))
		if token == "" {
			return "", fmt.Errorf("%w: JOBSTREET_BEARER_TOKEN is empty", scraper.ErrTokenUnavailable)
		}
		return token, nil
	default:
		return "", fmt.Errorf("%w: source %s does not have a token mapping", scraper.ErrTokenUnavailable, source)
	}
}
