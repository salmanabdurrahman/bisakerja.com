package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/growth"
)

type GrowthRepository struct {
	pool *pgxpool.Pool
}

func NewGrowthRepository(pool *pgxpool.Pool) *GrowthRepository {
	return &GrowthRepository{pool: pool}
}

func (r *GrowthRepository) CreateSavedSearch(
	ctx context.Context,
	input growth.CreateSavedSearchInput,
) (growth.SavedSearch, error) {
	normalizedUserID := strings.TrimSpace(input.UserID)
	normalizedQuery := strings.TrimSpace(input.Query)
	normalizedLocation := strings.TrimSpace(input.Location)
	normalizedSource := strings.TrimSpace(strings.ToLower(input.Source))
	normalizedFrequency := normalizeFrequency(input.Frequency)

	// Case-insensitive dedupe check (same behavior as memory repository).
	checkQuery := `
SELECT id::text
FROM saved_searches
WHERE user_id::text = $1
  AND LOWER(query) = LOWER($2)
  AND LOWER(location) = LOWER($3)
  AND LOWER(source) = LOWER($4)
  AND (($5::bigint IS NULL AND salary_min IS NULL) OR salary_min = $5)
  AND frequency = $6
LIMIT 1
`

	var existingID string
	checkErr := r.pool.QueryRow(
		ctx,
		checkQuery,
		normalizedUserID,
		normalizedQuery,
		normalizedLocation,
		normalizedSource,
		nullableInt64(input.SalaryMin),
		string(normalizedFrequency),
	).Scan(&existingID)
	if checkErr == nil {
		return growth.SavedSearch{}, growth.ErrSavedSearchAlreadyExists
	}
	if checkErr != nil && !errors.Is(checkErr, pgx.ErrNoRows) {
		return growth.SavedSearch{}, fmt.Errorf("check saved search dedupe: %w", checkErr)
	}

	insertQuery := `
WITH selected_user AS (
  SELECT id FROM users WHERE id::text = $1
)
INSERT INTO saved_searches (user_id, query, location, source, salary_min, frequency, is_active, created_at, updated_at)
SELECT selected_user.id, $2, $3, $4, $5, $6, $7, now(), now()
FROM selected_user
RETURNING id::text, user_id::text, query, location, source, salary_min, frequency, is_active, created_at, updated_at
`

	created, err := scanSavedSearch(
		r.pool.QueryRow(
			ctx,
			insertQuery,
			normalizedUserID,
			normalizedQuery,
			normalizedLocation,
			normalizedSource,
			nullableInt64(input.SalaryMin),
			string(normalizedFrequency),
			input.IsActive,
		),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return growth.SavedSearch{}, growth.ErrSavedSearchNotFound
		}
		return growth.SavedSearch{}, fmt.Errorf("create saved search: %w", err)
	}

	return created, nil
}

func (r *GrowthRepository) ListSavedSearchesByUser(
	ctx context.Context,
	userID string,
) ([]growth.SavedSearch, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []growth.SavedSearch{}, nil
	}

	query := `
SELECT id::text, user_id::text, query, location, source, salary_min, frequency, is_active, created_at, updated_at
FROM saved_searches
WHERE user_id::text = $1
ORDER BY created_at DESC, id DESC
`

	rows, err := r.pool.Query(ctx, query, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list saved searches by user: %w", err)
	}
	defer rows.Close()

	result := make([]growth.SavedSearch, 0)
	for rows.Next() {
		item, scanErr := scanSavedSearch(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list saved searches by user rows: %w", err)
	}

	return result, nil
}

func (r *GrowthRepository) DeleteSavedSearchByUserAndID(
	ctx context.Context,
	userID string,
	savedSearchID string,
) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedSavedSearchID := strings.TrimSpace(savedSearchID)
	if normalizedUserID == "" || normalizedSavedSearchID == "" {
		return growth.ErrSavedSearchNotFound
	}

	query := `
DELETE FROM saved_searches
WHERE id::text = $1 AND user_id::text = $2
`

	commandTag, err := r.pool.Exec(ctx, query, normalizedSavedSearchID, normalizedUserID)
	if err != nil {
		return fmt.Errorf("delete saved search: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return growth.ErrSavedSearchNotFound
	}
	return nil
}

