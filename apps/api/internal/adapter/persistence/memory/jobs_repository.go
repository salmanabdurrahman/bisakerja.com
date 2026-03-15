package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

// JobsRepository represents jobs repository.
type JobsRepository struct {
	mu            sync.RWMutex
	byID          map[string]job.Job
	bySourceKey   map[string]string
	scrapeRunLogs []job.ScrapeRun
}

// NewJobsRepository creates a new jobs repository instance.
func NewJobsRepository() *JobsRepository {
	return &JobsRepository{
		byID:        make(map[string]job.Job),
		bySourceKey: make(map[string]string),
	}
}

// UpsertMany upserts many.
func (r *JobsRepository) UpsertMany(_ context.Context, source job.Source, inputs []job.UpsertInput) (job.UpsertResult, error) {
	if source == "" {
		return job.UpsertResult{}, errors.New("source is required")
	}

	now := time.Now().UTC()
	inserted := make([]job.Job, 0, len(inputs))
	duplicateCount := 0

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, input := range inputs {
		normalizedInput, err := normalizeInput(input)
		if err != nil {
			return job.UpsertResult{}, err
		}

		sourceKey := sourceJobKey(source, normalizedInput.OriginalJobID)
		if _, exists := r.bySourceKey[sourceKey]; exists {
			duplicateCount++
			continue
		}

		jobID := "job_" + randomHex(12)
		record := job.Job{
			ID:            jobID,
			Source:        source,
			OriginalJobID: normalizedInput.OriginalJobID,
			Title:         normalizedInput.Title,
			Company:       normalizedInput.Company,
			Location:      normalizedInput.Location,
			Description:   normalizedInput.Description,
			URL:           normalizedInput.URL,
			SalaryMin:     normalizedInput.SalaryMin,
			SalaryMax:     normalizedInput.SalaryMax,
			SalaryRange:   normalizedInput.SalaryRange,
			PostedAt:      normalizedInput.PostedAt,
			CreatedAt:     now,
			UpdatedAt:     now,
			RawData:       normalizedInput.RawData,
		}

		r.byID[jobID] = record
		r.bySourceKey[sourceKey] = jobID
		inserted = append(inserted, record)
	}

	return job.UpsertResult{
		Inserted:       inserted,
		InsertedCount:  len(inserted),
		DuplicateCount: duplicateCount,
	}, nil
}

// Search searches.
func (r *JobsRepository) Search(_ context.Context, query job.SearchQuery) (job.SearchResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Sort == "" {
		query.Sort = "-posted_at"
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	filtered := make([]job.Job, 0, len(r.byID))
	for _, candidate := range r.byID {
		if !matchesQuery(candidate, query) {
			continue
		}
		filtered = append(filtered, candidate)
	}

	sortJobs(filtered, query.Sort)

	totalRecords := len(filtered)
	totalPages := 0
	if totalRecords > 0 {
		totalPages = (totalRecords + query.Limit - 1) / query.Limit
	}

	start := (query.Page - 1) * query.Limit
	if start >= totalRecords {
		return job.SearchResult{
			Data:         []job.Job{},
			Page:         query.Page,
			Limit:        query.Limit,
			TotalPages:   totalPages,
			TotalRecords: totalRecords,
		}, nil
	}

	end := start + query.Limit
	if end > totalRecords {
		end = totalRecords
	}

	return job.SearchResult{
		Data:         append([]job.Job(nil), filtered[start:end]...),
		Page:         query.Page,
		Limit:        query.Limit,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
	}, nil
}

// GetByID returns by id.
func (r *JobsRepository) GetByID(_ context.Context, id string) (job.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, ok := r.byID[strings.TrimSpace(id)]
	if !ok {
		return job.Job{}, job.ErrNotFound
	}

	return record, nil
}

// RecordScrapeRun records scrape run.
func (r *JobsRepository) RecordScrapeRun(_ context.Context, run job.ScrapeRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if run.ID == "" {
		run.ID = "scrape_run_" + randomHex(10)
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.FinishedAt.IsZero() {
		run.FinishedAt = run.StartedAt
	}
	r.scrapeRunLogs = append(r.scrapeRunLogs, run)
	return nil
}

func sourceJobKey(source job.Source, originalJobID string) string {
	return string(source) + "|" + strings.TrimSpace(originalJobID)
}

func matchesQuery(candidate job.Job, query job.SearchQuery) bool {
	if query.Source != "" && candidate.Source != query.Source {
		return false
	}

	if query.Q != "" {
		needle := strings.ToLower(strings.TrimSpace(query.Q))
		haystack := strings.ToLower(strings.Join([]string{
			candidate.Title,
			candidate.Company,
			candidate.Description,
		}, " "))
		if !strings.Contains(haystack, needle) {
			return false
		}
	}

	if query.Location != "" {
		locationNeedle := strings.ToLower(strings.TrimSpace(query.Location))
		if !strings.Contains(strings.ToLower(candidate.Location), locationNeedle) {
			return false
		}
	}

	if query.HasSalaryMin {
		if candidate.SalaryMin == nil || *candidate.SalaryMin < query.SalaryMin {
			return false
		}
	}

	return true
}

func sortJobs(items []job.Job, sortKey string) {
	switch sortKey {
	case "posted_at":
		sort.SliceStable(items, func(i, j int) bool {
			return sortTime(items[i], true).Before(sortTime(items[j], true))
		})
	case "-created_at":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		})
	case "created_at":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].CreatedAt.Before(items[j].CreatedAt)
		})
	default:
		sort.SliceStable(items, func(i, j int) bool {
			return sortTime(items[i], true).After(sortTime(items[j], true))
		})
	}
}

func sortTime(record job.Job, preferPosted bool) time.Time {
	if preferPosted && record.PostedAt != nil {
		return record.PostedAt.UTC()
	}
	if !record.CreatedAt.IsZero() {
		return record.CreatedAt.UTC()
	}
	return time.Time{}
}

func normalizeInput(input job.UpsertInput) (job.UpsertInput, error) {
	input.OriginalJobID = strings.TrimSpace(input.OriginalJobID)
	input.Title = strings.TrimSpace(input.Title)
	input.Company = strings.TrimSpace(input.Company)
	input.Location = strings.TrimSpace(input.Location)
	input.Description = strings.TrimSpace(input.Description)
	input.URL = strings.TrimSpace(input.URL)
	input.SalaryRange = strings.TrimSpace(input.SalaryRange)
	if input.RawData == nil {
		input.RawData = map[string]any{}
	}

	if input.OriginalJobID == "" {
		return job.UpsertInput{}, errors.New("original_job_id is required")
	}
	if input.Title == "" {
		return job.UpsertInput{}, errors.New("title is required")
	}
	if input.URL == "" {
		return job.UpsertInput{}, errors.New("url is required")
	}

	return input, nil
}

func randomHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(b)
}

// SearchTitles searches distinct job titles matching a prefix.
func (r *JobsRepository) SearchTitles(_ context.Context, query job.TitleSearchQuery) ([]string, error) {
	// Memory adapter stub: return empty slice
	return []string{}, nil
}
