package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	notificationqueue "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/queue/memory"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

type fakeEmailSender struct {
	send func(ctx context.Context, message EmailMessage) error
}

func (f fakeEmailSender) Send(ctx context.Context, message EmailMessage) error {
	if f.send != nil {
		return f.send(ctx, message)
	}
	return nil
}

func TestNotifier_RunOnce_MarksSentOnSuccess(t *testing.T) {
	ctx := context.Background()
	repository := memory.NewNotificationRepository()
	queue := notificationqueue.NewQueue()

	record, err := repository.CreatePending(ctx, notification.CreateInput{
		UserID:  "usr_1",
		JobID:   "job_1",
		Channel: notification.ChannelEmail,
	})
	if err != nil {
		t.Fatalf("create pending notification: %v", err)
	}

	if err := queue.EnqueueDeliveryTask(ctx, notification.DeliveryTask{
		NotificationID: record.ID,
		UserID:         "usr_1",
		UserEmail:      "user@example.com",
		UserName:       "Budi",
		JobID:          "job_1",
		Channel:        notification.ChannelEmail,
		JobTitle:       "Backend Engineer",
		Company:        "Bisakerja",
		Location:       "Jakarta",
		URL:            "https://example.com/jobs/job_1",
	}); err != nil {
		t.Fatalf("enqueue delivery task: %v", err)
	}

	notifier := NewNotifier(nil, repository, queue, fakeEmailSender{}, 10)
	summary, err := notifier.RunOnce(ctx)
	if err != nil {
		t.Fatalf("run notifier: %v", err)
	}
	if summary.SentCount != 1 || summary.FailedCount != 0 {
		t.Fatalf("unexpected notifier summary: %+v", summary)
	}

	updated, err := repository.GetByID(ctx, record.ID)
	if err != nil {
		t.Fatalf("get notification by id: %v", err)
	}
	if updated.Status != notification.StatusSent {
		t.Fatalf("expected sent status, got %s", updated.Status)
	}
}

func TestNotifier_RunOnce_MarksFailedOnSendError(t *testing.T) {
	ctx := context.Background()
	repository := memory.NewNotificationRepository()
	queue := notificationqueue.NewQueue()

	record, err := repository.CreatePending(ctx, notification.CreateInput{
		UserID:  "usr_2",
		JobID:   "job_2",
		Channel: notification.ChannelEmail,
	})
	if err != nil {
		t.Fatalf("create pending notification: %v", err)
	}

	if err := queue.EnqueueDeliveryTask(ctx, notification.DeliveryTask{
		NotificationID: record.ID,
		UserID:         "usr_2",
		UserEmail:      "user2@example.com",
		UserName:       "Siti",
		JobID:          "job_2",
		Channel:        notification.ChannelEmail,
		JobTitle:       "Backend Engineer",
		Company:        "Bisakerja",
		Location:       "Remote",
		URL:            "https://example.com/jobs/job_2",
	}); err != nil {
		t.Fatalf("enqueue delivery task: %v", err)
	}

	notifier := NewNotifier(nil, repository, queue, fakeEmailSender{
		send: func(_ context.Context, _ EmailMessage) error {
			return errors.New("smtp unavailable")
		},
	}, 10)
	summary, err := notifier.RunOnce(ctx)
	if err != nil {
		t.Fatalf("run notifier: %v", err)
	}
	if summary.FailedCount != 1 {
		t.Fatalf("expected failed count 1, got %+v", summary)
	}

	updated, err := repository.GetByID(ctx, record.ID)
	if err != nil {
		t.Fatalf("get notification by id: %v", err)
	}
	if updated.Status != notification.StatusFailed {
		t.Fatalf("expected failed status, got %s", updated.Status)
	}
	if updated.ErrorMessage == "" {
		t.Fatal("expected failed notification to contain error message")
	}
}