func (r *GrowthRepository) CreateWatchlistCompany(
	ctx context.Context,
	userID string,
	companySlug string,
) (growth.CompanyWatchlist, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedCompanySlug := strings.ToLower(strings.TrimSpace(companySlug))

	query := `
WITH selected_user AS (
  SELECT id FROM users WHERE id::text = $1
)
INSERT INTO watchlist_companies (user_id, company_slug, created_at)
SELECT selected_user.id, $2, now()
FROM selected_user
ON CONFLICT (user_id, company_slug) DO NOTHING
RETURNING user_id::text, company_slug, created_at
`

	created, err := scanWatchlistCompany(r.pool.QueryRow(ctx, query, normalizedUserID, normalizedCompanySlug))
	if err == nil {
		return created, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return growth.CompanyWatchlist{}, fmt.Errorf("create watchlist company: %w", err)
	}

	// Differentiate duplicate from missing user.
	existsQuery := `
SELECT 1
FROM watchlist_companies
WHERE user_id::text = $1 AND company_slug = $2
`
	var exists int
	existsErr := r.pool.QueryRow(ctx, existsQuery, normalizedUserID, normalizedCompanySlug).Scan(&exists)
	if existsErr == nil {
		return growth.CompanyWatchlist{}, growth.ErrWatchlistCompanyAlreadyExists
	}
	if existsErr != nil && !errors.Is(existsErr, pgx.ErrNoRows) {
		return growth.CompanyWatchlist{}, fmt.Errorf("lookup existing watchlist company: %w", existsErr)
	}

	return growth.CompanyWatchlist{}, growth.ErrWatchlistCompanyNotFound
}

func (r *GrowthRepository) ListWatchlistCompaniesByUser(
	ctx context.Context,
	userID string,
) ([]growth.CompanyWatchlist, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []growth.CompanyWatchlist{}, nil
	}

	query := `
SELECT user_id::text, company_slug, created_at
FROM watchlist_companies
WHERE user_id::text = $1
ORDER BY created_at DESC, company_slug DESC
`

	rows, err := r.pool.Query(ctx, query, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list watchlist companies by user: %w", err)
	}
	defer rows.Close()

	result := make([]growth.CompanyWatchlist, 0)
	for rows.Next() {
		item, scanErr := scanWatchlistCompany(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list watchlist companies by user rows: %w", err)
	}

	return result, nil
}

func (r *GrowthRepository) DeleteWatchlistCompanyByUserAndSlug(
	ctx context.Context,
	userID string,
	companySlug string,
) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedCompanySlug := strings.ToLower(strings.TrimSpace(companySlug))

	query := `
DELETE FROM watchlist_companies
WHERE user_id::text = $1 AND company_slug = $2
`

	commandTag, err := r.pool.Exec(ctx, query, normalizedUserID, normalizedCompanySlug)
	if err != nil {
		return fmt.Errorf("delete watchlist company: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return growth.ErrWatchlistCompanyNotFound
	}
	return nil
}

type savedSearchScanner interface {
	Scan(dest ...any) error
}

func scanSavedSearch(scanner savedSearchScanner) (growth.SavedSearch, error) {
	var (
		item      growth.SavedSearch
		salaryMin sql.NullInt64
		frequency string
	)

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.Query,
		&item.Location,
		&item.Source,
		&salaryMin,
		&frequency,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return growth.SavedSearch{}, err
	}

	item.Frequency = growth.AlertFrequency(frequency)
	if salaryMin.Valid {
		value := salaryMin.Int64
		item.SalaryMin = &value
	}
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}

type watchlistScanner interface {
	Scan(dest ...any) error
}

func scanWatchlistCompany(scanner watchlistScanner) (growth.CompanyWatchlist, error) {
	var item growth.CompanyWatchlist
	if err := scanner.Scan(&item.UserID, &item.CompanySlug, &item.CreatedAt); err != nil {
		return growth.CompanyWatchlist{}, err
	}
	item.CreatedAt = item.CreatedAt.UTC()
	return item, nil
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

var _ growth.Repository = (*GrowthRepository)(nil)
