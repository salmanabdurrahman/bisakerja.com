package notification

import (
	"context"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	notificationqueue "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/queue/memory"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

func TestMatcher_RunOnce_MatchesPremiumUserAndEnqueuesDelivery(t *testing.T) {
	ctx := context.Background()
	jobsRepository := memory.NewJobsRepository()
	identityRepository := memory.NewIdentityRepository()
	notificationRepository := memory.NewNotificationRepository()
	queue := notificationqueue.NewQueue()

	expiredAt := time.Now().UTC().Add(24 * time.Hour)
	premiumUser, err := identityRepository.CreateUser(ctx, identity.CreateUserInput{
		Email:            "premium@example.com",
		PasswordHash:     "hash",
		Name:             "Premium User",
		Role:             identity.RoleUser,
		IsPremium:        true,
		PremiumExpiredAt: &expiredAt,
	})
	if err != nil {
		t.Fatalf("create premium user: %v", err)
	}
	digestUser, err := identityRepository.CreateUser(ctx, identity.CreateUserInput{
		Email:            "digest@example.com",
		PasswordHash:     "hash",
		Name:             "Digest User",
		Role:             identity.RoleUser,
		IsPremium:        true,
		PremiumExpiredAt: &expiredAt,
	})
	if err != nil {
		t.Fatalf("create digest user: %v", err)
	}
	_, err = identityRepository.CreateUser(ctx, identity.CreateUserInput{
		Email:        "free@example.com",
		PasswordHash: "hash",
		Name:         "Free User",
		Role:         identity.RoleUser,
		IsPremium:    false,
	})
	if err != nil {
		t.Fatalf("create free user: %v", err)
	}

	now := time.Now().UTC()
	_, err = identityRepository.SavePreferences(ctx, identity.Preferences{
		UserID:    premiumUser.ID,
		Keywords:  []string{"golang", "backend"},
		Locations: []string{"jakarta"},
		JobTypes:  []string{"fulltime"},
		SalaryMin: 10000000,
		UpdatedAt: &now,
	})
	if err != nil {
		t.Fatalf("save preferences: %v", err)
	}
	_, err = identityRepository.SavePreferences(ctx, identity.Preferences{
		UserID:     digestUser.ID,
		Keywords:   []string{"golang"},
		Locations:  []string{"jakarta"},
		JobTypes:   []string{"fulltime"},
		SalaryMin:  5000000,
		AlertMode:  identity.NotificationAlertModeDailyDigest,
		DigestHour: nil,
		UpdatedAt:  &now,
	})
	if err != nil {
		t.Fatalf("save digest preferences: %v", err)
	}

	salaryMin := int64(12000000)
	upserted, err := jobsRepository.UpsertMany(ctx, job.SourceGlints, []job.UpsertInput{{
		OriginalJobID: "g-101",
		Title:         "Backend Golang Engineer",
		Company:       "Bisakerja",
		Location:      "Jakarta",
		Description:   "Golang backend role",
		URL:           "https://example.com/jobs/g-101",
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

	matcher := NewMatcher(nil, jobsRepository, identityRepository, notificationRepository, queue, 10)
	summary, err := matcher.RunOnce(ctx)
	if err != nil {
		t.Fatalf("run matcher: %v", err)
	}
	if summary.ProcessedEvents != 1 || summary.MatchedUsers != 2 || summary.EnqueuedDeliveries != 1 || summary.DeferredDigest != 1 {
		t.Fatalf("unexpected matcher summary: %+v", summary)
	}

	tasks, err := queue.DequeueDeliveryTasks(ctx, 10)
	if err != nil {
		t.Fatalf("dequeue delivery tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 delivery task, got %d", len(tasks))
	}
	if tasks[0].UserID != premiumUser.ID {
		t.Fatalf("expected delivery for premium user %s, got %s", premiumUser.ID, tasks[0].UserID)
	}

	if err := queue.EnqueueJobEvent(ctx, notification.JobEvent{JobID: jobID}); err != nil {
		t.Fatalf("enqueue duplicate job event: %v", err)
	}
	duplicateSummary, err := matcher.RunOnce(ctx)
	if err != nil {
		t.Fatalf("run matcher duplicate: %v", err)
	}
	if duplicateSummary.DuplicateCount != 2 {
		t.Fatalf("expected duplicate count 2, got %+v", duplicateSummary)
	}
}
