package scraper

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

type fakeRepository struct {
	insertedBySource map[job.Source]int
	runs             []job.ScrapeRun
}

func (r *fakeRepository) UpsertMany(_ context.Context, source job.Source, jobs []job.UpsertInput) (job.UpsertResult, error) {
	if r.insertedBySource == nil {
		r.insertedBySource = map[job.Source]int{}
	}
	r.insertedBySource[source] += len(jobs)
	inserted := make([]job.Job, 0, len(jobs))
	for index := range jobs {
		inserted = append(inserted, job.Job{
			ID:            "job-test",
			Source:        source,
			OriginalJobID: jobs[index].OriginalJobID,
		})
	}
	return job.UpsertResult{Inserted: inserted, InsertedCount: len(jobs), DuplicateCount: 0}, nil
}

func (r *fakeRepository) Search(_ context.Context, _ job.SearchQuery) (job.SearchResult, error) {
	return job.SearchResult{}, nil
}

func (r *fakeRepository) GetByID(_ context.Context, _ string) (job.Job, error) {
	return job.Job{}, job.ErrNotFound
}

func (r *fakeRepository) RecordScrapeRun(_ context.Context, run job.ScrapeRun) error {
	r.runs = append(r.runs, run)
	return nil
}

func (r *fakeRepository) SearchTitles(_ context.Context, _ job.TitleSearchQuery) ([]string, error) {
	return []string{}, nil
}

type fakeAdapter struct {
	source       job.Source
	requiresAuth bool
	fetchErr     error
}

func (a fakeAdapter) Source() job.Source {
	return a.source
}

func (a fakeAdapter) RequiresAuth() bool {
	return a.requiresAuth
}

func (a fakeAdapter) Fetch(_ context.Context, req FetchRequest) (FetchResult, error) {
	if a.fetchErr != nil {
		return FetchResult{}, a.fetchErr
	}
	return FetchResult{
		Jobs: []job.UpsertInput{
			{
				OriginalJobID: string(a.source) + "-" + req.Keyword,
				Title:         "Backend Engineer",
				Company:       "Acme",
				URL:           "https://example.com/" + string(a.source),
			},
		},
		HasMore: false,
	}, nil
}

type fakeTokenProvider struct {
	token string
	err   error
}

func (p fakeTokenProvider) Resolve(_ context.Context, _ job.Source) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	return p.token, nil
}

func TestRunOnce_ContinuesWhenAuthSourceMissingToken(t *testing.T) {
	repository := &fakeRepository{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	orchestrator := NewOrchestrator(
		logger,
		repository,
		fakeTokenProvider{err: errors.New("missing token")},
		[]SourceAdapter{
			fakeAdapter{source: job.SourceGlints, requiresAuth: false},
			fakeAdapter{source: job.SourceJobstreet, requiresAuth: true},
		},
		Config{
			Keywords: []string{"backend"},
			PageSize: 10,
			MaxPages: 1,
		},
	)

	summary, err := orchestrator.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run once: %v", err)
	}

	if summary.SuccessSources != 1 {
		t.Fatalf("expected 1 successful source, got %d", summary.SuccessSources)
	}
	if summary.FailedSources != 1 {
		t.Fatalf("expected 1 failed source, got %d", summary.FailedSources)
	}
	if repository.insertedBySource[job.SourceGlints] != 1 {
		t.Fatalf("expected glints inserted count 1, got %d", repository.insertedBySource[job.SourceGlints])
	}
}

func TestRunOnce_RecordsPartialStatusOnFetchError(t *testing.T) {
	repository := &fakeRepository{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	orchestrator := NewOrchestrator(
		logger,
		repository,
		fakeTokenProvider{token: "token"},
		[]SourceAdapter{
			fakeAdapter{source: job.SourceKalibrr, requiresAuth: false, fetchErr: errors.New("timeout")},
		},
		Config{
			Keywords: []string{"backend"},
			PageSize: 10,
			MaxPages: 1,
		},
	)

	_, err := orchestrator.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run once: %v", err)
	}

	if len(repository.runs) != 1 {
		t.Fatalf("expected 1 run record, got %d", len(repository.runs))
	}
	if repository.runs[0].Status != job.ScrapeRunPartial {
		t.Fatalf("expected partial status, got %s", repository.runs[0].Status)
	}
	if repository.runs[0].ErrorClass != "source_fetch_error" {
		t.Fatalf("expected source_fetch_error class, got %s", repository.runs[0].ErrorClass)
	}
	if repository.runs[0].FinishedAt.Before(repository.runs[0].StartedAt) {
		t.Fatalf("expected finished_at after started_at, got start=%s finish=%s", repository.runs[0].StartedAt, repository.runs[0].FinishedAt)
	}
}

