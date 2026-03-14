package growth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/growth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

var (
	ErrInvalidSavedSearchQuery     = errors.New("saved search query is invalid")
	ErrInvalidSavedSearchLocation  = errors.New("saved search location is invalid")
	ErrInvalidSavedSearchSource    = errors.New("saved search source is invalid")
	ErrInvalidSavedSearchSalaryMin = errors.New("saved search salary_min is invalid")
	ErrInvalidSavedSearchFrequency = errors.New("saved search frequency is invalid")
	ErrInvalidCompanySlug          = errors.New("company slug is invalid")
)

var companySlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,79}$`)

type Service struct {
	identityRepository identity.Repository
	repository         growth.Repository
}

type CreateSavedSearchInput struct {
	UserID    string
	Query     string
	Location  string
	Source    string
	SalaryMin *int64
	Frequency string
	IsActive  bool
}

func NewService(identityRepository identity.Repository, repository growth.Repository) *Service {
	return &Service{
		identityRepository: identityRepository,
		repository:         repository,
	}
}

func (s *Service) CreateSavedSearch(ctx context.Context, input CreateSavedSearchInput) (growth.SavedSearch, error) {
	if s.identityRepository == nil || s.repository == nil {
		return growth.SavedSearch{}, errors.New("growth service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(input.UserID)
	if normalizedUserID == "" {
		return growth.SavedSearch{}, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return growth.SavedSearch{}, fmt.Errorf("get user profile: %w", err)
	}

	query := strings.TrimSpace(input.Query)
	if len(query) < 2 || len(query) > 200 {
		return growth.SavedSearch{}, ErrInvalidSavedSearchQuery
	}

	location := strings.TrimSpace(input.Location)
	if len(location) > 100 {
		return growth.SavedSearch{}, ErrInvalidSavedSearchLocation
	}

	source := strings.TrimSpace(strings.ToLower(input.Source))
	if source != "" {
		if _, ok := job.ParseSource(source); !ok {
			return growth.SavedSearch{}, ErrInvalidSavedSearchSource
		}
	}

	if input.SalaryMin != nil && *input.SalaryMin < 0 {
		return growth.SavedSearch{}, ErrInvalidSavedSearchSalaryMin
	}

	frequency, ok := parseAlertFrequency(input.Frequency)
	if !ok {
		return growth.SavedSearch{}, ErrInvalidSavedSearchFrequency
	}

	created, err := s.repository.CreateSavedSearch(ctx, growth.CreateSavedSearchInput{
		UserID:    normalizedUserID,
		Query:     query,
		Location:  location,
		Source:    source,
		SalaryMin: cloneInt64(input.SalaryMin),
		Frequency: frequency,
		IsActive:  input.IsActive,
	})
	if err != nil {
		return growth.SavedSearch{}, fmt.Errorf("create saved search: %w", err)
	}
	return created, nil
}

func (s *Service) ListSavedSearches(ctx context.Context, userID string) ([]growth.SavedSearch, error) {
	if s.identityRepository == nil || s.repository == nil {
		return nil, errors.New("growth service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	result, err := s.repository.ListSavedSearchesByUser(ctx, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list saved searches: %w", err)
	}
	return result, nil
}

func (s *Service) DeleteSavedSearch(ctx context.Context, userID string, savedSearchID string) error {
	if s.identityRepository == nil || s.repository == nil {
		return errors.New("growth service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return fmt.Errorf("get user profile: %w", err)
	}

	if err := s.repository.DeleteSavedSearchByUserAndID(
		ctx,
		normalizedUserID,
		strings.TrimSpace(savedSearchID),
	); err != nil {
		return fmt.Errorf("delete saved search: %w", err)
	}
	return nil
}

func (s *Service) AddWatchlistCompany(ctx context.Context, userID string, companySlug string) (growth.CompanyWatchlist, error) {
	if s.identityRepository == nil || s.repository == nil {
		return growth.CompanyWatchlist{}, errors.New("growth service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return growth.CompanyWatchlist{}, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return growth.CompanyWatchlist{}, fmt.Errorf("get user profile: %w", err)
	}

	normalizedCompanySlug := strings.ToLower(strings.TrimSpace(companySlug))
	if !companySlugPattern.MatchString(normalizedCompanySlug) {
		return growth.CompanyWatchlist{}, ErrInvalidCompanySlug
	}

	created, err := s.repository.CreateWatchlistCompany(ctx, normalizedUserID, normalizedCompanySlug)
	if err != nil {
		return growth.CompanyWatchlist{}, fmt.Errorf("create watchlist company: %w", err)
	}
	return created, nil
}

func (s *Service) ListWatchlistCompanies(ctx context.Context, userID string) ([]growth.CompanyWatchlist, error) {
	if s.identityRepository == nil || s.repository == nil {
		return nil, errors.New("growth service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	result, err := s.repository.ListWatchlistCompaniesByUser(ctx, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list watchlist companies: %w", err)
	}
	return result, nil
}

func (s *Service) RemoveWatchlistCompany(ctx context.Context, userID string, companySlug string) error {
	if s.identityRepository == nil || s.repository == nil {
		return errors.New("growth service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return fmt.Errorf("get user profile: %w", err)
	}

	normalizedCompanySlug := strings.ToLower(strings.TrimSpace(companySlug))
	if !companySlugPattern.MatchString(normalizedCompanySlug) {
		return ErrInvalidCompanySlug
	}

	if err := s.repository.DeleteWatchlistCompanyByUserAndSlug(ctx, normalizedUserID, normalizedCompanySlug); err != nil {
		return fmt.Errorf("delete watchlist company: %w", err)
	}
	return nil
}

func parseAlertFrequency(raw string) (growth.AlertFrequency, bool) {
	if strings.TrimSpace(raw) == "" {
		return growth.AlertFrequencyInstant, true
	}
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(growth.AlertFrequencyInstant):
		return growth.AlertFrequencyInstant, true
	case string(growth.AlertFrequencyDailyDigest):
		return growth.AlertFrequencyDailyDigest, true
	case string(growth.AlertFrequencyWeeklyDigest):
		return growth.AlertFrequencyWeeklyDigest, true
	default:
		return "", false
	}
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
