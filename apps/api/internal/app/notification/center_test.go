package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	identitydomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	notificationdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

func TestCenterService_ListNotificationsAndMarkRead(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identitydomain.CreateUserInput{
		Email:        "center@example.com",
		PasswordHash: "hash",
		Name:         "Center User",
		Role:         identitydomain.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	notificationRepository := memory.NewNotificationRepository()
	first, err := notificationRepository.CreatePending(context.Background(), notificationdomain.CreateInput{
		UserID:  user.ID,
		JobID:   "job_center_1",
		Channel: notificationdomain.ChannelEmail,
	})
	if err != nil {
		t.Fatalf("create first notification: %v", err)
	}
	_, err = notificationRepository.CreatePending(context.Background(), notificationdomain.CreateInput{
		UserID:  user.ID,
		JobID:   "job_center_2",
		Channel: notificationdomain.ChannelEmail,
	})
	if err != nil {
		t.Fatalf("create second notification: %v", err)
	}

	service := NewCenterService(identityRepository, notificationRepository)
	page, err := service.ListNotifications(context.Background(), ListNotificationsInput{
		UserID: user.ID,
		Page:   1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if page.TotalRecords != 2 || len(page.Data) != 2 {
		t.Fatalf("expected two notifications, got total=%d len=%d", page.TotalRecords, len(page.Data))
	}

	updated, err := service.MarkNotificationRead(context.Background(), user.ID, first.ID)
	if err != nil {
		t.Fatalf("mark notification read: %v", err)
	}
	if updated.ReadAt == nil {
		t.Fatal("expected read_at to be set after mark read")
	}

	unreadPage, err := service.ListNotifications(context.Background(), ListNotificationsInput{
		UserID:     user.ID,
		Page:       1,
		Limit:      10,
		UnreadOnly: true,
	})
	if err != nil {
		t.Fatalf("list unread notifications: %v", err)
	}
	if unreadPage.TotalRecords != 1 || len(unreadPage.Data) != 1 {
		t.Fatalf("expected one unread notification, got total=%d len=%d", unreadPage.TotalRecords, len(unreadPage.Data))
	}
}

func TestCenterService_ListNotifications_InvalidPage(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	notificationRepository := memory.NewNotificationRepository()
	service := NewCenterService(identityRepository, notificationRepository)

	_, err := service.ListNotifications(context.Background(), ListNotificationsInput{
		UserID: "usr_x",
		Page:   -1,
		Limit:  10,
	})
	if !errors.Is(err, identitydomain.ErrUserNotFound) {
		t.Fatalf("expected user not found for invalid user first, got %v", err)
	}

	user, createErr := identityRepository.CreateUser(context.Background(), identitydomain.CreateUserInput{
		Email:        "center-page@example.com",
		PasswordHash: "hash",
		Name:         "Center Page",
		Role:         identitydomain.RoleUser,
	})
	if createErr != nil {
		t.Fatalf("create user: %v", createErr)
	}

	_, err = service.ListNotifications(context.Background(), ListNotificationsInput{
		UserID: user.ID,
		Page:   -1,
		Limit:  10,
	})
	if !errors.Is(err, ErrInvalidPage) {
		t.Fatalf("expected ErrInvalidPage, got %v", err)
	}
}
