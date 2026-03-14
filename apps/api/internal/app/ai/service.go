package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	aidomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/ai"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

const (
	defaultPromptMinChars    = 5
	defaultPromptMaxChars    = 500
	defaultFocusMaxChars     = 300
	defaultDailyQuotaFree    = 5
	defaultDailyQuotaPremium = 30

	tierFree            = "free"
	tierPremium         = "premium"
	jobFitVerdictStrong = "strong_match"
	jobFitVerdictMedium = "moderate_match"
	jobFitVerdictLow    = "low_match"
)

var (
	ErrPromptRequired      = errors.New("prompt is required")
	ErrPromptTooShort      = errors.New("prompt is too short")
	ErrPromptTooLong       = errors.New("prompt is too long")
	ErrFocusTooLong        = errors.New("focus is too long")
	ErrJobIDRequired       = errors.New("job_id is required")
	ErrPremiumRequired     = errors.New("premium subscription is required")
	ErrInvalidFeature      = errors.New("invalid ai feature")
	ErrQuotaExceeded       = errors.New("ai quota exceeded")
	ErrProviderRateLimited = errors.New("ai provider rate limited")
	ErrProviderUpstream    = errors.New("ai provider upstream error")
	ErrServiceUnavailable  = errors.New("ai service unavailable")
)

var allowedJobTypes = map[string]struct{}{
	"fulltime":   {},
	"parttime":   {},
	"contract":   {},
	"internship": {},
}

// Config stores configuration values for AI service.
type Config struct {
	DailyQuotaFree    int
	DailyQuotaPremium int
}

// Service coordinates application use cases for the package.
type Service struct {
	identityRepository identity.Repository
	jobsRepository     job.Repository
	repository         aidomain.Repository
	provider           aidomain.Provider
	dailyQuotaFree     int
	dailyQuotaPremium  int
	now                func() time.Time
}

// SearchAssistantContext contains optional user context for search assistant.
type SearchAssistantContext struct {
	Location  string
	JobTypes  []string
	SalaryMin *int64
}

// GenerateSearchAssistantInput contains input parameters for search assistant generation.
type GenerateSearchAssistantInput struct {
	UserID  string
	Prompt  string
	Context SearchAssistantContext
}

// UsageQuota represents usage quota detail.
type UsageQuota struct {
	DailyQuota int
	Used       int
	Remaining  int
	ResetAt    time.Time
}

// SearchAssistantResult represents AI search assistant output.
type SearchAssistantResult struct {
	Feature            aidomain.Feature
	Prompt             string
	SuggestedQuery     string
	SuggestedLocations []string
	SuggestedJobTypes  []string
	SuggestedSalaryMin *int64
	Summary            string
	Tier               string
	Provider           string
	Model              string
	Quota              UsageQuota
}

// GenerateJobFitSummaryInput contains input parameters for job fit summary generation.
type GenerateJobFitSummaryInput struct {
	UserID string
	JobID  string
	Focus  string
}

// JobFitSummaryResult represents AI job-fit summary output.
type JobFitSummaryResult struct {
	Feature     aidomain.Feature
	JobID       string
	FitScore    int
	Verdict     string
	Strengths   []string
	Gaps        []string
	NextActions []string
	Summary     string
	Tier        string
	Provider    string
	Model       string
	Quota       UsageQuota
}

// GetUsageInput contains input parameters for usage query.
type GetUsageInput struct {
	UserID  string
	Feature string
}

// UsageSnapshot represents user AI usage state.
type UsageSnapshot struct {
	Feature aidomain.Feature
	Tier    string
	Quota   UsageQuota
}

