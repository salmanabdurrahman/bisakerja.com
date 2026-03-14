package memory

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/growth"
)

type GrowthRepository struct {
	mu                  sync.RWMutex
	savedSearchByID     map[string]growth.SavedSearch
	savedSearchByUnique map[string]string
	watchlistByUnique   map[string]growth.CompanyWatchlist
}

func NewGrowthRepository() *GrowthRepository {
	return &GrowthRepository{
		savedSearchByID:     make(map[string]growth.SavedSearch),
		savedSearchByUnique: make(map[string]string),
		watchlistByUnique:   make(map[string]growth.CompanyWatchlist),
	}
}

func (r *GrowthRepository) CreateSavedSearch(
	_ context.Context,
	input growth.CreateSavedSearchInput,
) (growth.SavedSearch, error) {
	normalizedUserID := strings.TrimSpace(input.UserID)
	normalizedQuery := strings.TrimSpace(input.Query)
	normalizedLocation := strings.TrimSpace(input.Location)
	normalizedSource := strings.TrimSpace(strings.ToLower(input.Source))
	normalizedFrequency := normalizeFrequency(input.Frequency)
	uniqueKey := buildSavedSearchUniqueKey(
		normalizedUserID,
		normalizedQuery,
		normalizedLocation,
		normalizedSource,
		input.SalaryMin,
		normalizedFrequency,
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.savedSearchByUnique[uniqueKey]; exists {
		return growth.SavedSearch{}, growth.ErrSavedSearchAlreadyExists
	}

	now := time.Now().UTC()
	record := growth.SavedSearch{
		ID:        "ss_" + randomHex(12),
		UserID:    normalizedUserID,
		Query:     normalizedQuery,
		Location:  normalizedLocation,
		Source:    normalizedSource,
		SalaryMin: cloneInt64(input.SalaryMin),
		Frequency: normalizedFrequency,
		IsActive:  input.IsActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if !input.IsActive {
		record.IsActive = false
	} else {
		record.IsActive = true
	}

	r.savedSearchByID[record.ID] = record
	r.savedSearchByUnique[uniqueKey] = record.ID
	return cloneSavedSearch(record), nil
}

func (r *GrowthRepository) ListSavedSearchesByUser(
	_ context.Context,
	userID string,
) ([]growth.SavedSearch, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []growth.SavedSearch{}, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]growth.SavedSearch, 0)
	for _, item := range r.savedSearchByID {
		if item.UserID != normalizedUserID {
			continue
		}
		result = append(result, cloneSavedSearch(item))
	}
	sortSavedSearches(result)
	return result, nil
}

func (r *GrowthRepository) DeleteSavedSearchByUserAndID(
	_ context.Context,
	userID string,
	savedSearchID string,
) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedSavedSearchID := strings.TrimSpace(savedSearchID)
	if normalizedUserID == "" || normalizedSavedSearchID == "" {
		return growth.ErrSavedSearchNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.savedSearchByID[normalizedSavedSearchID]
	if !exists || item.UserID != normalizedUserID {
		return growth.ErrSavedSearchNotFound
	}
	delete(r.savedSearchByID, normalizedSavedSearchID)
	delete(
		r.savedSearchByUnique,
		buildSavedSearchUniqueKey(item.UserID, item.Query, item.Location, item.Source, item.SalaryMin, item.Frequency),
	)
	return nil
}

func (r *GrowthRepository) CreateWatchlistCompany(
	_ context.Context,
	userID string,
	companySlug string,
) (growth.CompanyWatchlist, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedCompanySlug := strings.ToLower(strings.TrimSpace(companySlug))
	uniqueKey := buildWatchlistUniqueKey(normalizedUserID, normalizedCompanySlug)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.watchlistByUnique[uniqueKey]; exists {
		return growth.CompanyWatchlist{}, growth.ErrWatchlistCompanyAlreadyExists
	}

	record := growth.CompanyWatchlist{
		UserID:      normalizedUserID,
		CompanySlug: normalizedCompanySlug,
		CreatedAt:   time.Now().UTC(),
	}
	r.watchlistByUnique[uniqueKey] = record
	return cloneCompanyWatchlist(record), nil
}

func (r *GrowthRepository) ListWatchlistCompaniesByUser(
	_ context.Context,
	userID string,
) ([]growth.CompanyWatchlist, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []growth.CompanyWatchlist{}, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]growth.CompanyWatchlist, 0)
	for _, item := range r.watchlistByUnique {
		if item.UserID != normalizedUserID {
			continue
		}
		result = append(result, cloneCompanyWatchlist(item))
	}
	slices.SortFunc(result, func(left, right growth.CompanyWatchlist) int {
		if left.CreatedAt.Equal(right.CreatedAt) {
			return strings.Compare(right.CompanySlug, left.CompanySlug)
		}
		if left.CreatedAt.After(right.CreatedAt) {
			return -1
		}
		return 1
	})
	return result, nil
}

func (r *GrowthRepository) DeleteWatchlistCompanyByUserAndSlug(
	_ context.Context,
	userID string,
	companySlug string,
) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedCompanySlug := strings.ToLower(strings.TrimSpace(companySlug))
	uniqueKey := buildWatchlistUniqueKey(normalizedUserID, normalizedCompanySlug)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.watchlistByUnique[uniqueKey]; !exists {
		return growth.ErrWatchlistCompanyNotFound
	}
	delete(r.watchlistByUnique, uniqueKey)
	return nil
}

func buildSavedSearchUniqueKey(
	userID string,
	query string,
	location string,
	source string,
	salaryMin *int64,
	frequency growth.AlertFrequency,
) string {
	salaryValue := ""
	if salaryMin != nil {
		salaryValue = strconv.FormatInt(*salaryMin, 10)
	}
	return strings.ToLower(strings.TrimSpace(userID)) + "|" +
		strings.ToLower(strings.TrimSpace(query)) + "|" +
		strings.ToLower(strings.TrimSpace(location)) + "|" +
		strings.ToLower(strings.TrimSpace(source)) + "|" +
		salaryValue + "|" +
		string(normalizeFrequency(frequency))
}

func buildWatchlistUniqueKey(userID string, companySlug string) string {
	return strings.ToLower(strings.TrimSpace(userID)) + "|" + strings.ToLower(strings.TrimSpace(companySlug))
}

func normalizeFrequency(value growth.AlertFrequency) growth.AlertFrequency {
	switch growth.AlertFrequency(strings.ToLower(strings.TrimSpace(string(value)))) {
	case growth.AlertFrequencyDailyDigest:
		return growth.AlertFrequencyDailyDigest
	case growth.AlertFrequencyWeeklyDigest:
		return growth.AlertFrequencyWeeklyDigest
	default:
		return growth.AlertFrequencyInstant
	}
}

func sortSavedSearches(items []growth.SavedSearch) {
	slices.SortFunc(items, func(left, right growth.SavedSearch) int {
		if left.CreatedAt.Equal(right.CreatedAt) {
			return strings.Compare(right.ID, left.ID)
		}
		if left.CreatedAt.After(right.CreatedAt) {
			return -1
		}
		return 1
	})
}

func cloneSavedSearch(value growth.SavedSearch) growth.SavedSearch {
	result := value
	result.SalaryMin = cloneInt64(value.SalaryMin)
	return result
}

func cloneCompanyWatchlist(value growth.CompanyWatchlist) growth.CompanyWatchlist {
	return value
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

var _ growth.Repository = (*GrowthRepository)(nil)
