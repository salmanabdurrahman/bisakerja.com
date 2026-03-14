package job

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Source string

const (
	SourceGlints    Source = "glints"
	SourceKalibrr   Source = "kalibrr"
	SourceJobstreet Source = "jobstreet"
)

var (
	ErrNotFound = errors.New("job not found")
)

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

type UpsertResult struct {
	Inserted       []Job
	InsertedCount  int
	DuplicateCount int
}

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

type SearchResult struct {
	Data         []Job
	Page         int
	Limit        int
	TotalPages   int
	TotalRecords int
}

type ScrapeRunStatus string

const (
	ScrapeRunSuccess    ScrapeRunStatus = "success"
	ScrapeRunPartial    ScrapeRunStatus = "partial"
	ScrapeRunFailed     ScrapeRunStatus = "failed"
	ScrapeRunFailedAuth ScrapeRunStatus = "failed_auth"
)

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

type Repository interface {
	UpsertMany(ctx context.Context, source Source, jobs []UpsertInput) (UpsertResult, error)
	Search(ctx context.Context, query SearchQuery) (SearchResult, error)
	GetByID(ctx context.Context, id string) (Job, error)
	RecordScrapeRun(ctx context.Context, run ScrapeRun) error
}

func ParseSource(raw string) (Source, bool) {
	value := Source(strings.TrimSpace(strings.ToLower(raw)))
	switch value {
	case SourceGlints, SourceKalibrr, SourceJobstreet:
		return value, true
	default:
		return "", false
	}
}

func IsSupportedSource(raw string) bool {
	_, ok := ParseSource(raw)
	return ok
}

func SupportedSources() []string {
	return []string{string(SourceGlints), string(SourceKalibrr), string(SourceJobstreet)}
}