// NewService creates a new service instance.
func NewService(
	identityRepository identity.Repository,
	jobsRepository job.Repository,
	repository aidomain.Repository,
	provider aidomain.Provider,
	config Config,
) *Service {
	dailyQuotaFree := config.DailyQuotaFree
	if dailyQuotaFree <= 0 {
		dailyQuotaFree = defaultDailyQuotaFree
	}

	dailyQuotaPremium := config.DailyQuotaPremium
	if dailyQuotaPremium <= 0 {
		dailyQuotaPremium = defaultDailyQuotaPremium
	}

	return &Service{
		identityRepository: identityRepository,
		jobsRepository:     jobsRepository,
		repository:         repository,
		provider:           provider,
		dailyQuotaFree:     dailyQuotaFree,
		dailyQuotaPremium:  dailyQuotaPremium,
		now:                func() time.Time { return time.Now().UTC() },
	}
}

// GenerateSearchAssistant generates AI suggestions for search query.
func (s *Service) GenerateSearchAssistant(
	ctx context.Context,
	input GenerateSearchAssistantInput,
) (SearchAssistantResult, error) {
	if s.identityRepository == nil || s.repository == nil || s.provider == nil {
		return SearchAssistantResult{}, errors.New("ai service dependency is not fully configured")
	}

	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return SearchAssistantResult{}, identity.ErrUserNotFound
	}

	prompt, validationErr := normalizePrompt(input.Prompt)
	if validationErr != nil {
		return SearchAssistantResult{}, validationErr
	}

	now := s.now().UTC()
	user, err := s.identityRepository.GetUserByID(ctx, userID)
	if err != nil {
		return SearchAssistantResult{}, fmt.Errorf("get user profile: %w", err)
	}

	feature := aidomain.FeatureSearchAssistant
	tier := resolveTier(user, now)
	quotaLimit := s.quotaForTier(tier)
	windowStart := usageWindowStart(now)
	resetAt := windowStart.Add(24 * time.Hour)
	usedCount, err := s.repository.CountUsageSince(ctx, userID, feature, windowStart)
	if err != nil {
		return SearchAssistantResult{}, fmt.Errorf("count ai usage: %w", err)
	}
	if usedCount >= quotaLimit {
		return SearchAssistantResult{}, ErrQuotaExceeded
	}

	providerResult, err := s.provider.GenerateSearchAssistant(ctx, aidomain.SearchAssistantInput{
		Prompt: prompt,
		Context: aidomain.SearchAssistantContext{
			Location:  strings.TrimSpace(input.Context.Location),
			JobTypes:  normalizeStringList(input.Context.JobTypes, 4, 50),
			SalaryMin: cloneInt64(input.Context.SalaryMin),
		},
	})
	if err != nil {
		return SearchAssistantResult{}, mapProviderError(err)
	}

	normalizedResult := normalizeProviderResult(prompt, providerResult)
	createdAt := s.now().UTC()
	_, err = s.repository.CreateUsageLog(ctx, aidomain.CreateUsageLogInput{
		UserID:     userID,
		Feature:    feature,
		Tier:       tier,
		Provider:   normalizedResult.Provider,
		Model:      normalizedResult.Model,
		TokensIn:   max(normalizedResult.TokensIn, 0),
		TokensOut:  max(normalizedResult.TokensOut, 0),
		PromptHash: hashPrompt(prompt),
		Metadata: map[string]any{
			"prompt_length": len([]rune(prompt)),
		},
		CreatedAt: createdAt,
	})
	if err != nil {
		return SearchAssistantResult{}, fmt.Errorf("create ai usage log: %w", err)
	}

	usedAfter := usedCount + 1
	return SearchAssistantResult{
		Feature:            feature,
		Prompt:             prompt,
		SuggestedQuery:     normalizedResult.SuggestedQuery,
		SuggestedLocations: normalizedResult.SuggestedLocations,
		SuggestedJobTypes:  normalizedResult.SuggestedJobTypes,
		SuggestedSalaryMin: cloneInt64(normalizedResult.SuggestedSalaryMin),
		Summary:            normalizedResult.Summary,
		Tier:               tier,
		Provider:           normalizedResult.Provider,
		Model:              normalizedResult.Model,
		Quota: UsageQuota{
			DailyQuota: quotaLimit,
			Used:       usedAfter,
			Remaining:  max(quotaLimit-usedAfter, 0),
			ResetAt:    resetAt,
		},
	}, nil
}

