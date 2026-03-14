package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	notificationapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/notification"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

type NotificationCenterService interface {
	ListNotifications(ctx context.Context, input notificationapp.ListNotificationsInput) (notificationapp.ListNotificationsResult, error)
	MarkNotificationRead(ctx context.Context, userID string, notificationID string) (notification.Notification, error)
}

type NotificationHandler struct {
	service NotificationCenterService
}

type notificationsQuery struct {
	Page       int
	Limit      int
	UnreadOnly bool
}

func NewNotificationHandler(service NotificationCenterService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	query, err := parseNotificationsQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: err.Error(),
		}})
		return
	}

	result, serviceErr := h.service.ListNotifications(r.Context(), notificationapp.ListNotificationsInput{
		UserID:     authUser.UserID,
		Page:       query.Page,
		Limit:      query.Limit,
		UnreadOnly: query.UnreadOnly,
	})
	if serviceErr != nil {
		switch {
		case errors.Is(serviceErr, notificationapp.ErrInvalidPage):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "page",
				Code:    errcode.InvalidPage,
				Message: "page must be an integer >= 1",
			}})
		case errors.Is(serviceErr, notificationapp.ErrInvalidLimit):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "limit",
				Code:    errcode.InvalidLimit,
				Message: "limit must be between 1 and 100",
			}})
		case errors.Is(serviceErr, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to list notifications",
			}})
		}
		return
	}

	data := make([]map[string]any, 0, len(result.Data))
	for _, item := range result.Data {
		data = append(data, map[string]any{
			"id":            item.ID,
			"job_id":        item.JobID,
			"channel":       item.Channel,
			"status":        item.Status,
			"error_message": item.ErrorMessage,
			"sent_at":       item.SentAt,
			"read_at":       item.ReadAt,
			"created_at":    item.CreatedAt,
		})
	}

	response.WriteSuccessWithPagination(
		w,
		http.StatusOK,
		"Notifications retrieved",
		requestID,
		data,
		response.Pagination{
			Page:         result.Page,
			Limit:        result.Limit,
			TotalPages:   result.TotalPages,
			TotalRecords: result.TotalRecords,
		},
	)
}

func (h *NotificationHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	notificationID := strings.TrimSpace(r.PathValue("id"))
	updated, err := h.service.MarkNotificationRead(r.Context(), authUser.UserID, notificationID)
	if err != nil {
		switch {
		case errors.Is(err, notificationapp.ErrNotificationIDRequired):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "id",
				Code:    errcode.BadRequest,
				Message: "notification id is required",
			}})
		case errors.Is(err, notification.ErrNotificationNotFound):
			response.WriteError(w, http.StatusNotFound, "Not found", requestID, []response.ErrorItem{{
				Code:    errcode.NotFound,
				Message: "notification not found",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to mark notification read",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Notification marked as read", requestID, map[string]any{
		"id":            updated.ID,
		"job_id":        updated.JobID,
		"channel":       updated.Channel,
		"status":        updated.Status,
		"error_message": updated.ErrorMessage,
		"sent_at":       updated.SentAt,
		"read_at":       updated.ReadAt,
		"created_at":    updated.CreatedAt,
	})
}

func parseNotificationsQuery(r *http.Request) (notificationsQuery, error) {
	values := r.URL.Query()
	result := notificationsQuery{
		Page:  1,
		Limit: 20,
	}

	if rawPage := strings.TrimSpace(values.Get("page")); rawPage != "" {
		page, err := strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			return notificationsQuery{}, errors.New("page must be an integer >= 1")
		}
		result.Page = page
	}

	if rawLimit := strings.TrimSpace(values.Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > 100 {
			return notificationsQuery{}, errors.New("limit must be between 1 and 100")
		}
		result.Limit = limit
	}

	if rawUnreadOnly := strings.TrimSpace(values.Get("unread_only")); rawUnreadOnly != "" {
		parsed, err := strconv.ParseBool(rawUnreadOnly)
		if err != nil {
			return notificationsQuery{}, errors.New("unread_only must be a boolean")
		}
		result.UnreadOnly = parsed
	}

	return result, nil
}
