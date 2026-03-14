package notification

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

// EmailMessage represents email message.
type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

// EmailSender defines behavior for email sender.
type EmailSender interface {
	Send(ctx context.Context, message EmailMessage) error
}

// Notifier represents notifier.
type Notifier struct {
	logger                 *slog.Logger
	notificationRepository notification.Repository
	queue                  notification.Queue
	emailSender            EmailSender
	batchSize              int
	now                    func() time.Time
}

// NotifySummary summarizes execution details for notify.
type NotifySummary struct {
	ProcessedTasks int
	SentCount      int
	FailedCount    int
}

// NewNotifier creates a new notifier instance.
func NewNotifier(
	logger *slog.Logger,
	notificationRepository notification.Repository,
	queue notification.Queue,
	emailSender EmailSender,
	batchSize int,
) *Notifier {
	if logger == nil {
		logger = slog.Default()
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	return &Notifier{
		logger:                 logger,
		notificationRepository: notificationRepository,
		queue:                  queue,
		emailSender:            emailSender,
		batchSize:              batchSize,
		now:                    func() time.Time { return time.Now().UTC() },
	}
}

// RunOnce runs once.
func (n *Notifier) RunOnce(ctx context.Context) (NotifySummary, error) {
	if n.notificationRepository == nil || n.queue == nil || n.emailSender == nil {
		return NotifySummary{}, errors.New("notifier dependency is not fully configured")
	}

	tasks, err := n.queue.DequeueDeliveryTasks(ctx, n.batchSize)
	if err != nil {
		return NotifySummary{}, fmt.Errorf("dequeue delivery tasks: %w", err)
	}

	summary := NotifySummary{ProcessedTasks: len(tasks)}
	for _, task := range tasks {
		emailMessage := EmailMessage{
			To:      task.UserEmail,
			Subject: buildEmailSubject(task),
			Body:    buildEmailBody(task),
		}

		sendErr := n.emailSender.Send(ctx, emailMessage)
		if sendErr != nil {
			_, markErr := n.notificationRepository.MarkFailed(ctx, task.NotificationID, sendErr.Error())
			if markErr != nil {
				return summary, fmt.Errorf("mark notification failed %s: %w", task.NotificationID, markErr)
			}
			summary.FailedCount++
			continue
		}

		_, markErr := n.notificationRepository.MarkSent(ctx, task.NotificationID, n.now())
		if markErr != nil {
			return summary, fmt.Errorf("mark notification sent %s: %w", task.NotificationID, markErr)
		}
		summary.SentCount++
	}

	if summary.ProcessedTasks > 0 {
		n.logger.Info(
			"notifier run completed",
			"processed", summary.ProcessedTasks,
			"sent", summary.SentCount,
			"failed", summary.FailedCount,
		)
	}

	return summary, nil
}

func buildEmailSubject(task notification.DeliveryTask) string {
	return "Lowongan baru cocok: " + task.JobTitle
}

func buildEmailBody(task notification.DeliveryTask) string {
	return "Halo " + task.UserName + ",\n\n" +
		"Lowongan baru yang cocok untukmu:\n" +
		"- Posisi: " + task.JobTitle + "\n" +
		"- Perusahaan: " + task.Company + "\n" +
		"- Lokasi: " + task.Location + "\n" +
		"- Link: " + task.URL + "\n\n" +
		"Semoga membantu pencarian kerjamu."
}