// GenerateJobFitSummary generates AI job-fit summary for premium users.
func (s *Service) GenerateJobFitSummary(
	ctx context.Context,
	input GenerateJobFitSummaryInput,
) (JobFitSummaryResult, error) {
	if s.identityRepository == nil || s.jobsRepository == nil || s.repository == nil || s.provider == nil {
		return JobFitSummaryResult{}, errors.New("ai service dependency is not fully configured")
	}

	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return JobFitSummaryResult{}, identity.ErrUserNotFound
	}

	jobID := strings.TrimSpace(input.JobID)
	if jobID == "" {
		return JobFitSummaryResult{}, ErrJobIDRequired
	}

	focus, err := normalizeFocus(input.Focus)
	if err != nil {
		return JobFitSummaryResult{}, err
	}

	now := s.now().UTC()
	user, err := s.identityRepository.GetUserByID(ctx, userID)
	if err != nil {
		return JobFitSummaryResult{}, fmt.Errorf("get user profile: %w", err)
	}

	tier := resolveTier(user, now)
	if tier != tierPremium {
		return JobFitSummaryResult{}, ErrPremiumRequired
	}

	jobDetail, err := s.jobsRepository.GetByID(ctx, jobID)
	if err != nil {
		return JobFitSummaryResult{}, fmt.Errorf("get job detail: %w", err)
	}

	preferences, err := s.identityRepository.GetPreferences(ctx, userID)
	if err != nil {
		return JobFitSummaryResult{}, fmt.Errorf("get user preferences: %w", err)
	}

	feature := aidomain.FeatureJobFitSummary
	quotaLimit := s.quotaForTier(tier)
	windowStart := usageWindowStart(now)
	resetAt := windowStart.Add(24 * time.Hour)
	usedCount, err := s.repository.CountUsageSince(ctx, userID, feature, windowStart)
	if err != nil {
		return JobFitSummaryResult{}, fmt.Errorf("count ai usage: %w", err)
	}
	if usedCount >= quotaLimit {
		return JobFitSummaryResult{}, ErrQuotaExceeded
	}

	providerResult, err := s.provider.GenerateJobFitSummary(ctx, aidomain.JobFitSummaryInput{
		Focus: focus,
		Job: aidomain.JobFitJobContext{
			JobID:       jobDetail.ID,
			Title:       strings.TrimSpace(jobDetail.Title),
			Company:     strings.TrimSpace(jobDetail.Company),
			Location:    strings.TrimSpace(jobDetail.Location),
			Description: strings.TrimSpace(jobDetail.Description),
			SalaryMin:   cloneInt64(jobDetail.SalaryMin),
			SalaryMax:   cloneInt64(jobDetail.SalaryMax),
			SalaryRange: strings.TrimSpace(jobDetail.SalaryRange),
			PublishedAt: cloneTime(jobDetail.PostedAt),
		},
		Preferences: aidomain.JobFitUserPreferences{
			Keywords:  normalizeStringList(preferences.Keywords, 8, 60),
			Locations: normalizeStringList(preferences.Locations, 8, 60),
			JobTypes:  normalizeJobTypes(preferences.JobTypes),
			SalaryMin: max(preferences.SalaryMin, 0),
		},
	})
	if err != nil {
		return JobFitSummaryResult{}, mapProviderError(err)
	}

	normalizedResult := normalizeJobFitProviderResult(providerResult)
	createdAt := s.now().UTC()
	_, err = s.repository.CreateUsageLog(ctx, aidomain.CreateUsageLogInput{
		UserID:     userID,
		Feature:    feature,
		Tier:       tier,
		Provider:   normalizedResult.Provider,
		Model:      normalizedResult.Model,
		TokensIn:   max(normalizedResult.TokensIn, 0),
		TokensOut:  max(normalizedResult.TokensOut, 0),
		PromptHash: hashPrompt(fmt.Sprintf("job:%s|focus:%s", jobID, focus)),
		Metadata: map[string]any{
			"job_id":       jobID,
			"focus_length": len([]rune(focus)),
		},
		CreatedAt: createdAt,
	})
	if err != nil {
		return JobFitSummaryResult{}, fmt.Errorf("create ai usage log: %w", err)
	}

	usedAfter := usedCount + 1
	return JobFitSummaryResult{
		Feature:     feature,
		JobID:       jobID,
		FitScore:    normalizedResult.FitScore,
		Verdict:     normalizedResult.Verdict,
		Strengths:   normalizedResult.Strengths,
		Gaps:        normalizedResult.Gaps,
		NextActions: normalizedResult.NextActions,
		Summary:     normalizedResult.Summary,
		Tier:        tier,
		Provider:    normalizedResult.Provider,
		Model:       normalizedResult.Model,
		Quota: UsageQuota{
			DailyQuota: quotaLimit,
			Used:       usedAfter,
			Remaining:  max(quotaLimit-usedAfter, 0),
			ResetAt:    resetAt,
		},
	}, nil
}

