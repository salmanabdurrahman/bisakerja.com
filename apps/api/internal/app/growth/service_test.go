package growth

import (
	"context"
	"errors"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	growthdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/growth"
	identitydomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

func setupGrowthService(t *testing.T) (*Service, string) {
	t.Helper()

	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identitydomain.CreateUserInput{
		Email:        "growth@example.com",
		PasswordHash: "hash",
		Name:         "Growth User",
		Role:         identitydomain.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := NewService(identityRepository, memory.NewGrowthRepository())
	return service, user.ID
}

func TestService_SavedSearchCRUD(t *testing.T) {
	service, userID := setupGrowthService(t)
	salaryMin := int64(12_000_000)

	created, err := service.CreateSavedSearch(context.Background(), CreateSavedSearchInput{
		UserID:    userID,
		Query:     "golang backend",
		Location:  "jakarta",
		Source:    "glints",
		SalaryMin: &salaryMin,
		Frequency: "daily_digest",
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("create saved search: %v", err)
	}
	if created.Frequency != growthdomain.AlertFrequencyDailyDigest {
		t.Fatalf("expected daily_digest frequency, got %s", created.Frequency)
	}

	items, err := service.ListSavedSearches(context.Background(), userID)
	if err != nil {
		t.Fatalf("list saved searches: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one saved search, got %d", len(items))
	}

	if err := service.DeleteSavedSearch(context.Background(), userID, created.ID); err != nil {
		t.Fatalf("delete saved search: %v", err)
	}
	if err := service.DeleteSavedSearch(context.Background(), userID, created.ID); !errors.Is(err, growthdomain.ErrSavedSearchNotFound) {
		t.Fatalf("expected ErrSavedSearchNotFound, got %v", err)
	}
}

func TestService_WatchlistCRUD(t *testing.T) {
	service, userID := setupGrowthService(t)

	created, err := service.AddWatchlistCompany(context.Background(), userID, "Acme-Group")
	if err != nil {
		t.Fatalf("add watchlist company: %v", err)
	}
	if created.CompanySlug != "acme-group" {
		t.Fatalf("expected normalized slug acme-group, got %s", created.CompanySlug)
	}

	items, err := service.ListWatchlistCompanies(context.Background(), userID)
	if err != nil {
		t.Fatalf("list watchlist companies: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one watchlist company, got %d", len(items))
	}

	if err := service.RemoveWatchlistCompany(context.Background(), userID, "acme-group"); err != nil {
		t.Fatalf("remove watchlist company: %v", err)
	}
	if err := service.RemoveWatchlistCompany(context.Background(), userID, "acme-group"); !errors.Is(err, growthdomain.ErrWatchlistCompanyNotFound) {
		t.Fatalf("expected ErrWatchlistCompanyNotFound, got %v", err)
	}
}

func TestService_CreateSavedSearch_InvalidFrequency(t *testing.T) {
	service, userID := setupGrowthService(t)

	_, err := service.CreateSavedSearch(context.Background(), CreateSavedSearchInput{
		UserID:    userID,
		Query:     "go",
		Frequency: "hourly",
		IsActive:  true,
	})
	if !errors.Is(err, ErrInvalidSavedSearchFrequency) {
		t.Fatalf("expected ErrInvalidSavedSearchFrequency, got %v", err)
	}
}
