package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	aidomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/ai"
)

// AIRepository represents in-memory AI repository.
type AIRepository struct {
	mu        sync.RWMutex
	usageLogs []aidomain.UsageLog
}

// NewAIRepository creates a new in-memory AI repository instance.
func NewAIRepository() *AIRepository {
	return &AIRepository{
		usageLogs: make([]aidomain.UsageLog, 0),
	}
}

// CountUsageSince returns usage count since provided time for user+feature.
func (r *AIRepository) CountUsageSince(
	_ context.Context,
	userID string,
	feature aidomain.Feature,
	since time.Time,
) (int, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedFeature := strings.TrimSpace(string(feature))
	sinceUTC := since.UTC()

	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, item := range r.usageLogs {
		if item.UserID != normalizedUserID {
			continue
		}
		if string(item.Feature) != normalizedFeature {
			continue
		}
		if item.CreatedAt.Before(sinceUTC) {
			continue
		}
		count++
	}
	return count, nil
}

// CreateUsageLog creates AI usage log.
func (r *AIRepository) CreateUsageLog(
	_ context.Context,
	input aidomain.CreateUsageLogInput,
) (aidomain.UsageLog, error) {
	createdAt := input.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	record := aidomain.UsageLog{
		ID:         "ai_" + randomHex(12),
		UserID:     strings.TrimSpace(input.UserID),
		Feature:    aidomain.Feature(strings.TrimSpace(string(input.Feature))),
		Tier:       strings.TrimSpace(strings.ToLower(input.Tier)),
		Provider:   strings.TrimSpace(input.Provider),
		Model:      strings.TrimSpace(input.Model),
		TokensIn:   max(input.TokensIn, 0),
		TokensOut:  max(input.TokensOut, 0),
		PromptHash: strings.TrimSpace(strings.ToLower(input.PromptHash)),
		Metadata:   cloneMetadata(input.Metadata),
		CreatedAt:  createdAt,
	}

	r.usageLogs = append(r.usageLogs, record)
	return cloneUsageLog(record), nil
}

func cloneUsageLog(value aidomain.UsageLog) aidomain.UsageLog {
	return aidomain.UsageLog{
		ID:         value.ID,
		UserID:     value.UserID,
		Feature:    value.Feature,
		Tier:       value.Tier,
		Provider:   value.Provider,
		Model:      value.Model,
		TokensIn:   value.TokensIn,
		TokensOut:  value.TokensOut,
		PromptHash: value.PromptHash,
		Metadata:   cloneMetadata(value.Metadata),
		CreatedAt:  value.CreatedAt,
	}
}

func cloneMetadata(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}

var _ aidomain.Repository = (*AIRepository)(nil)