// GetUsage returns AI usage quota for the user.
func (s *Service) GetUsage(ctx context.Context, input GetUsageInput) (UsageSnapshot, error) {
	if s.identityRepository == nil || s.repository == nil {
		return UsageSnapshot{}, errors.New("ai service dependency is not fully configured")
	}

	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return UsageSnapshot{}, identity.ErrUserNotFound
	}

	feature, err := parseFeature(input.Feature)
	if err != nil {
		return UsageSnapshot{}, err
	}

	now := s.now().UTC()
	user, err := s.identityRepository.GetUserByID(ctx, userID)
	if err != nil {
		return UsageSnapshot{}, fmt.Errorf("get user profile: %w", err)
	}

	tier := resolveTier(user, now)
	quotaLimit := s.quotaForTier(tier)
	windowStart := usageWindowStart(now)
	resetAt := windowStart.Add(24 * time.Hour)
	usedCount, err := s.repository.CountUsageSince(ctx, userID, feature, windowStart)
	if err != nil {
		return UsageSnapshot{}, fmt.Errorf("count ai usage: %w", err)
	}

	return UsageSnapshot{
		Feature: feature,
		Tier:    tier,
		Quota: UsageQuota{
			DailyQuota: quotaLimit,
			Used:       usedCount,
			Remaining:  max(quotaLimit-usedCount, 0),
			ResetAt:    resetAt,
		},
	}, nil
}

func normalizePrompt(raw string) (string, error) {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return "", ErrPromptRequired
	}

	length := len([]rune(normalized))
	if length < defaultPromptMinChars {
		return "", ErrPromptTooShort
	}
	if length > defaultPromptMaxChars {
		return "", ErrPromptTooLong
	}

	return normalized, nil
}

func parseFeature(raw string) (aidomain.Feature, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return aidomain.FeatureSearchAssistant, nil
	}
	if normalized == string(aidomain.FeatureSearchAssistant) {
		return aidomain.FeatureSearchAssistant, nil
	}
	if normalized == string(aidomain.FeatureJobFitSummary) {
		return aidomain.FeatureJobFitSummary, nil
	}
	return "", ErrInvalidFeature
}

func resolveTier(user identity.User, now time.Time) string {
	if !user.IsPremium {
		return tierFree
	}
	if user.PremiumExpiredAt == nil {
		return tierPremium
	}
	if user.PremiumExpiredAt.After(now) {
		return tierPremium
	}
	return tierFree
}

func (s *Service) quotaForTier(tier string) int {
	if tier == tierPremium {
		return s.dailyQuotaPremium
	}
	return s.dailyQuotaFree
}

