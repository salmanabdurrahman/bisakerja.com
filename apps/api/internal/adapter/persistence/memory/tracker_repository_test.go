package memory

import (
	"context"
	"errors"
	"testing"

	trackerdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/tracker"
)

func TestTrackerRepository_BookmarkCRUD(t *testing.T) {
	repo := NewTrackerRepository()
	ctx := context.Background()

	created, err := repo.CreateBookmark(ctx, trackerdomain.CreateBookmarkInput{
		UserID: "user_1",
		JobID:  "job_abc",
	})
	if err != nil {
		t.Fatalf("create bookmark: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected bookmark id")
	}

	items, err := repo.ListBookmarksByUser(ctx, "user_1")
	if err != nil {
		t.Fatalf("list bookmarks: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one bookmark, got %d", len(items))
	}

	isBookmarked, err := repo.IsBookmarked(ctx, "user_1", "job_abc")
	if err != nil {
		t.Fatalf("is bookmarked: %v", err)
	}
	if !isBookmarked {
		t.Fatal("expected job_abc to be bookmarked")
	}

	if err := repo.DeleteBookmarkByUserAndJobID(ctx, "user_1", "job_abc"); err != nil {
		t.Fatalf("delete bookmark: %v", err)
	}

	if err := repo.DeleteBookmarkByUserAndJobID(ctx, "user_1", "job_abc"); !errors.Is(err, trackerdomain.ErrBookmarkNotFound) {
		t.Fatalf("expected ErrBookmarkNotFound, got %v", err)
	}
}

func TestTrackerRepository_BookmarkDuplicate(t *testing.T) {
	repo := NewTrackerRepository()
	ctx := context.Background()

	if _, err := repo.CreateBookmark(ctx, trackerdomain.CreateBookmarkInput{
		UserID: "user_dup",
		JobID:  "job_dup",
	}); err != nil {
		t.Fatalf("first bookmark: %v", err)
	}

	_, err := repo.CreateBookmark(ctx, trackerdomain.CreateBookmarkInput{
		UserID: "user_dup",
		JobID:  "job_dup",
	})
	if !errors.Is(err, trackerdomain.ErrBookmarkAlreadyExists) {
		t.Fatalf("expected ErrBookmarkAlreadyExists, got %v", err)
	}
}

func TestTrackerRepository_ApplicationCRUD(t *testing.T) {
	repo := NewTrackerRepository()
	ctx := context.Background()

	created, err := repo.CreateTrackedApplication(ctx, trackerdomain.CreateTrackedApplicationInput{
		UserID: "user_1",
		JobID:  "job_app",
		Notes:  "applied",
	})
	if err != nil {
		t.Fatalf("create tracked application: %v", err)
	}
	if created.Status != trackerdomain.ApplicationStatusApplied {
		t.Fatalf("expected applied, got %s", created.Status)
	}

	updated, err := repo.UpdateTrackedApplicationStatus(ctx, trackerdomain.UpdateTrackedApplicationStatusInput{
		UserID:        "user_1",
		ApplicationID: created.ID,
		Status:        trackerdomain.ApplicationStatusInterview,
	})
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.Status != trackerdomain.ApplicationStatusInterview {
		t.Fatalf("expected interview, got %s", updated.Status)
	}

	items, err := repo.ListTrackedApplicationsByUser(ctx, "user_1")
	if err != nil {
		t.Fatalf("list applications: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one application, got %d", len(items))
	}

	count, err := repo.CountActiveTrackedApplicationsByUser(ctx, "user_1")
	if err != nil {
		t.Fatalf("count active: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 active application, got %d", count)
	}

	// set to rejected — should not count as active
	if _, err := repo.UpdateTrackedApplicationStatus(ctx, trackerdomain.UpdateTrackedApplicationStatusInput{
		UserID:        "user_1",
		ApplicationID: created.ID,
		Status:        trackerdomain.ApplicationStatusRejected,
	}); err != nil {
		t.Fatalf("update to rejected: %v", err)
	}
	count, err = repo.CountActiveTrackedApplicationsByUser(ctx, "user_1")
	if err != nil {
		t.Fatalf("count after reject: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 active after rejection, got %d", count)
	}

	if err := repo.DeleteTrackedApplicationByUserAndID(ctx, "user_1", created.ID); err != nil {
		t.Fatalf("delete application: %v", err)
	}
	if err := repo.DeleteTrackedApplicationByUserAndID(ctx, "user_1", created.ID); !errors.Is(err, trackerdomain.ErrApplicationNotFound) {
		t.Fatalf("expected ErrApplicationNotFound, got %v", err)
	}
}
