package memory

import (
	"context"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

func TestJobsRepository_UpsertMany_DeduplicatesBySourceAndOriginalID(t *testing.T) {
	repository := NewJobsRepository()
	ctx := context.Background()

	inputs := []job.UpsertInput{
		{
			OriginalJobID: "glints-1",
			Title:         "Backend Engineer",
			Company:       "Acme",
			URL:           "https://example.com/jobs/1",
		},
		{
			OriginalJobID: "glints-1",
			Title:         "Backend Engineer Duplicate",
			Company:       "Acme",
			URL:           "https://example.com/jobs/1",
		},
	}

	result, err := repository.UpsertMany(ctx, job.SourceGlints, inputs)
	if err != nil {
		t.Fatalf("upsert many: %v", err)
	}

	if result.InsertedCount != 1 {
		t.Fatalf("expected inserted count 1, got %d", result.InsertedCount)
	}
	if result.DuplicateCount != 1 {
		t.Fatalf("expected duplicate count 1, got %d", result.DuplicateCount)
	}
}

func TestJobsRepository_Search_AppliesFilterSortAndPagination(t *testing.T) {
	repository := NewJobsRepository()
	ctx := context.Background()
	now := time.Now().UTC()
	earlier := now.Add(-2 * time.Hour)
	later := now.Add(2 * time.Hour)

	salaryA := int64(9_000_000)
	salaryB := int64(12_000_000)

	_, _ = repository.UpsertMany(ctx, job.SourceGlints, []job.UpsertInput{
		{
			OriginalJobID: "g-1",
			Title:         "Backend Engineer",
			Company:       "Acme",
			Location:      "Jakarta",
			Description:   "Golang service",
			URL:           "https://example.com/g-1",
			SalaryMin:     &salaryB,
			PostedAt:      &later,
		},
		{
			OriginalJobID: "g-2",
			Title:         "Frontend Engineer",
			Company:       "Acme",
			Location:      "Bandung",
			Description:   "React",
			URL:           "https://example.com/g-2",
			SalaryMin:     &salaryA,
			PostedAt:      &earlier,
		},
	})

	result, err := repository.Search(ctx, job.SearchQuery{
		Q:            "engineer",
		Source:       job.SourceGlints,
		Location:     "jakarta",
		HasSalaryMin: true,
		SalaryMin:    10_000_000,
		Sort:         "-posted_at",
		Page:         1,
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("search jobs: %v", err)
	}

	if result.TotalRecords != 1 {
		t.Fatalf("expected total records 1, got %d", result.TotalRecords)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 result row, got %d", len(result.Data))
	}
	if result.Data[0].OriginalJobID != "g-1" {
		t.Fatalf("expected original_job_id g-1, got %s", result.Data[0].OriginalJobID)
	}
}