func usageWindowStart(now time.Time) time.Time {
	utc := now.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func mapProviderError(err error) error {
	switch {
	case errors.Is(err, aidomain.ErrProviderRateLimited):
		return ErrProviderRateLimited
	case errors.Is(err, aidomain.ErrProviderUnavailable):
		return ErrServiceUnavailable
	case errors.Is(err, aidomain.ErrProviderUpstream):
		return ErrProviderUpstream
	default:
		return fmt.Errorf("generate ai response: %w", err)
	}
}

func normalizeProviderResult(prompt string, result aidomain.SearchAssistantResult) aidomain.SearchAssistantResult {
	normalizedQuery := strings.TrimSpace(result.SuggestedQuery)
	if normalizedQuery == "" {
		normalizedQuery = prompt
	}

	normalizedSummary := strings.TrimSpace(result.Summary)
	if normalizedSummary == "" {
		normalizedSummary = "Refined job search suggestion generated."
	}

	normalizedProvider := strings.TrimSpace(result.Provider)
	if normalizedProvider == "" {
		normalizedProvider = "openai_compatible"
	}

	normalizedModel := strings.TrimSpace(result.Model)
	if normalizedModel == "" {
		normalizedModel = "unknown"
	}

	return aidomain.SearchAssistantResult{
		SuggestedQuery:     normalizedQuery,
		SuggestedLocations: normalizeStringList(result.SuggestedLocations, 5, 80),
		SuggestedJobTypes:  normalizeJobTypes(result.SuggestedJobTypes),
		SuggestedSalaryMin: normalizeSalaryMin(result.SuggestedSalaryMin),
		Summary:            normalizedSummary,
		Provider:           normalizedProvider,
		Model:              normalizedModel,
		TokensIn:           max(result.TokensIn, 0),
		TokensOut:          max(result.TokensOut, 0),
	}
}

func normalizeStringList(values []string, maxItems int, maxLength int) []string {
	if maxItems <= 0 || maxLength <= 0 {
		return []string{}
	}

	result := make([]string, 0, maxItems)
	seen := make(map[string]struct{}, maxItems)
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" || len([]rune(normalized)) > maxLength {
			continue
		}

		key := strings.ToLower(normalized)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalized)
		if len(result) >= maxItems {
			break
		}
	}
	return result
}

func normalizeJobTypes(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if _, ok := allowedJobTypes[normalized]; !ok {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	slices.Sort(result)
	return result
}

func normalizeSalaryMin(value *int64) *int64 {
	if value == nil {
		return nil
	}
	if *value < 0 {
		return nil
	}
	cloned := *value
	return &cloned
}

func normalizeFocus(raw string) (string, error) {
	normalized := strings.TrimSpace(raw)
	if len([]rune(normalized)) > defaultFocusMaxChars {
		return "", ErrFocusTooLong
	}
	return normalized, nil
}

func normalizeJobFitProviderResult(result aidomain.JobFitSummaryResult) aidomain.JobFitSummaryResult {
	score := max(min(result.FitScore, 100), 0)
	verdict := strings.TrimSpace(strings.ToLower(result.Verdict))
	switch verdict {
	case jobFitVerdictStrong, jobFitVerdictMedium, jobFitVerdictLow:
	default:
		verdict = deriveFitVerdict(score)
	}

	summary := strings.TrimSpace(result.Summary)
	if summary == "" {
		summary = "Profile and job fit summary generated."
	}

	provider := strings.TrimSpace(result.Provider)
	if provider == "" {
		provider = "openai_compatible"
	}

	model := strings.TrimSpace(result.Model)
	if model == "" {
		model = "unknown"
	}

	return aidomain.JobFitSummaryResult{
		FitScore:    score,
		Verdict:     verdict,
		Strengths:   normalizeStringList(result.Strengths, 5, 160),
		Gaps:        normalizeStringList(result.Gaps, 5, 160),
		NextActions: normalizeStringList(result.NextActions, 5, 160),
		Summary:     summary,
		Provider:    provider,
		Model:       model,
		TokensIn:    max(result.TokensIn, 0),
		TokensOut:   max(result.TokensOut, 0),
	}
}

func deriveFitVerdict(score int) string {
	switch {
	case score >= 75:
		return jobFitVerdictStrong
	case score >= 50:
		return jobFitVerdictMedium
	default:
		return jobFitVerdictLow
	}
}

func hashPrompt(prompt string) string {
	sum := sha256.Sum256([]byte(prompt))
	return hex.EncodeToString(sum[:])
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}
