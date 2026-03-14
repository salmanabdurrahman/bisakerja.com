package integration

import (
	"context"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	queuememory "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/queue/memory"
	notificationapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/notification"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

type integrationEmailSender struct{}

func (integrationEmailSender) Send(_ context.Context, _ notificationapp.EmailMessage) error {
	return nil
}

func TestNotificationFlow_MatcherToNotifier(t *testing.T) {
	ctx := context.Background()
	jobsRepository := memory.NewJobsRepository()
	identityRepository := memory.NewIdentityRepository()
	notificationRepository := memory.NewNotificationRepository()
	queue := queuememory.NewQueue()

	expiredAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	premiumUser, err := identityRepository.CreateUser(ctx, identity.CreateUserInput{
		Email:            "premium-flow@example.com",
		PasswordHash:     "hash",
		Name:             "Premium Flow",
		Role:             identity.RoleUser,
		IsPremium:        true,
		PremiumExpiredAt: &expiredAt,
	})
	if err != nil {
		t.Fatalf("create premium user: %v", err)
	}
	now := time.Now().UTC()
	_, err = identityRepository.SavePreferences(ctx, identity.Preferences{
		UserID:    premiumUser.ID,
		Keywords:  []string{"golang"},
		Locations: []string{"remote"},
		JobTypes:  []string{"fulltime"},
		SalaryMin: 8000000,
		UpdatedAt: &now,
	})
	if err != nil {
		t.Fatalf("save preferences: %v", err)
	}

	salaryMin := int64(10000000)
	upserted, err := jobsRepository.UpsertMany(ctx, job.SourceKalibrr, []job.UpsertInput{{
		OriginalJobID: "k-777",
		Title:         "Golang Backend Engineer",
		Company:       "Acme",
		Location:      "Remote",
		Description:   "Build backend with Golang",
		URL:           "https://example.com/jobs/k-777",
		SalaryMin:     &salaryMin,
		RawData: map[string]any{
			"job_type": "fulltime",
		},
	}})
	if err != nil {
		t.Fatalf("upsert job: %v", err)
	}
	jobID := upserted.Inserted[0].ID
	if err := queue.EnqueueJobEvent(ctx, notification.JobEvent{JobID: jobID}); err != nil {
		t.Fatalf("enqueue job event: %v", err)
	}

	matcher := notificationapp.NewMatcher(nil, jobsRepository, identityRepository, notificationRepository, queue, 20)
	if _, err := matcher.RunOnce(ctx); err != nil {
		t.Fatalf("run matcher: %v", err)
	}

	deliveryTasks, err := queue.DequeueDeliveryTasks(ctx, 20)
	if err != nil {
		t.Fatalf("dequeue delivery tasks: %v", err)
	}
	if len(deliveryTasks) != 1 {
		t.Fatalf("expected 1 delivery task after matcher, got %d", len(deliveryTasks))
	}
	if err := queue.EnqueueDeliveryTask(ctx, deliveryTasks[0]); err != nil {
		t.Fatalf("re-enqueue delivery task: %v", err)
	}

	notifier := notificationapp.NewNotifier(nil, notificationRepository, queue, integrationEmailSender{}, 20)
	if _, err := notifier.RunOnce(ctx); err != nil {
		t.Fatalf("run notifier: %v", err)
	}

	updated, err := notificationRepository.GetByID(ctx, deliveryTasks[0].NotificationID)
	if err != nil {
		t.Fatalf("get notification by id: %v", err)
	}
	if updated.Status != notification.StatusSent {
		t.Fatalf("expected notification status sent, got %s", updated.Status)
	}

	tasks, err := queue.DequeueDeliveryTasks(ctx, 20)
	if err != nil {
		t.Fatalf("dequeue delivery tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected empty delivery queue after notifier run, got %d", len(tasks))
	}
}
