package job

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Source represents source.
type Source string

const (
	SourceGlints    Source = "glints"
	SourceKalibrr   Source = "kalibrr"
	SourceJobstreet Source = "jobstreet"
)

var (
	ErrNotFound = errors.New("job not found")
)

// Job represents job.
type Job struct {
	ID            string
	Source        Source
	OriginalJobID string
	Title         string
	Company       string
	Location      string
	Description   string
	URL           string
	SalaryMin     *int64
	SalaryMax     *int64
	SalaryRange   string
	PostedAt      *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	RawData       map[string]any
}

// UpsertInput contains input parameters for upsert.
type UpsertInput struct {
	OriginalJobID string
	Title         string
	Company       string
	Location      string
	Description   string
	URL           string
	SalaryMin     *int64
	SalaryMax     *int64
	SalaryRange   string
	PostedAt      *time.Time
	RawData       map[string]any
}

// UpsertResult contains result values for upsert.
type UpsertResult struct {
	Inserted       []Job
	InsertedCount  int
	DuplicateCount int
}

// SearchQuery represents search query.
type SearchQuery struct {
	Q            string
	Location     string
	Source       Source
	SalaryMin    int64
	HasSalaryMin bool
	Sort         string
	Page         int
	Limit        int
}

// SearchResult contains result values for search.
type SearchResult struct {
	Data         []Job
	Page         int
	Limit        int
	TotalPages   int
	TotalRecords int
}

// ScrapeRunStatus describes status details for scrape run.
type ScrapeRunStatus string

const (
	ScrapeRunSuccess    ScrapeRunStatus = "success"
	ScrapeRunPartial    ScrapeRunStatus = "partial"
	ScrapeRunFailed     ScrapeRunStatus = "failed"
	ScrapeRunFailedAuth ScrapeRunStatus = "failed_auth"
)

// ScrapeRun represents scrape run.
type ScrapeRun struct {
	ID             string
	Source         Source
	Status         ScrapeRunStatus
	ErrorClass     string
	ErrorMessage   string
	FetchedCount   int
	InsertedCount  int
	DuplicateCount int
	StartedAt      time.Time
	FinishedAt     time.Time
}

// Repository defines behavior for repository.
type Repository interface {
	UpsertMany(ctx context.Context, source Source, jobs []UpsertInput) (UpsertResult, error)
	Search(ctx context.Context, query SearchQuery) (SearchResult, error)
	GetByID(ctx context.Context, id string) (Job, error)
	RecordScrapeRun(ctx context.Context, run ScrapeRun) error
}

// ParseSource parses source.
func ParseSource(raw string) (Source, bool) {
	value := Source(strings.TrimSpace(strings.ToLower(raw)))
	switch value {
	case SourceGlints, SourceKalibrr, SourceJobstreet:
		return value, true
	default:
		return "", false
	}
}

// IsSupportedSource handles is supported source.
func IsSupportedSource(raw string) bool {
	_, ok := ParseSource(raw)
	return ok
}

// SupportedSources handles supported sources.
func SupportedSources() []string {
	return []string{string(SourceGlints), string(SourceKalibrr), string(SourceJobstreet)}
}
