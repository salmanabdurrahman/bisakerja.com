package memory

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	trackerdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/tracker"
)

// TrackerRepository represents tracker repository.
type TrackerRepository struct {
	mu                   sync.RWMutex
	bookmarksByID        map[string]trackerdomain.Bookmark
	bookmarksByUnique    map[string]string // userID|jobID -> bookmarkID
	applicationsByID     map[string]trackerdomain.TrackedApplication
	applicationsByUnique map[string]string // userID|jobID -> applicationID
}

// NewTrackerRepository creates a new tracker repository instance.
func NewTrackerRepository() *TrackerRepository {
	return &TrackerRepository{
		bookmarksByID:        make(map[string]trackerdomain.Bookmark),
		bookmarksByUnique:    make(map[string]string),
		applicationsByID:     make(map[string]trackerdomain.TrackedApplication),
		applicationsByUnique: make(map[string]string),
	}
}

// CreateBookmark creates a bookmark.
func (r *TrackerRepository) CreateBookmark(
	_ context.Context,
	input trackerdomain.CreateBookmarkInput,
) (trackerdomain.Bookmark, error) {
	normalizedUserID := strings.TrimSpace(input.UserID)
	normalizedJobID := strings.TrimSpace(input.JobID)
	uniqueKey := buildTrackerUniqueKey(normalizedUserID, normalizedJobID)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bookmarksByUnique[uniqueKey]; exists {
		return trackerdomain.Bookmark{}, trackerdomain.ErrBookmarkAlreadyExists
	}

	record := trackerdomain.Bookmark{
		ID:        "bm_" + randomHex(12),
		UserID:    normalizedUserID,
		JobID:     normalizedJobID,
		CreatedAt: time.Now().UTC(),
	}
	r.bookmarksByID[record.ID] = record
	r.bookmarksByUnique[uniqueKey] = record.ID
	return record, nil
}

// DeleteBookmarkByUserAndJobID deletes bookmark by user and job id.
func (r *TrackerRepository) DeleteBookmarkByUserAndJobID(
	_ context.Context,
	userID string,
	jobID string,
) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedJobID := strings.TrimSpace(jobID)
	uniqueKey := buildTrackerUniqueKey(normalizedUserID, normalizedJobID)

	r.mu.Lock()
	defer r.mu.Unlock()

	bookmarkID, exists := r.bookmarksByUnique[uniqueKey]
	if !exists {
		return trackerdomain.ErrBookmarkNotFound
	}
	delete(r.bookmarksByID, bookmarkID)
	delete(r.bookmarksByUnique, uniqueKey)
	return nil
}

// ListBookmarksByUser returns bookmarks by user.
func (r *TrackerRepository) ListBookmarksByUser(
	_ context.Context,
	userID string,
) ([]trackerdomain.Bookmark, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []trackerdomain.Bookmark{}, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]trackerdomain.Bookmark, 0)
	for _, item := range r.bookmarksByID {
		if item.UserID != normalizedUserID {
			continue
		}
		result = append(result, item)
	}
	slices.SortFunc(result, func(left, right trackerdomain.Bookmark) int {
		if left.CreatedAt.Equal(right.CreatedAt) {
			return strings.Compare(right.ID, left.ID)
		}
		if left.CreatedAt.After(right.CreatedAt) {
			return -1
		}
		return 1
	})
	return result, nil
}

