package notification

import (
	"context"
	"errors"
	"time"
)

type Channel string

const (
	ChannelEmail Channel = "email"
)

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

type Notification struct {
	ID           string
	UserID       string
	JobID        string
	Channel      Channel
	Status       Status
	ErrorMessage string
	SentAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateInput struct {
	UserID  string
	JobID   string
	Channel Channel
}

type JobEvent struct {
	JobID string
}

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

type Repository interface {
	CreatePending(ctx context.Context, input CreateInput) (Notification, error)
	GetByID(ctx context.Context, notificationID string) (Notification, error)
	MarkSent(ctx context.Context, notificationID string, sentAt time.Time) (Notification, error)
	MarkFailed(ctx context.Context, notificationID, errorMessage string) (Notification, error)
}

type Queue interface {
	EnqueueJobEvent(ctx context.Context, event JobEvent) error
	DequeueJobEvents(ctx context.Context, limit int) ([]JobEvent, error)
	EnqueueDeliveryTask(ctx context.Context, task DeliveryTask) error
	DequeueDeliveryTasks(ctx context.Context, limit int) ([]DeliveryTask, error)
}
