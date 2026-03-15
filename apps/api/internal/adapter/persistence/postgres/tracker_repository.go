package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	trackerdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/tracker"
)

// TrackerRepository represents tracker repository.
type TrackerRepository struct {
	pool *pgxpool.Pool
}

// NewTrackerRepository creates a new tracker repository instance.
func NewTrackerRepository(pool *pgxpool.Pool) *TrackerRepository {
	return &TrackerRepository{pool: pool}
}

// CreateBookmark creates a bookmark for a job.
func (r *TrackerRepository) CreateBookmark(
	ctx context.Context,
	input trackerdomain.CreateBookmarkInput,
) (trackerdomain.Bookmark, error) {
	normalizedUserID := strings.TrimSpace(input.UserID)
	normalizedJobID := strings.TrimSpace(input.JobID)

	query := `
WITH selected_user AS (
  SELECT id FROM users WHERE id::text = $1
)
INSERT INTO bookmarks (user_id, job_id, created_at)
SELECT selected_user.id, $2, now()
FROM selected_user
ON CONFLICT (user_id, job_id) DO NOTHING
RETURNING id::text, user_id::text, job_id, created_at
`

	created, err := scanBookmark(r.pool.QueryRow(ctx, query, normalizedUserID, normalizedJobID))
	if err == nil {
		return created, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return trackerdomain.Bookmark{}, fmt.Errorf("create bookmark: %w", err)
	}

	// Differentiate duplicate from missing user.
	existsQuery := `
SELECT 1
FROM bookmarks
WHERE user_id::text = $1 AND job_id = $2
`
	var exists int
	existsErr := r.pool.QueryRow(ctx, existsQuery, normalizedUserID, normalizedJobID).Scan(&exists)
	if existsErr == nil {
		return trackerdomain.Bookmark{}, trackerdomain.ErrBookmarkAlreadyExists
	}
	if existsErr != nil && !errors.Is(existsErr, pgx.ErrNoRows) {
		return trackerdomain.Bookmark{}, fmt.Errorf("lookup existing bookmark: %w", existsErr)
	}

	return trackerdomain.Bookmark{}, trackerdomain.ErrBookmarkNotFound
}

// DeleteBookmarkByUserAndJobID deletes a bookmark by user and job id.
func (r *TrackerRepository) DeleteBookmarkByUserAndJobID(
	ctx context.Context,
	userID string,
	jobID string,
) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedJobID := strings.TrimSpace(jobID)

	query := `
DELETE FROM bookmarks
WHERE user_id::text = $1 AND job_id = $2
`

	commandTag, err := r.pool.Exec(ctx, query, normalizedUserID, normalizedJobID)
	if err != nil {
		return fmt.Errorf("delete bookmark: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return trackerdomain.ErrBookmarkNotFound
	}
	return nil
}

// ListBookmarksByUser returns bookmarks for a user.
func (r *TrackerRepository) ListBookmarksByUser(
	ctx context.Context,
	userID string,
) ([]trackerdomain.Bookmark, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []trackerdomain.Bookmark{}, nil
	}

	query := `
SELECT id::text, user_id::text, job_id, created_at
FROM bookmarks
WHERE user_id::text = $1
ORDER BY created_at DESC, id DESC
`

	rows, err := r.pool.Query(ctx, query, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list bookmarks by user: %w", err)
	}
	defer rows.Close()

	result := make([]trackerdomain.Bookmark, 0)
	for rows.Next() {
		item, scanErr := scanBookmark(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list bookmarks by user rows: %w", err)
	}

	return result, nil
}

// IsBookmarked checks if a job is bookmarked by the user.
func (r *TrackerRepository) IsBookmarked(
	ctx context.Context,
	userID string,
	jobID string,
) (bool, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedJobID := strings.TrimSpace(jobID)

	query := `
SELECT 1
FROM bookmarks
WHERE user_id::text = $1 AND job_id = $2
LIMIT 1
`

	var exists int
	err := r.pool.QueryRow(ctx, query, normalizedUserID, normalizedJobID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("is bookmarked: %w", err)
	}
	return true, nil
}

