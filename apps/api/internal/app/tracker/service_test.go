package tracker

import (
	"context"
	"errors"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	identitydomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	trackerdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/tracker"
)

func setupTrackerService(t *testing.T) (*Service, string, string) {
	t.Helper()

	identityRepository := memory.NewIdentityRepository()
	freeUser, err := identityRepository.CreateUser(context.Background(), identitydomain.CreateUserInput{
		Email:        "tracker-free@example.com",
		PasswordHash: "hash",
		Name:         "Tracker Free User",
		Role:         identitydomain.RoleUser,
	})
	if err != nil {
		t.Fatalf("create free user: %v", err)
	}

	premiumUser, err := identityRepository.CreateUser(context.Background(), identitydomain.CreateUserInput{
		Email:        "tracker-premium@example.com",
		PasswordHash: "hash",
		Name:         "Tracker Premium User",
		Role:         identitydomain.RoleUser,
		IsPremium:    true,
	})
	if err != nil {
		t.Fatalf("create premium user: %v", err)
	}

	service := NewService(identityRepository, memory.NewTrackerRepository())
	return service, freeUser.ID, premiumUser.ID
}

func TestService_BookmarkCRUD(t *testing.T) {
	service, userID, _ := setupTrackerService(t)

	created, err := service.CreateBookmark(context.Background(), CreateBookmarkInput{
		UserID: userID,
		JobID:  "job_123",
	})
	if err != nil {
		t.Fatalf("create bookmark: %v", err)
	}
	if created.JobID != "job_123" {
		t.Fatalf("expected job_id job_123, got %s", created.JobID)
	}

	items, err := service.ListBookmarks(context.Background(), userID)
	if err != nil {
		t.Fatalf("list bookmarks: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one bookmark, got %d", len(items))
	}

	isBookmarked, err := service.IsBookmarked(context.Background(), userID, "job_123")
	if err != nil {
		t.Fatalf("is bookmarked: %v", err)
	}
	if !isBookmarked {
		t.Fatal("expected job to be bookmarked")
	}

	if err := service.DeleteBookmark(context.Background(), userID, "job_123"); err != nil {
		t.Fatalf("delete bookmark: %v", err)
	}

	if err := service.DeleteBookmark(context.Background(), userID, "job_123"); !errors.Is(err, trackerdomain.ErrBookmarkNotFound) {
		t.Fatalf("expected ErrBookmarkNotFound, got %v", err)
	}
}

func TestService_BookmarkAlreadyExists(t *testing.T) {
	service, userID, _ := setupTrackerService(t)

	if _, err := service.CreateBookmark(context.Background(), CreateBookmarkInput{
		UserID: userID,
		JobID:  "job_dup",
	}); err != nil {
		t.Fatalf("create bookmark first: %v", err)
	}

	_, err := service.CreateBookmark(context.Background(), CreateBookmarkInput{
		UserID: userID,
		JobID:  "job_dup",
	})
	if !errors.Is(err, trackerdomain.ErrBookmarkAlreadyExists) {
		t.Fatalf("expected ErrBookmarkAlreadyExists, got %v", err)
	}
}

func TestService_TrackedApplicationCRUD(t *testing.T) {
	service, userID, _ := setupTrackerService(t)

	created, err := service.CreateTrackedApplication(context.Background(), CreateTrackedApplicationInput{
		UserID: userID,
		JobID:  "job_abc",
		Notes:  "Applied via website",
	})
	if err != nil {
		t.Fatalf("create tracked application: %v", err)
	}
	if created.Status != trackerdomain.ApplicationStatusApplied {
		t.Fatalf("expected applied status, got %s", created.Status)
	}

	updated, err := service.UpdateApplicationStatus(context.Background(), UpdateApplicationStatusInput{
		UserID:        userID,
		ApplicationID: created.ID,
		Status:        "interview",
	})
	if err != nil {
		t.Fatalf("update application status: %v", err)
	}
	if updated.Status != trackerdomain.ApplicationStatusInterview {
		t.Fatalf("expected interview status, got %s", updated.Status)
	}

	items, err := service.ListTrackedApplications(context.Background(), userID)
	if err != nil {
		t.Fatalf("list tracked applications: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one tracked application, got %d", len(items))
	}

	if err := service.DeleteTrackedApplication(context.Background(), userID, created.ID); err != nil {
		t.Fatalf("delete tracked application: %v", err)
	}

	if err := service.DeleteTrackedApplication(context.Background(), userID, created.ID); !errors.Is(err, trackerdomain.ErrApplicationNotFound) {
		t.Fatalf("expected ErrApplicationNotFound, got %v", err)
	}
}

func TestService_FreeTierApplicationLimit(t *testing.T) {
	service, userID, _ := setupTrackerService(t)

	for i := 0; i < trackerdomain.FreeTierApplicationLimit; i++ {
		_, err := service.CreateTrackedApplication(context.Background(), CreateTrackedApplicationInput{
			UserID: userID,
			JobID:  "job_limit_" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatalf("create application %d: %v", i, err)
		}
	}

	_, err := service.CreateTrackedApplication(context.Background(), CreateTrackedApplicationInput{
		UserID: userID,
		JobID:  "job_over_limit",
	})
	if !errors.Is(err, trackerdomain.ErrApplicationLimitExceeded) {
		t.Fatalf("expected ErrApplicationLimitExceeded, got %v", err)
	}
}

func TestService_PremiumUnlimitedApplications(t *testing.T) {
	service, _, premiumUserID := setupTrackerService(t)

	for i := 0; i <= trackerdomain.FreeTierApplicationLimit+2; i++ {
		_, err := service.CreateTrackedApplication(context.Background(), CreateTrackedApplicationInput{
			UserID: premiumUserID,
			JobID:  "job_premium_" + string(rune('0'+i)),
		})
		if err != nil {
			t.Fatalf("create premium application %d: %v", i, err)
		}
	}
}

func TestService_InvalidApplicationStatus(t *testing.T) {
	service, userID, _ := setupTrackerService(t)

	created, err := service.CreateTrackedApplication(context.Background(), CreateTrackedApplicationInput{
		UserID: userID,
		JobID:  "job_status_test",
	})
	if err != nil {
		t.Fatalf("create tracked application: %v", err)
	}

	_, err = service.UpdateApplicationStatus(context.Background(), UpdateApplicationStatusInput{
		UserID:        userID,
		ApplicationID: created.ID,
		Status:        "nonexistent_status",
	})
	if !errors.Is(err, trackerdomain.ErrInvalidApplicationStatus) {
		t.Fatalf("expected ErrInvalidApplicationStatus, got %v", err)
	}
}
