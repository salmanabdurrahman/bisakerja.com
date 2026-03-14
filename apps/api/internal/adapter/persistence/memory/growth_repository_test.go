package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/growth"
)

func TestGrowthRepository_SavedSearchCRUD(t *testing.T) {
	repository := NewGrowthRepository()
	salaryMin := int64(10_000_000)

	created, err := repository.CreateSavedSearch(context.Background(), growth.CreateSavedSearchInput{
		UserID:    "usr_growth",
		Query:     "golang",
		Location:  "jakarta",
		Source:    "glints",
		SalaryMin: &salaryMin,
		Frequency: growth.AlertFrequencyDailyDigest,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("create saved search: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected saved search id")
	}

	_, err = repository.CreateSavedSearch(context.Background(), growth.CreateSavedSearchInput{
		UserID:    "usr_growth",
		Query:     "golang",
		Location:  "jakarta",
		Source:    "glints",
		SalaryMin: &salaryMin,
		Frequency: growth.AlertFrequencyDailyDigest,
		IsActive:  true,
	})
	if !errors.Is(err, growth.ErrSavedSearchAlreadyExists) {
		t.Fatalf("expected ErrSavedSearchAlreadyExists, got %v", err)
	}

	items, err := repository.ListSavedSearchesByUser(context.Background(), "usr_growth")
	if err != nil {
		t.Fatalf("list saved searches: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one saved search, got %d", len(items))
	}

	if err := repository.DeleteSavedSearchByUserAndID(context.Background(), "usr_growth", created.ID); err != nil {
		t.Fatalf("delete saved search: %v", err)
	}
	if err := repository.DeleteSavedSearchByUserAndID(context.Background(), "usr_growth", created.ID); !errors.Is(err, growth.ErrSavedSearchNotFound) {
		t.Fatalf("expected ErrSavedSearchNotFound, got %v", err)
	}
}

func TestGrowthRepository_WatchlistCRUD(t *testing.T) {
	repository := NewGrowthRepository()

	created, err := repository.CreateWatchlistCompany(context.Background(), "usr_watch", "Acme-Group")
	if err != nil {
		t.Fatalf("create watchlist company: %v", err)
	}
	if created.CompanySlug != "acme-group" {
		t.Fatalf("expected normalized slug acme-group, got %s", created.CompanySlug)
	}

	_, err = repository.CreateWatchlistCompany(context.Background(), "usr_watch", "acme-group")
	if !errors.Is(err, growth.ErrWatchlistCompanyAlreadyExists) {
		t.Fatalf("expected ErrWatchlistCompanyAlreadyExists, got %v", err)
	}

	items, err := repository.ListWatchlistCompaniesByUser(context.Background(), "usr_watch")
	if err != nil {
		t.Fatalf("list watchlist companies: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one watchlist company, got %d", len(items))
	}

	if err := repository.DeleteWatchlistCompanyByUserAndSlug(context.Background(), "usr_watch", "acme-group"); err != nil {
		t.Fatalf("delete watchlist company: %v", err)
	}
	if err := repository.DeleteWatchlistCompanyByUserAndSlug(context.Background(), "usr_watch", "acme-group"); !errors.Is(err, growth.ErrWatchlistCompanyNotFound) {
		t.Fatalf("expected ErrWatchlistCompanyNotFound, got %v", err)
	}
}
