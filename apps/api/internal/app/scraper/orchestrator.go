package scraper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

var (
	ErrTokenUnavailable   = errors.New("token unavailable")
	ErrSourceUnauthorized = errors.New("source unauthorized")
)

// FetchRequest represents fetch request.
type FetchRequest struct {
	Keyword string
	Page    int
	Limit   int
	Token   string
}

// FetchResult contains result values for fetch.
type FetchResult struct {
	Jobs    []job.UpsertInput
	HasMore bool
}

// SourceAdapter defines behavior for source adapter.
type SourceAdapter interface {
	Source() job.Source
	RequiresAuth() bool
	Fetch(ctx context.Context, request FetchRequest) (FetchResult, error)
}

// TokenProvider defines behavior for token provider.
type TokenProvider interface {
	Resolve(ctx context.Context, source job.Source) (string, error)
}

// Config stores configuration values for config.
type Config struct {
	Keywords []string
	PageSize int
	MaxPages int
}

// RunSummary summarizes execution details for run.
type RunSummary struct {
	Sources        int
	SuccessSources int
	PartialSources int
	FailedSources  int
	InsertedCount  int
	DuplicateCount int
	ProcessedAt    time.Time
}

// Orchestrator represents orchestrator.
type Orchestrator struct {
	logger        *slog.Logger
	repository    job.Repository
	tokenProvider TokenProvider
	adapters      []SourceAdapter
	config        Config
	onJobInserted func(context.Context, job.Job) error
}

// NewOrchestrator creates a new orchestrator instance.
func NewOrchestrator(
	logger *slog.Logger,
	repository job.Repository,
	tokenProvider TokenProvider,
	adapters []SourceAdapter,
	config Config,
) *Orchestrator {
	if logger == nil {
		logger = slog.Default()
	}
	if config.PageSize <= 0 {
		config.PageSize = 30
	}
	if config.MaxPages <= 0 {
		config.MaxPages = 1
	}
	if len(config.Keywords) == 0 {
		config.Keywords = []string{"backend", "frontend", "intern"}
	}

	return &Orchestrator{
		logger:        logger,
		repository:    repository,
		tokenProvider: tokenProvider,
		adapters:      adapters,
		config:        config,
	}
}

// RunOnce runs once.
func (o *Orchestrator) RunOnce(ctx context.Context) (RunSummary, error) {
	if o.repository == nil {
		return RunSummary{}, errors.New("repository is required")
	}

	summary := RunSummary{
		Sources:     len(o.adapters),
		ProcessedAt: time.Now().UTC(),
	}

	for _, adapter := range o.adapters {
		runRecord := job.ScrapeRun{
			Source:    adapter.Source(),
			Status:    job.ScrapeRunSuccess,
			StartedAt: time.Now().UTC(),
		}

		token, err := o.resolveToken(ctx, adapter)
		if err != nil {
			runRecord.Status = job.ScrapeRunFailedAuth
			runRecord.ErrorClass = "auth_missing"
			runRecord.ErrorMessage = err.Error()
			runRecord.FinishedAt = time.Now().UTC()
			_ = o.repository.RecordScrapeRun(ctx, runRecord)
			summary.FailedSources++
			continue
		}

		var sourceError error
		for _, keyword := range o.config.Keywords {
			for page := 1; page <= o.config.MaxPages; page++ {
				result, fetchErr := adapter.Fetch(ctx, FetchRequest{
					Keyword: keyword,
					Page:    page,
					Limit:   o.config.PageSize,
					Token:   token,
				})
				if fetchErr != nil {
					sourceError = fetchErr
					if errors.Is(fetchErr, ErrSourceUnauthorized) {
						runRecord.Status = job.ScrapeRunFailedAuth
						runRecord.ErrorClass = "auth_failed"
					} else {
						runRecord.Status = job.ScrapeRunPartial
						runRecord.ErrorClass = "source_fetch_error"
					}
					runRecord.ErrorMessage = fetchErr.Error()
					break
				}

				runRecord.FetchedCount += len(result.Jobs)
				upsertResult, upsertErr := o.repository.UpsertMany(ctx, adapter.Source(), result.Jobs)
				if upsertErr != nil {
					sourceError = upsertErr
					runRecord.Status = job.ScrapeRunFailed
					runRecord.ErrorClass = "repository_error"
					runRecord.ErrorMessage = upsertErr.Error()
					break
				}

				runRecord.InsertedCount += upsertResult.InsertedCount
				runRecord.DuplicateCount += upsertResult.DuplicateCount
				summary.InsertedCount += upsertResult.InsertedCount
				summary.DuplicateCount += upsertResult.DuplicateCount
				if o.onJobInserted != nil {
					for _, inserted := range upsertResult.Inserted {
						if publishErr := o.onJobInserted(ctx, inserted); publishErr != nil {
							o.logger.Error(
								"job inserted hook failed",
								"source", adapter.Source(),
								"job_id", inserted.ID,
								"error", publishErr.Error(),
							)
						}
					}
				}

				if !result.HasMore {
					break
				}
			}

			if sourceError != nil {
				break
			}
		}

		runRecord.FinishedAt = time.Now().UTC()
		if recordErr := o.repository.RecordScrapeRun(ctx, runRecord); recordErr != nil {
			return summary, fmt.Errorf("record scrape run for source %s: %w", adapter.Source(), recordErr)
		}

		switch runRecord.Status {
		case job.ScrapeRunSuccess:
			summary.SuccessSources++
		case job.ScrapeRunPartial:
			summary.PartialSources++
		default:
			summary.FailedSources++
		}

		o.logger.Info(
			"scrape source processed",
			"source", adapter.Source(),
			"status", runRecord.Status,
			"fetched_count", runRecord.FetchedCount,
			"inserted_count", runRecord.InsertedCount,
			"duplicate_count", runRecord.DuplicateCount,
			"error_class", runRecord.ErrorClass,
		)
	}

	return summary, nil
}

func (o *Orchestrator) resolveToken(ctx context.Context, adapter SourceAdapter) (string, error) {
	if !adapter.RequiresAuth() {
		return "", nil
	}
	if o.tokenProvider == nil {
		return "", fmt.Errorf("%w for source %s: token provider is nil", ErrTokenUnavailable, adapter.Source())
	}

	token, err := o.tokenProvider.Resolve(ctx, adapter.Source())
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", fmt.Errorf("%w for source %s: empty token", ErrTokenUnavailable, adapter.Source())
	}

	return token, nil
}

// SetOnJobInserted sets on job inserted.
func (o *Orchestrator) SetOnJobInserted(callback func(context.Context, job.Job) error) {
	o.onJobInserted = callback
}
