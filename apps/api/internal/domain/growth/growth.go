package growth

import (
	"context"
	"errors"
	"time"
)

// AlertFrequency represents alert frequency.
type AlertFrequency string

const (
	AlertFrequencyInstant      AlertFrequency = "instant"
	AlertFrequencyDailyDigest  AlertFrequency = "daily_digest"
	AlertFrequencyWeeklyDigest AlertFrequency = "weekly_digest"
)

var (
	ErrSavedSearchNotFound           = errors.New("saved search not found")
	ErrSavedSearchAlreadyExists      = errors.New("saved search already exists")
	ErrWatchlistCompanyNotFound      = errors.New("watchlist company not found")
	ErrWatchlistCompanyAlreadyExists = errors.New("watchlist company already exists")
)

// SavedSearch represents saved search.
type SavedSearch struct {
	ID        string
	UserID    string
	Query     string
	Location  string
	Source    string
	SalaryMin *int64
	Frequency AlertFrequency
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateSavedSearchInput contains input parameters for create saved search.
type CreateSavedSearchInput struct {
	UserID    string
	Query     string
	Location  string
	Source    string
	SalaryMin *int64
	Frequency AlertFrequency
	IsActive  bool
}

// CompanyWatchlist represents company watchlist.
type CompanyWatchlist struct {
	UserID      string
	CompanySlug string
	CreatedAt   time.Time
}

// Repository defines behavior for repository.
type Repository interface {
	CreateSavedSearch(ctx context.Context, input CreateSavedSearchInput) (SavedSearch, error)
	ListSavedSearchesByUser(ctx context.Context, userID string) ([]SavedSearch, error)
	DeleteSavedSearchByUserAndID(ctx context.Context, userID, savedSearchID string) error
	CreateWatchlistCompany(ctx context.Context, userID, companySlug string) (CompanyWatchlist, error)
	ListWatchlistCompaniesByUser(ctx context.Context, userID string) ([]CompanyWatchlist, error)
	DeleteWatchlistCompanyByUserAndSlug(ctx context.Context, userID, companySlug string) error
}
