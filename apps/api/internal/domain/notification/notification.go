package notification

import (
	"context"
	"errors"
	"time"
)

// Channel represents channel.
type Channel string

const (
	ChannelEmail Channel = "email"
)

// Status describes status details for status.
type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

var (
	ErrNotificationNotFound  = errors.New("notification not found")
	ErrDuplicateNotification = errors.New("duplicate notification")
)

// Notification represents notification.
type Notification struct {
	ID           string
	UserID       string
	JobID        string
	Channel      Channel
	Status       Status
	ErrorMessage string
	SentAt       *time.Time
	ReadAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateInput contains input parameters for create.
type CreateInput struct {
	UserID  string
	JobID   string
	Channel Channel
}

// JobEvent represents job event.
type JobEvent struct {
	JobID string
}

// DeliveryTask represents delivery task.
type DeliveryTask struct {
	NotificationID string
	UserID         string
	UserEmail      string
	UserName       string
	JobID          string
	Channel        Channel
	JobTitle       string
	Company        string
	Location       string
	URL            string
}

// Repository defines behavior for repository.
type Repository interface {
	CreatePending(ctx context.Context, input CreateInput) (Notification, error)
	GetByID(ctx context.Context, notificationID string) (Notification, error)
	ListByUser(ctx context.Context, userID string) ([]Notification, error)
	MarkSent(ctx context.Context, notificationID string, sentAt time.Time) (Notification, error)
	MarkFailed(ctx context.Context, notificationID, errorMessage string) (Notification, error)
	MarkRead(ctx context.Context, notificationID, userID string, readAt time.Time) (Notification, error)
}

// Queue defines behavior for queue.
type Queue interface {
	EnqueueJobEvent(ctx context.Context, event JobEvent) error
	DequeueJobEvents(ctx context.Context, limit int) ([]JobEvent, error)
	EnqueueDeliveryTask(ctx context.Context, task DeliveryTask) error
	DequeueDeliveryTasks(ctx context.Context, limit int) ([]DeliveryTask, error)
}
