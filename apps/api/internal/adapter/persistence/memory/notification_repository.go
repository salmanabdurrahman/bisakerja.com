package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

type NotificationRepository struct {
	mu        sync.RWMutex
	byID      map[string]notification.Notification
	byUniqKey map[string]string
}

func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{
		byID:      make(map[string]notification.Notification),
		byUniqKey: make(map[string]string),
	}
}

func (r *NotificationRepository) CreatePending(_ context.Context, input notification.CreateInput) (notification.Notification, error) {
	userID := strings.TrimSpace(input.UserID)
	jobID := strings.TrimSpace(input.JobID)
	channel := notification.Channel(strings.TrimSpace(string(input.Channel)))
	if userID == "" || jobID == "" || channel == "" {
		return notification.Notification{}, fmt.Errorf("create notification: user_id, job_id, and channel are required")
	}

	now := time.Now().UTC()
	uniqueKey := notificationUniqueKey(userID, jobID, channel)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byUniqKey[uniqueKey]; exists {
		return notification.Notification{}, notification.ErrDuplicateNotification
	}

	record := notification.Notification{
		ID:        "notif_" + randomHex(12),
		UserID:    userID,
		JobID:     jobID,
		Channel:   channel,
		Status:    notification.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.byID[record.ID] = record
	r.byUniqKey[uniqueKey] = record.ID
	return record, nil
}

func (r *NotificationRepository) GetByID(_ context.Context, notificationID string) (notification.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, ok := r.byID[strings.TrimSpace(notificationID)]
	if !ok {
		return notification.Notification{}, notification.ErrNotificationNotFound
	}
	return record, nil
}

func (r *NotificationRepository) MarkSent(_ context.Context, notificationID string, sentAt time.Time) (notification.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, ok := r.byID[strings.TrimSpace(notificationID)]
	if !ok {
		return notification.Notification{}, notification.ErrNotificationNotFound
	}

	sentTime := sentAt.UTC()
	record.Status = notification.StatusSent
	record.SentAt = &sentTime
	record.ErrorMessage = ""
	record.UpdatedAt = sentTime
	r.byID[record.ID] = record
	return record, nil
}

func (r *NotificationRepository) MarkFailed(_ context.Context, notificationID, errorMessage string) (notification.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, ok := r.byID[strings.TrimSpace(notificationID)]
	if !ok {
		return notification.Notification{}, notification.ErrNotificationNotFound
	}

	record.Status = notification.StatusFailed
	record.ErrorMessage = strings.TrimSpace(errorMessage)
	record.UpdatedAt = time.Now().UTC()
	r.byID[record.ID] = record
	return record, nil
}

func notificationUniqueKey(userID, jobID string, channel notification.Channel) string {
	return userID + "|" + jobID + "|" + string(channel)
}
