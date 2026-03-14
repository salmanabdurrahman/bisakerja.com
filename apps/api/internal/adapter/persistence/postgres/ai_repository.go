package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	aidomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/ai"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

// AIRepository represents PostgreSQL AI repository.
type AIRepository struct {
	pool *pgxpool.Pool
}

// NewAIRepository creates a new PostgreSQL AI repository instance.
func NewAIRepository(pool *pgxpool.Pool) *AIRepository {
	return &AIRepository{pool: pool}
}

// CountUsageSince returns usage count since provided time for user+feature.
func (r *AIRepository) CountUsageSince(
	ctx context.Context,
	userID string,
	feature aidomain.Feature,
	since time.Time,
) (int, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedFeature := strings.TrimSpace(string(feature))
	if normalizedUserID == "" || normalizedFeature == "" {
		return 0, nil
	}

	query := `
SELECT COUNT(1)
FROM ai_usage_logs
WHERE user_id::text = $1
  AND feature = $2
  AND created_at >= $3
`

	var count int
	if err := r.pool.QueryRow(ctx, query, normalizedUserID, normalizedFeature, since.UTC()).Scan(&count); err != nil {
		return 0, fmt.Errorf("count ai usage: %w", err)
	}
	return count, nil
}

// CreateUsageLog creates AI usage log.
func (r *AIRepository) CreateUsageLog(
	ctx context.Context,
	input aidomain.CreateUsageLogInput,
) (aidomain.UsageLog, error) {
	metadata, err := encodeJSON(input.Metadata)
	if err != nil {
		return aidomain.UsageLog{}, fmt.Errorf("encode usage metadata: %w", err)
	}

	createdAt := input.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	query := `
WITH selected_user AS (
  SELECT id FROM users WHERE id::text = $1
)
INSERT INTO ai_usage_logs (
  user_id,
  feature,
  tier,
  provider,
  model,
  tokens_in,
  tokens_out,
  prompt_hash,
  metadata,
  created_at
)
SELECT
  selected_user.id,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9::jsonb,
  $10
FROM selected_user
RETURNING
  id::text,
  user_id::text,
  feature,
  tier,
  provider,
  model,
  tokens_in,
  tokens_out,
  prompt_hash,
  metadata,
  created_at
`

	var (
		item        aidomain.UsageLog
		metadataRaw []byte
	)
	err = r.pool.QueryRow(
		ctx,
		query,
		strings.TrimSpace(input.UserID),
		strings.TrimSpace(string(input.Feature)),
		strings.TrimSpace(strings.ToLower(input.Tier)),
		strings.TrimSpace(input.Provider),
		strings.TrimSpace(input.Model),
		max(input.TokensIn, 0),
		max(input.TokensOut, 0),
		strings.TrimSpace(strings.ToLower(input.PromptHash)),
		metadata,
		createdAt,
	).Scan(
		&item.ID,
		&item.UserID,
		&item.Feature,
		&item.Tier,
		&item.Provider,
		&item.Model,
		&item.TokensIn,
		&item.TokensOut,
		&item.PromptHash,
		&metadataRaw,
		&item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return aidomain.UsageLog{}, identity.ErrUserNotFound
		}
		return aidomain.UsageLog{}, fmt.Errorf("create ai usage log: %w", err)
	}

	decodedMetadata, decodeErr := decodeJSON(metadataRaw)
	if decodeErr != nil {
		return aidomain.UsageLog{}, fmt.Errorf("decode usage metadata: %w", decodeErr)
	}

	item.Metadata = decodedMetadata
	item.CreatedAt = item.CreatedAt.UTC()
	return item, nil
}

var _ aidomain.Repository = (*AIRepository)(nil)
