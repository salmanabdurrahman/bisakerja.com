package tracker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	trackerdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/tracker"
)

var (
	ErrInvalidJobID    = errors.New("job id is invalid")
	ErrInvalidNotes    = errors.New("notes is too long")
	ErrPremiumRequired = errors.New("premium subscription is required for unlimited tracking")
)

// Service coordinates application use cases for the tracker package.
type Service struct {
	identityRepository identity.Repository
	repository         trackerdomain.Repository
}

// NewService creates a new service instance.
func NewService(identityRepository identity.Repository, repository trackerdomain.Repository) *Service {
	return &Service{
		identityRepository: identityRepository,
		repository:         repository,
	}
}

// CreateBookmarkInput contains input parameters for create bookmark.
type CreateBookmarkInput struct {
	UserID string
	JobID  string
}

// UpdateApplicationStatusInput contains input parameters for update application status.
type UpdateApplicationStatusInput struct {
	UserID        string
	ApplicationID string
	Status        string
}

// CreateTrackedApplicationInput contains input parameters for create tracked application.
type CreateTrackedApplicationInput struct {
	UserID string
	JobID  string
	Notes  string
}

// CreateBookmark creates a bookmark for a job.
func (s *Service) CreateBookmark(ctx context.Context, input CreateBookmarkInput) (trackerdomain.Bookmark, error) {
	if s.identityRepository == nil || s.repository == nil {
		return trackerdomain.Bookmark{}, errors.New("tracker service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(input.UserID)
	if normalizedUserID == "" {
		return trackerdomain.Bookmark{}, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return trackerdomain.Bookmark{}, fmt.Errorf("get user profile: %w", err)
	}

	normalizedJobID := strings.TrimSpace(input.JobID)
	if len(normalizedJobID) < 1 || len(normalizedJobID) > 100 {
		return trackerdomain.Bookmark{}, ErrInvalidJobID
	}

	created, err := s.repository.CreateBookmark(ctx, trackerdomain.CreateBookmarkInput{
		UserID: normalizedUserID,
		JobID:  normalizedJobID,
	})
	if err != nil {
		return trackerdomain.Bookmark{}, fmt.Errorf("create bookmark: %w", err)
	}
	return created, nil
}

// DeleteBookmark removes a bookmark.
func (s *Service) DeleteBookmark(ctx context.Context, userID, jobID string) error {
	if s.identityRepository == nil || s.repository == nil {
		return errors.New("tracker service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return fmt.Errorf("get user profile: %w", err)
	}

	normalizedJobID := strings.TrimSpace(jobID)
	if len(normalizedJobID) < 1 || len(normalizedJobID) > 100 {
		return ErrInvalidJobID
	}

	if err := s.repository.DeleteBookmarkByUserAndJobID(ctx, normalizedUserID, normalizedJobID); err != nil {
		return fmt.Errorf("delete bookmark: %w", err)
	}
	return nil
}

// ListBookmarks returns bookmarks for a user.
func (s *Service) ListBookmarks(ctx context.Context, userID string) ([]trackerdomain.Bookmark, error) {
	if s.identityRepository == nil || s.repository == nil {
		return nil, errors.New("tracker service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	result, err := s.repository.ListBookmarksByUser(ctx, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list bookmarks: %w", err)
	}
	return result, nil
}

// IsBookmarked checks if a job is bookmarked by the user.
func (s *Service) IsBookmarked(ctx context.Context, userID, jobID string) (bool, error) {
	if s.identityRepository == nil || s.repository == nil {
		return false, errors.New("tracker service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return false, identity.ErrUserNotFound
	}

	normalizedJobID := strings.TrimSpace(jobID)
	if normalizedJobID == "" {
		return false, nil
	}

	return s.repository.IsBookmarked(ctx, normalizedUserID, normalizedJobID)
}

// CreateTrackedApplication creates a tracked application.
func (s *Service) CreateTrackedApplication(ctx context.Context, input CreateTrackedApplicationInput) (trackerdomain.TrackedApplication, error) {
	if s.identityRepository == nil || s.repository == nil {
		return trackerdomain.TrackedApplication{}, errors.New("tracker service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(input.UserID)
	if normalizedUserID == "" {
		return trackerdomain.TrackedApplication{}, identity.ErrUserNotFound
	}
	user, err := s.identityRepository.GetUserByID(ctx, normalizedUserID)
	if err != nil {
		return trackerdomain.TrackedApplication{}, fmt.Errorf("get user profile: %w", err)
	}

	normalizedJobID := strings.TrimSpace(input.JobID)
	if len(normalizedJobID) < 1 || len(normalizedJobID) > 100 {
		return trackerdomain.TrackedApplication{}, ErrInvalidJobID
	}

	normalizedNotes := strings.TrimSpace(input.Notes)
	if len(normalizedNotes) > 2000 {
		return trackerdomain.TrackedApplication{}, ErrInvalidNotes
	}

	if !user.IsPremium {
		count, countErr := s.repository.CountActiveTrackedApplicationsByUser(ctx, normalizedUserID)
		if countErr != nil {
			return trackerdomain.TrackedApplication{}, fmt.Errorf("count active applications: %w", countErr)
		}
		if count >= trackerdomain.FreeTierApplicationLimit {
			return trackerdomain.TrackedApplication{}, trackerdomain.ErrApplicationLimitExceeded
		}
	}

	created, err := s.repository.CreateTrackedApplication(ctx, trackerdomain.CreateTrackedApplicationInput{
		UserID: normalizedUserID,
		JobID:  normalizedJobID,
		Notes:  normalizedNotes,
	})
	if err != nil {
		return trackerdomain.TrackedApplication{}, fmt.Errorf("create tracked application: %w", err)
	}
	return created, nil
}

// UpdateApplicationStatus updates the status of a tracked application.
func (s *Service) UpdateApplicationStatus(ctx context.Context, input UpdateApplicationStatusInput) (trackerdomain.TrackedApplication, error) {
	if s.identityRepository == nil || s.repository == nil {
		return trackerdomain.TrackedApplication{}, errors.New("tracker service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(input.UserID)
	if normalizedUserID == "" {
		return trackerdomain.TrackedApplication{}, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return trackerdomain.TrackedApplication{}, fmt.Errorf("get user profile: %w", err)
	}

	status, ok := parseApplicationStatus(input.Status)
	if !ok {
		return trackerdomain.TrackedApplication{}, trackerdomain.ErrInvalidApplicationStatus
	}

	updated, err := s.repository.UpdateTrackedApplicationStatus(ctx, trackerdomain.UpdateTrackedApplicationStatusInput{
		UserID:        normalizedUserID,
		ApplicationID: strings.TrimSpace(input.ApplicationID),
		Status:        status,
	})
	if err != nil {
		return trackerdomain.TrackedApplication{}, fmt.Errorf("update application status: %w", err)
	}
	return updated, nil
}

// DeleteTrackedApplication removes a tracked application.
func (s *Service) DeleteTrackedApplication(ctx context.Context, userID, applicationID string) error {
	if s.identityRepository == nil || s.repository == nil {
		return errors.New("tracker service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return fmt.Errorf("get user profile: %w", err)
	}

	if err := s.repository.DeleteTrackedApplicationByUserAndID(
		ctx,
		normalizedUserID,
		strings.TrimSpace(applicationID),
	); err != nil {
		return fmt.Errorf("delete tracked application: %w", err)
	}
	return nil
}

// ListTrackedApplications returns tracked applications for a user.
func (s *Service) ListTrackedApplications(ctx context.Context, userID string) ([]trackerdomain.TrackedApplication, error) {
	if s.identityRepository == nil || s.repository == nil {
		return nil, errors.New("tracker service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, identity.ErrUserNotFound
	}
	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	result, err := s.repository.ListTrackedApplicationsByUser(ctx, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list tracked applications: %w", err)
	}
	return result, nil
}

func parseApplicationStatus(raw string) (trackerdomain.ApplicationStatus, bool) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(trackerdomain.ApplicationStatusApplied):
		return trackerdomain.ApplicationStatusApplied, true
	case string(trackerdomain.ApplicationStatusInterview):
		return trackerdomain.ApplicationStatusInterview, true
	case string(trackerdomain.ApplicationStatusOffer):
		return trackerdomain.ApplicationStatusOffer, true
	case string(trackerdomain.ApplicationStatusRejected):
		return trackerdomain.ApplicationStatusRejected, true
	case string(trackerdomain.ApplicationStatusWithdrawn):
		return trackerdomain.ApplicationStatusWithdrawn, true
	default:
		return "", false
	}
}
