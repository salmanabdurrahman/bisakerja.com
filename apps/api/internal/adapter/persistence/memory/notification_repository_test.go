package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

func TestNotificationRepository_CreatePending_DeduplicatesByUserJobChannel(t *testing.T) {
	repository := NewNotificationRepository()

	first, err := repository.CreatePending(context.Background(), notification.CreateInput{
		UserID:  "usr_1",
		JobID:   "job_1",
		Channel: notification.ChannelEmail,
	})
	if err != nil {
		t.Fatalf("create first notification: %v", err)
	}
	if first.Status != notification.StatusPending {
		t.Fatalf("expected pending status, got %s", first.Status)
	}

	_, err = repository.CreatePending(context.Background(), notification.CreateInput{
		UserID:  "usr_1",
		JobID:   "job_1",
		Channel: notification.ChannelEmail,
	})
	if !errors.Is(err, notification.ErrDuplicateNotification) {
		t.Fatalf("expected duplicate notification error, got %v", err)
	}
}

func TestNotificationRepository_MarkSentAndFailed(t *testing.T) {
	repository := NewNotificationRepository()
	record, err := repository.CreatePending(context.Background(), notification.CreateInput{
		UserID:  "usr_1",
		JobID:   "job_2",
		Channel: notification.ChannelEmail,
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}

	sentAt := time.Now().UTC()
	sent, err := repository.MarkSent(context.Background(), record.ID, sentAt)
	if err != nil {
		t.Fatalf("mark sent: %v", err)
	}
	if sent.Status != notification.StatusSent || sent.SentAt == nil {
		t.Fatalf("expected sent notification with sent_at, got %+v", sent)
	}

	failed, err := repository.MarkFailed(context.Background(), record.ID, "smtp timeout")
	if err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	if failed.Status != notification.StatusFailed {
		t.Fatalf("expected failed status, got %s", failed.Status)
	}
	if failed.ErrorMessage != "smtp timeout" {
		t.Fatalf("expected error message smtp timeout, got %s", failed.ErrorMessage)
	}
}
