package tracker

import (
	"context"
	"errors"
	"time"
)

// ApplicationStatus represents application tracking status.
type ApplicationStatus string

const (
	ApplicationStatusApplied   ApplicationStatus = "applied"
	ApplicationStatusInterview ApplicationStatus = "interview"
	ApplicationStatusOffer     ApplicationStatus = "offer"
	ApplicationStatusRejected  ApplicationStatus = "rejected"
	ApplicationStatusWithdrawn ApplicationStatus = "withdrawn"
)

// FreeTierApplicationLimit is the maximum number of active tracked applications for free users.
const FreeTierApplicationLimit = 5

var (
	ErrBookmarkNotFound         = errors.New("bookmark not found")
	ErrBookmarkAlreadyExists    = errors.New("bookmark already exists")
	ErrApplicationNotFound      = errors.New("tracked application not found")
	ErrApplicationAlreadyExists = errors.New("tracked application already exists")
	ErrApplicationLimitExceeded = errors.New("tracked application limit exceeded")
	ErrInvalidApplicationStatus = errors.New("invalid application status")
)

// Bookmark represents a bookmarked job.
type Bookmark struct {
	ID        string
	UserID    string
	JobID     string
	CreatedAt time.Time
}

// CreateBookmarkInput contains input parameters for create bookmark.
type CreateBookmarkInput struct {
	UserID string
	JobID  string
}

// TrackedApplication represents a tracked job application.
type TrackedApplication struct {
	ID        string
	UserID    string
	JobID     string
	Status    ApplicationStatus
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateTrackedApplicationInput contains input parameters for create tracked application.
type CreateTrackedApplicationInput struct {
	UserID string
	JobID  string
	Notes  string
}

// UpdateTrackedApplicationStatusInput contains input parameters for update tracked application status.
type UpdateTrackedApplicationStatusInput struct {
	UserID        string
	ApplicationID string
	Status        ApplicationStatus
}

// Repository defines behavior for tracker repository.
type Repository interface {
	CreateBookmark(ctx context.Context, input CreateBookmarkInput) (Bookmark, error)
	DeleteBookmarkByUserAndJobID(ctx context.Context, userID, jobID string) error
	ListBookmarksByUser(ctx context.Context, userID string) ([]Bookmark, error)
	IsBookmarked(ctx context.Context, userID, jobID string) (bool, error)

	CreateTrackedApplication(ctx context.Context, input CreateTrackedApplicationInput) (TrackedApplication, error)
	UpdateTrackedApplicationStatus(ctx context.Context, input UpdateTrackedApplicationStatusInput) (TrackedApplication, error)
	DeleteTrackedApplicationByUserAndID(ctx context.Context, userID, applicationID string) error
	ListTrackedApplicationsByUser(ctx context.Context, userID string) ([]TrackedApplication, error)
	CountActiveTrackedApplicationsByUser(ctx context.Context, userID string) (int, error)
}