// IsBookmarked checks if a job is bookmarked by the user.
func (r *TrackerRepository) IsBookmarked(
	_ context.Context,
	userID string,
	jobID string,
) (bool, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedJobID := strings.TrimSpace(jobID)
	uniqueKey := buildTrackerUniqueKey(normalizedUserID, normalizedJobID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.bookmarksByUnique[uniqueKey]
	return exists, nil
}

// CreateTrackedApplication creates a tracked application.
func (r *TrackerRepository) CreateTrackedApplication(
	_ context.Context,
	input trackerdomain.CreateTrackedApplicationInput,
) (trackerdomain.TrackedApplication, error) {
	normalizedUserID := strings.TrimSpace(input.UserID)
	normalizedJobID := strings.TrimSpace(input.JobID)
	uniqueKey := buildTrackerUniqueKey(normalizedUserID, normalizedJobID)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.applicationsByUnique[uniqueKey]; exists {
		return trackerdomain.TrackedApplication{}, trackerdomain.ErrApplicationAlreadyExists
	}

	now := time.Now().UTC()
	record := trackerdomain.TrackedApplication{
		ID:        "app_" + randomHex(12),
		UserID:    normalizedUserID,
		JobID:     normalizedJobID,
		Status:    trackerdomain.ApplicationStatusApplied,
		Notes:     strings.TrimSpace(input.Notes),
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.applicationsByID[record.ID] = record
	r.applicationsByUnique[uniqueKey] = record.ID
	return record, nil
}

// UpdateTrackedApplicationStatus updates the status of a tracked application.
func (r *TrackerRepository) UpdateTrackedApplicationStatus(
	_ context.Context,
	input trackerdomain.UpdateTrackedApplicationStatusInput,
) (trackerdomain.TrackedApplication, error) {
	normalizedUserID := strings.TrimSpace(input.UserID)
	normalizedApplicationID := strings.TrimSpace(input.ApplicationID)
	if normalizedUserID == "" || normalizedApplicationID == "" {
		return trackerdomain.TrackedApplication{}, trackerdomain.ErrApplicationNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.applicationsByID[normalizedApplicationID]
	if !exists || item.UserID != normalizedUserID {
		return trackerdomain.TrackedApplication{}, trackerdomain.ErrApplicationNotFound
	}

	item.Status = input.Status
	item.UpdatedAt = time.Now().UTC()
	r.applicationsByID[normalizedApplicationID] = item
	return item, nil
}

// DeleteTrackedApplicationByUserAndID deletes a tracked application.
func (r *TrackerRepository) DeleteTrackedApplicationByUserAndID(
	_ context.Context,
	userID string,
	applicationID string,
) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedApplicationID := strings.TrimSpace(applicationID)
	if normalizedUserID == "" || normalizedApplicationID == "" {
		return trackerdomain.ErrApplicationNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.applicationsByID[normalizedApplicationID]
	if !exists || item.UserID != normalizedUserID {
		return trackerdomain.ErrApplicationNotFound
	}
	uniqueKey := buildTrackerUniqueKey(item.UserID, item.JobID)
	delete(r.applicationsByID, normalizedApplicationID)
	delete(r.applicationsByUnique, uniqueKey)
	return nil
}

// ListTrackedApplicationsByUser returns tracked applications by user.
func (r *TrackerRepository) ListTrackedApplicationsByUser(
	_ context.Context,
	userID string,
) ([]trackerdomain.TrackedApplication, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []trackerdomain.TrackedApplication{}, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]trackerdomain.TrackedApplication, 0)
	for _, item := range r.applicationsByID {
		if item.UserID != normalizedUserID {
			continue
		}
		result = append(result, item)
	}
	slices.SortFunc(result, func(left, right trackerdomain.TrackedApplication) int {
		if left.CreatedAt.Equal(right.CreatedAt) {
			return strings.Compare(right.ID, left.ID)
		}
		if left.CreatedAt.After(right.CreatedAt) {
			return -1
		}
		return 1
	})
	return result, nil
}

// CountActiveTrackedApplicationsByUser counts active tracked applications for a user.
// Active = status is not rejected or withdrawn.
func (r *TrackerRepository) CountActiveTrackedApplicationsByUser(
	_ context.Context,
	userID string,
) (int, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return 0, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, item := range r.applicationsByID {
		if item.UserID != normalizedUserID {
			continue
		}
		if item.Status == trackerdomain.ApplicationStatusRejected || item.Status == trackerdomain.ApplicationStatusWithdrawn {
			continue
		}
		count++
	}
	return count, nil
}

func buildTrackerUniqueKey(userID, jobID string) string {
	return strings.ToLower(strings.TrimSpace(userID)) + "|" + strings.ToLower(strings.TrimSpace(jobID))
}

var _ trackerdomain.Repository = (*TrackerRepository)(nil)