// CreateTrackedApplication creates a tracked application.
func (r *TrackerRepository) CreateTrackedApplication(
	ctx context.Context,
	input trackerdomain.CreateTrackedApplicationInput,
) (trackerdomain.TrackedApplication, error) {
	normalizedUserID := strings.TrimSpace(input.UserID)
	normalizedJobID := strings.TrimSpace(input.JobID)
	normalizedNotes := strings.TrimSpace(input.Notes)

	query := `
WITH selected_user AS (
  SELECT id FROM users WHERE id::text = $1
)
INSERT INTO tracked_applications (user_id, job_id, status, notes, created_at, updated_at)
SELECT selected_user.id, $2, $3, $4, now(), now()
FROM selected_user
ON CONFLICT (user_id, job_id) DO NOTHING
RETURNING id::text, user_id::text, job_id, status, notes, created_at, updated_at
`

	created, err := scanTrackedApplication(r.pool.QueryRow(
		ctx, query,
		normalizedUserID,
		normalizedJobID,
		string(trackerdomain.ApplicationStatusApplied),
		normalizedNotes,
	))
	if err == nil {
		return created, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return trackerdomain.TrackedApplication{}, fmt.Errorf("create tracked application: %w", err)
	}

	// Differentiate duplicate from missing user.
	existsQuery := `
SELECT 1
FROM tracked_applications
WHERE user_id::text = $1 AND job_id = $2
`
	var exists int
	existsErr := r.pool.QueryRow(ctx, existsQuery, normalizedUserID, normalizedJobID).Scan(&exists)
	if existsErr == nil {
		return trackerdomain.TrackedApplication{}, trackerdomain.ErrApplicationAlreadyExists
	}
	if existsErr != nil && !errors.Is(existsErr, pgx.ErrNoRows) {
		return trackerdomain.TrackedApplication{}, fmt.Errorf("lookup existing tracked application: %w", existsErr)
	}

	return trackerdomain.TrackedApplication{}, trackerdomain.ErrApplicationNotFound
}

// UpdateTrackedApplicationStatus updates the status of a tracked application.
func (r *TrackerRepository) UpdateTrackedApplicationStatus(
	ctx context.Context,
	input trackerdomain.UpdateTrackedApplicationStatusInput,
) (trackerdomain.TrackedApplication, error) {
	normalizedUserID := strings.TrimSpace(input.UserID)
	normalizedApplicationID := strings.TrimSpace(input.ApplicationID)

	query := `
UPDATE tracked_applications
SET status = $1, updated_at = now()
WHERE id::text = $2 AND user_id::text = $3
RETURNING id::text, user_id::text, job_id, status, notes, created_at, updated_at
`

	updated, err := scanTrackedApplication(r.pool.QueryRow(ctx, query, string(input.Status), normalizedApplicationID, normalizedUserID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return trackerdomain.TrackedApplication{}, trackerdomain.ErrApplicationNotFound
		}
		return trackerdomain.TrackedApplication{}, fmt.Errorf("update tracked application status: %w", err)
	}
	return updated, nil
}

// DeleteTrackedApplicationByUserAndID deletes a tracked application by user and id.
func (r *TrackerRepository) DeleteTrackedApplicationByUserAndID(
	ctx context.Context,
	userID string,
	applicationID string,
) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedApplicationID := strings.TrimSpace(applicationID)

	query := `
DELETE FROM tracked_applications
WHERE id::text = $1 AND user_id::text = $2
`

	commandTag, err := r.pool.Exec(ctx, query, normalizedApplicationID, normalizedUserID)
	if err != nil {
		return fmt.Errorf("delete tracked application: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return trackerdomain.ErrApplicationNotFound
	}
	return nil
}

// ListTrackedApplicationsByUser returns tracked applications for a user.
func (r *TrackerRepository) ListTrackedApplicationsByUser(
	ctx context.Context,
	userID string,
) ([]trackerdomain.TrackedApplication, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []trackerdomain.TrackedApplication{}, nil
	}

	query := `
SELECT id::text, user_id::text, job_id, status, notes, created_at, updated_at
FROM tracked_applications
WHERE user_id::text = $1
ORDER BY created_at DESC, id DESC
`

	rows, err := r.pool.Query(ctx, query, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list tracked applications by user: %w", err)
	}
	defer rows.Close()

	result := make([]trackerdomain.TrackedApplication, 0)
	for rows.Next() {
		item, scanErr := scanTrackedApplication(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tracked applications by user rows: %w", err)
	}

	return result, nil
}

// CountActiveTrackedApplicationsByUser counts active tracked applications for a user.
// Active means status is not 'rejected' or 'withdrawn'.
func (r *TrackerRepository) CountActiveTrackedApplicationsByUser(
	ctx context.Context,
	userID string,
) (int, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return 0, nil
	}

	query := `
SELECT COUNT(*)
FROM tracked_applications
WHERE user_id::text = $1
  AND status NOT IN ('rejected', 'withdrawn')
`

	var count int
	err := r.pool.QueryRow(ctx, query, normalizedUserID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active tracked applications: %w", err)
	}
	return count, nil
}

type bookmarkScanner interface {
	Scan(dest ...any) error
}

func scanBookmark(scanner bookmarkScanner) (trackerdomain.Bookmark, error) {
	var item trackerdomain.Bookmark
	if err := scanner.Scan(&item.ID, &item.UserID, &item.JobID, &item.CreatedAt); err != nil {
		return trackerdomain.Bookmark{}, err
	}
	item.CreatedAt = item.CreatedAt.UTC()
	return item, nil
}

type trackedApplicationScanner interface {
	Scan(dest ...any) error
}

func scanTrackedApplication(scanner trackedApplicationScanner) (trackerdomain.TrackedApplication, error) {
	var (
		item   trackerdomain.TrackedApplication
		status string
	)
	if err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.JobID,
		&status,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return trackerdomain.TrackedApplication{}, err
	}
	item.Status = trackerdomain.ApplicationStatus(status)
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}

var _ trackerdomain.Repository = (*TrackerRepository)(nil)
