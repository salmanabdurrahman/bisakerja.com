package ai

import (
	"context"
	"errors"
	"time"
)

// Feature represents AI feature capability.
type Feature string

const (
	FeatureSearchAssistant  Feature = "search_assistant"
	FeatureJobFitSummary    Feature = "job_fit_summary"
	FeatureCoverLetterDraft Feature = "cover_letter_draft"
)

var (
	ErrProviderRateLimited = errors.New("provider rate limited")
	ErrProviderUpstream    = errors.New("provider upstream error")
	ErrProviderUnavailable = errors.New("provider unavailable")
)

// SearchAssistantContext contains optional user context for AI generation.
type SearchAssistantContext struct {
	Location  string
	JobTypes  []string
	SalaryMin *int64
}

// SearchAssistantInput contains input parameters for search assistant generation.
type SearchAssistantInput struct {
	Prompt  string
	Context SearchAssistantContext
}

// SearchAssistantResult represents AI search assistant output.
type SearchAssistantResult struct {
	SuggestedQuery     string
	SuggestedLocations []string
	SuggestedJobTypes  []string
	SuggestedSalaryMin *int64
	Summary            string
	Provider           string
	Model              string
	TokensIn           int
	TokensOut          int
}

// JobFitJobContext contains job details used for fit summary generation.
type JobFitJobContext struct {
	JobID       string
	Title       string
	Company     string
	Location    string
	Description string
	SalaryMin   *int64
	SalaryMax   *int64
	SalaryRange string
	PublishedAt *time.Time
}

// JobFitUserPreferences contains user preference details used for fit summary generation.
type JobFitUserPreferences struct {
	Keywords  []string
	Locations []string
	JobTypes  []string
	SalaryMin int64
}

// JobFitSummaryInput contains input parameters for job fit summary generation.
type JobFitSummaryInput struct {
	Focus       string
	Job         JobFitJobContext
	Preferences JobFitUserPreferences
}

// JobFitSummaryResult represents AI job fit summary output.
type JobFitSummaryResult struct {
	FitScore    int
	Verdict     string
	Strengths   []string
	Gaps        []string
	NextActions []string
	Summary     string
	Provider    string
	Model       string
	TokensIn    int
	TokensOut   int
}

// CoverLetterDraftInput contains input parameters for cover letter draft generation.
type CoverLetterDraftInput struct {
	Tone        string
	Highlights  []string
	Job         JobFitJobContext
	Preferences JobFitUserPreferences
	UserName    string
}

// CoverLetterDraftResult represents AI cover letter draft output.
type CoverLetterDraftResult struct {
	Tone      string
	Draft     string
	KeyPoints []string
	Summary   string
	Provider  string
	Model     string
	TokensIn  int
	TokensOut int
}

// UsageLog represents persisted AI usage event.
type UsageLog struct {
	ID         string
	UserID     string
	Feature    Feature
	Tier       string
	Provider   string
	Model      string
	TokensIn   int
	TokensOut  int
	PromptHash string
	Metadata   map[string]any
	CreatedAt  time.Time
}

// CreateUsageLogInput contains input parameters for usage logging.
type CreateUsageLogInput struct {
	UserID     string
	Feature    Feature
	Tier       string
	Provider   string
	Model      string
	TokensIn   int
	TokensOut  int
	PromptHash string
	Metadata   map[string]any
	CreatedAt  time.Time
}

// Repository defines behavior for AI persistence.
type Repository interface {
	CountUsageSince(ctx context.Context, userID string, feature Feature, since time.Time) (int, error)
	CreateUsageLog(ctx context.Context, input CreateUsageLogInput) (UsageLog, error)
}

// Provider defines behavior for AI provider adapter.
type Provider interface {
	GenerateSearchAssistant(ctx context.Context, input SearchAssistantInput) (SearchAssistantResult, error)
	GenerateJobFitSummary(ctx context.Context, input JobFitSummaryInput) (JobFitSummaryResult, error)
	GenerateCoverLetterDraft(ctx context.Context, input CoverLetterDraftInput) (CoverLetterDraftResult, error)
}