func TestRunOnce_UsesDefaultConfigWhenEmpty(t *testing.T) {
	repository := &fakeRepository{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	orchestrator := NewOrchestrator(
		logger,
		repository,
		fakeTokenProvider{token: "token"},
		[]SourceAdapter{fakeAdapter{source: job.SourceGlints}},
		Config{},
	)

	summary, err := orchestrator.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run once: %v", err)
	}

	if summary.InsertedCount == 0 {
		t.Fatal("expected inserted count > 0 with default config")
	}
	if summary.ProcessedAt.After(time.Now().UTC().Add(2 * time.Second)) {
		t.Fatalf("unexpected processed_at value: %s", summary.ProcessedAt)
	}
}

func TestRunOnce_CallsInsertedHook(t *testing.T) {
	repository := &fakeRepository{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	orchestrator := NewOrchestrator(
		logger,
		repository,
		fakeTokenProvider{token: "token"},
		[]SourceAdapter{fakeAdapter{source: job.SourceGlints}},
		Config{Keywords: []string{"backend"}, PageSize: 10, MaxPages: 1},
	)

	hookCalled := 0
	orchestrator.SetOnJobInserted(func(context.Context, job.Job) error {
		hookCalled++
		return nil
	})

	_, err := orchestrator.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if hookCalled == 0 {
		t.Fatal("expected onJobInserted hook to be called")
	}
}

func TestRunOnce_LogsAuthMissingSourceDetails(t *testing.T) {
	repository := &fakeRepository{}
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	orchestrator := NewOrchestrator(
		logger,
		repository,
		fakeTokenProvider{err: errors.New("missing token")},
		[]SourceAdapter{
			fakeAdapter{source: job.SourceJobstreet, requiresAuth: true},
		},
		Config{
			Keywords: []string{"backend"},
			PageSize: 10,
			MaxPages: 1,
		},
	)

	_, err := orchestrator.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run once: %v", err)
	}

	logOutput := logBuffer.String()
	for _, fragment := range []string{
		"msg=\"scrape source processed\"",
		"source=jobstreet",
		"status=failed_auth",
		"error_class=auth_missing",
		"source_operation=resolve_token",
		"error_message=\"missing token\"",
	} {
		if !strings.Contains(logOutput, fragment) {
			t.Fatalf("expected log output to contain %q, got %s", fragment, logOutput)
		}
	}
}

func TestRunOnce_LogsFetchErrorMetadata(t *testing.T) {
	repository := &fakeRepository{}
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	orchestrator := NewOrchestrator(
		logger,
		repository,
		fakeTokenProvider{token: "token"},
		[]SourceAdapter{
			fakeAdapter{
				source:       job.SourceKalibrr,
				requiresAuth: false,
				fetchErr:     WrapSourceError("execute_request", 429, errors.New("unexpected upstream response")),
			},
		},
		Config{
			Keywords: []string{"backend"},
			PageSize: 10,
			MaxPages: 1,
		},
	)

	_, err := orchestrator.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("run once: %v", err)
	}

	logOutput := logBuffer.String()
	for _, fragment := range []string{
		"msg=\"scrape source processed\"",
		"source=kalibrr",
		"status=partial",
		"error_class=source_fetch_error",
		"keyword=backend",
		"page=1",
		"source_operation=execute_request",
		"http_status_last=429",
		"error_message=\"execute_request (status=429): unexpected upstream response\"",
	} {
		if !strings.Contains(logOutput, fragment) {
			t.Fatalf("expected log output to contain %q, got %s", fragment, logOutput)
		}
	}
}
