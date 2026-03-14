package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

// NotificationRepository represents notification repository.
type NotificationRepository struct {
	pool *pgxpool.Pool
}

// NewNotificationRepository creates a new notification repository instance.
func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

// CreatePending creates pending.
func (r *NotificationRepository) CreatePending(
	ctx context.Context,
	input notification.CreateInput,
) (notification.Notification, error) {
	userID := strings.TrimSpace(input.UserID)
	jobID := strings.TrimSpace(input.JobID)
	channel := notification.Channel(strings.TrimSpace(string(input.Channel)))
	if userID == "" || jobID == "" || channel == "" {
		return notification.Notification{}, fmt.Errorf("create notification: user_id, job_id, and channel are required")
	}

	insertQuery := `
WITH selected_user AS (
  SELECT id FROM users WHERE id::text = $1
),
selected_job AS (
  SELECT id FROM jobs WHERE id::text = $2
)
INSERT INTO notifications (user_id, job_id, channel, status, created_at, updated_at)
SELECT selected_user.id, selected_job.id, $3, $4, now(), now()
FROM selected_user, selected_job
ON CONFLICT (user_id, job_id, channel) DO NOTHING
RETURNING id::text, user_id::text, job_id::text, channel, status, COALESCE(error_message, ''), sent_at, read_at, created_at, updated_at
`

	created, err := scanNotification(
		r.pool.QueryRow(
			ctx,
			insertQuery,
			userID,
			jobID,
			string(channel),
			string(notification.StatusPending),
		),
	)
	if err == nil {
		return created, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return notification.Notification{}, fmt.Errorf("create pending notification: %w", err)
	}

	existsQuery := `
SELECT id
FROM notifications n
JOIN users u ON u.id = n.user_id
JOIN jobs j ON j.id = n.job_id
WHERE u.id::text = $1 AND j.id::text = $2 AND n.channel = $3
LIMIT 1
`
	var existingID string
	existsErr := r.pool.QueryRow(ctx, existsQuery, userID, jobID, string(channel)).Scan(&existingID)
	if existsErr == nil {
		return notification.Notification{}, notification.ErrDuplicateNotification
	}
	if !errors.Is(existsErr, pgx.ErrNoRows) {
		return notification.Notification{}, fmt.Errorf("lookup existing notification: %w", existsErr)
	}

	return notification.Notification{}, fmt.Errorf("create pending notification: user or job not found")
}

// GetByID returns by id.
func (r *NotificationRepository) GetByID(ctx context.Context, notificationID string) (notification.Notification, error) {
	normalizedID := strings.TrimSpace(notificationID)
	if normalizedID == "" {
		return notification.Notification{}, notification.ErrNotificationNotFound
	}

	query := `
SELECT id::text, user_id::text, job_id::text, channel, status, COALESCE(error_message, ''), sent_at, read_at, created_at, updated_at
FROM notifications
WHERE id::text = $1
`

	item, err := scanNotification(r.pool.QueryRow(ctx, query, normalizedID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notification.Notification{}, notification.ErrNotificationNotFound
		}
		return notification.Notification{}, fmt.Errorf("get notification by id: %w", err)
	}
	return item, nil
}

// ListByUser returns a list of by user.
func (r *NotificationRepository) ListByUser(ctx context.Context, userID string) ([]notification.Notification, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []notification.Notification{}, nil
	}

	query := `
SELECT id::text, user_id::text, job_id::text, channel, status, COALESCE(error_message, ''), sent_at, read_at, created_at, updated_at
FROM notifications
WHERE user_id::text = $1
ORDER BY created_at DESC, id DESC
`

	rows, err := r.pool.Query(ctx, query, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list notifications by user: %w", err)
	}
	defer rows.Close()

	result := make([]notification.Notification, 0)
	for rows.Next() {
		item, scanErr := scanNotification(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list notifications by user rows: %w", err)
	}

	return result, nil
}

// MarkSent marks sent.
func (r *NotificationRepository) MarkSent(
	ctx context.Context,
	notificationID string,
	sentAt time.Time,
) (notification.Notification, error) {
	query := `
UPDATE notifications
SET status = $2, sent_at = $3, error_message = '', updated_at = $3
WHERE id::text = $1
RETURNING id::text, user_id::text, job_id::text, channel, status, COALESCE(error_message, ''), sent_at, read_at, created_at, updated_at
`

	item, err := scanNotification(
		r.pool.QueryRow(ctx, query, strings.TrimSpace(notificationID), string(notification.StatusSent), sentAt.UTC()),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notification.Notification{}, notification.ErrNotificationNotFound
		}
		return notification.Notification{}, fmt.Errorf("mark notification sent: %w", err)
	}
	return item, nil
}

// MarkFailed marks failed.
func (r *NotificationRepository) MarkFailed(
	ctx context.Context,
	notificationID string,
	errorMessage string,
) (notification.Notification, error) {
	query := `
UPDATE notifications
SET status = $2, error_message = $3, updated_at = now()
WHERE id::text = $1
RETURNING id::text, user_id::text, job_id::text, channel, status, COALESCE(error_message, ''), sent_at, read_at, created_at, updated_at
`

	item, err := scanNotification(
		r.pool.QueryRow(
			ctx,
			query,
			strings.TrimSpace(notificationID),
			string(notification.StatusFailed),
			strings.TrimSpace(errorMessage),
		),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notification.Notification{}, notification.ErrNotificationNotFound
		}
		return notification.Notification{}, fmt.Errorf("mark notification failed: %w", err)
	}
	return item, nil
}

// MarkRead marks read.
func (r *NotificationRepository) MarkRead(
	ctx context.Context,
	notificationID string,
	userID string,
	readAt time.Time,
) (notification.Notification, error) {
	normalizedNotificationID := strings.TrimSpace(notificationID)
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedNotificationID == "" || normalizedUserID == "" {
		return notification.Notification{}, notification.ErrNotificationNotFound
	}

	query := `
UPDATE notifications
SET read_at = COALESCE(read_at, $3),
    updated_at = CASE WHEN read_at IS NULL THEN $3 ELSE updated_at END
WHERE id::text = $1 AND user_id::text = $2
RETURNING id::text, user_id::text, job_id::text, channel, status, COALESCE(error_message, ''), sent_at, read_at, created_at, updated_at
`

	item, err := scanNotification(
		r.pool.QueryRow(ctx, query, normalizedNotificationID, normalizedUserID, readAt.UTC()),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notification.Notification{}, notification.ErrNotificationNotFound
		}
		return notification.Notification{}, fmt.Errorf("mark notification read: %w", err)
	}
	return item, nil
}

type notificationScanner interface {
	Scan(dest ...any) error
}

func scanNotification(scanner notificationScanner) (notification.Notification, error) {
	var (
		item    notification.Notification
		channel string
		status  string
		sentAt  sql.NullTime
		readAt  sql.NullTime
	)

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.JobID,
		&channel,
		&status,
		&item.ErrorMessage,
		&sentAt,
		&readAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return notification.Notification{}, err
	}

	item.Channel = notification.Channel(channel)
	item.Status = notification.Status(status)
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	if sentAt.Valid {
		value := sentAt.Time.UTC()
		item.SentAt = &value
	}
	if readAt.Valid {
		value := readAt.Time.UTC()
		item.ReadAt = &value
	}

	return item, nil
}

var _ notification.Repository = (*NotificationRepository)(nil)
