package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

type Queue struct {
	pool *pgxpool.Pool
}

func NewQueue(pool *pgxpool.Pool) *Queue {
	return &Queue{pool: pool}
}

func (q *Queue) EnqueueJobEvent(ctx context.Context, event notification.JobEvent) error {
	jobID := strings.TrimSpace(event.JobID)
	if jobID == "" {
		return fmt.Errorf("enqueue job event: job id is required")
	}

	query := `
INSERT INTO notification_job_events (job_id, created_at)
VALUES ($1, now())
`
	if _, err := q.pool.Exec(ctx, query, jobID); err != nil {
		return fmt.Errorf("enqueue job event: %w", err)
	}
	return nil
}

func (q *Queue) DequeueJobEvents(ctx context.Context, limit int) ([]notification.JobEvent, error) {
	if limit <= 0 {
		return []notification.JobEvent{}, nil
	}

	query := `
WITH picked AS (
  SELECT id
  FROM notification_job_events
  ORDER BY created_at ASC, id ASC
  LIMIT $1
  FOR UPDATE SKIP LOCKED
)
DELETE FROM notification_job_events events
USING picked
WHERE events.id = picked.id
RETURNING events.job_id
`

	rows, err := q.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("dequeue job events: %w", err)
	}
	defer rows.Close()

	result := make([]notification.JobEvent, 0, limit)
	for rows.Next() {
		var jobID string
		if err := rows.Scan(&jobID); err != nil {
			return nil, fmt.Errorf("scan dequeued job event: %w", err)
		}
		result = append(result, notification.JobEvent{
			JobID: strings.TrimSpace(jobID),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dequeue job events rows: %w", err)
	}

	return result, nil
}

func (q *Queue) EnqueueDeliveryTask(ctx context.Context, task notification.DeliveryTask) error {
	if strings.TrimSpace(task.NotificationID) == "" {
		return fmt.Errorf("enqueue delivery task: notification id is required")
	}

	query := `
INSERT INTO notification_delivery_tasks (
  notification_id,
  user_id,
  user_email,
  user_name,
  job_id,
  channel,
  job_title,
  company,
  location,
  url,
  created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
`

	if _, err := q.pool.Exec(
		ctx,
		query,
		strings.TrimSpace(task.NotificationID),
		strings.TrimSpace(task.UserID),
		strings.TrimSpace(task.UserEmail),
		strings.TrimSpace(task.UserName),
		strings.TrimSpace(task.JobID),
		strings.TrimSpace(string(task.Channel)),
		strings.TrimSpace(task.JobTitle),
		strings.TrimSpace(task.Company),
		strings.TrimSpace(task.Location),
		strings.TrimSpace(task.URL),
	); err != nil {
		return fmt.Errorf("enqueue delivery task: %w", err)
	}

	return nil
}

func (q *Queue) DequeueDeliveryTasks(ctx context.Context, limit int) ([]notification.DeliveryTask, error) {
	if limit <= 0 {
		return []notification.DeliveryTask{}, nil
	}

	query := `
WITH picked AS (
  SELECT id
  FROM notification_delivery_tasks
  ORDER BY created_at ASC, id ASC
  LIMIT $1
  FOR UPDATE SKIP LOCKED
)
DELETE FROM notification_delivery_tasks tasks
USING picked
WHERE tasks.id = picked.id
RETURNING
  tasks.notification_id,
  tasks.user_id,
  tasks.user_email,
  tasks.user_name,
  tasks.job_id,
  tasks.channel,
  tasks.job_title,
  tasks.company,
  tasks.location,
  tasks.url
`

	rows, err := q.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("dequeue delivery tasks: %w", err)
	}
	defer rows.Close()

	result := make([]notification.DeliveryTask, 0, limit)
	for rows.Next() {
		var (
			task    notification.DeliveryTask
			channel string
		)
		if err := rows.Scan(
			&task.NotificationID,
			&task.UserID,
			&task.UserEmail,
			&task.UserName,
			&task.JobID,
			&channel,
			&task.JobTitle,
			&task.Company,
			&task.Location,
			&task.URL,
		); err != nil {
			return nil, fmt.Errorf("scan dequeued delivery task: %w", err)
		}
		task.Channel = notification.Channel(strings.TrimSpace(channel))
		result = append(result, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dequeue delivery tasks rows: %w", err)
	}

	return result, nil
}

var _ notification.Queue = (*Queue)(nil)
