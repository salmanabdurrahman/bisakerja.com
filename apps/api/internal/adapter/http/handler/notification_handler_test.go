package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	notificationapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/notification"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	notificationdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

func setupNotificationHandler(t *testing.T) (*NotificationHandler, string, string) {
	t.Helper()

	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "notification-handler@example.com",
		PasswordHash: "hash",
		Name:         "Notification Handler User",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	notificationRepository := memory.NewNotificationRepository()
	created, err := notificationRepository.CreatePending(context.Background(), notificationdomain.CreateInput{
		UserID:  user.ID,
		JobID:   "job_notification_handler",
		Channel: notificationdomain.ChannelEmail,
	})
	if err != nil {
		t.Fatalf("seed notification: %v", err)
	}

	service := notificationapp.NewCenterService(identityRepository, notificationRepository)
	return NewNotificationHandler(service), user.ID, created.ID
}

func TestNotificationHandler_ListAndMarkRead(t *testing.T) {
	handler, userID, notificationID := setupNotificationHandler(t)

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?page=1&limit=10&unread_only=true", nil)
	listRequest = listRequest.WithContext(observability.WithRequestID(listRequest.Context(), "req_notification_list"))
	listRequest = listRequest.WithContext(middleware.WithAuthUser(listRequest.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	listResponse := httptest.NewRecorder()
	handler.ListNotifications(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected notifications list status 200, got %d (%s)", listResponse.Code, listResponse.Body.String())
	}

	markReadRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+notificationID+"/read", nil)
	markReadRequest = markReadRequest.WithContext(observability.WithRequestID(markReadRequest.Context(), "req_notification_mark_read"))
	markReadRequest = markReadRequest.WithContext(middleware.WithAuthUser(markReadRequest.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	markReadRequest.SetPathValue("id", notificationID)
	markReadResponse := httptest.NewRecorder()
	handler.MarkNotificationRead(markReadResponse, markReadRequest)
	if markReadResponse.Code != http.StatusOK {
		t.Fatalf("expected notification mark read status 200, got %d (%s)", markReadResponse.Code, markReadResponse.Body.String())
	}
}
