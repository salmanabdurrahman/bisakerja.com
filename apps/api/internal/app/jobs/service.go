package jobs

import (
	"context"
	"fmt"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

type Service struct {
	repository job.Repository
}

func NewService(repository job.Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Search(ctx context.Context, query job.SearchQuery) (job.SearchResult, error) {
	result, err := s.repository.Search(ctx, query)
	if err != nil {
		return job.SearchResult{}, fmt.Errorf("search jobs: %w", err)
	}

	return result, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (job.Job, error) {
	result, err := s.repository.GetByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return job.Job{}, fmt.Errorf("get job by id: %w", err)
	}

	return result, nil
}

func (s *Service) UpsertFromScrape(ctx context.Context, source job.Source, jobs []job.UpsertInput) (job.UpsertResult, error) {
	result, err := s.repository.UpsertMany(ctx, source, jobs)
	if err != nil {
		return job.UpsertResult{}, fmt.Errorf("upsert jobs from scrape: %w", err)
	}

	return result, nil
}

func (s *Service) RecordScrapeRun(ctx context.Context, run job.ScrapeRun) error {
	if err := s.repository.RecordScrapeRun(ctx, run); err != nil {
		return fmt.Errorf("record scrape run: %w", err)
	}

	return nil
}
