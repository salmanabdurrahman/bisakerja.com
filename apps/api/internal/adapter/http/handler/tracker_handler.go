package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	trackerapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/tracker"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	trackerdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/tracker"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

// TrackerService defines behavior for tracker service.
type TrackerService interface {
	CreateBookmark(ctx context.Context, input trackerapp.CreateBookmarkInput) (trackerdomain.Bookmark, error)
	DeleteBookmark(ctx context.Context, userID, jobID string) error
	ListBookmarks(ctx context.Context, userID string) ([]trackerdomain.Bookmark, error)
	CreateTrackedApplication(ctx context.Context, input trackerapp.CreateTrackedApplicationInput) (trackerdomain.TrackedApplication, error)
	UpdateApplicationStatus(ctx context.Context, input trackerapp.UpdateApplicationStatusInput) (trackerdomain.TrackedApplication, error)
	DeleteTrackedApplication(ctx context.Context, userID, applicationID string) error
	ListTrackedApplications(ctx context.Context, userID string) ([]trackerdomain.TrackedApplication, error)
}

// TrackerHandler represents tracker handler.
type TrackerHandler struct {
	service TrackerService
}

// NewTrackerHandler creates a new tracker handler instance.
func NewTrackerHandler(service TrackerService) *TrackerHandler {
	return &TrackerHandler{service: service}
}

type createBookmarkRequest struct {
	JobID string `json:"job_id"`
}

type createTrackedApplicationRequest struct {
	JobID string `json:"job_id"`
	Notes string `json:"notes"`
}

type updateApplicationStatusRequest struct {
	Status string `json:"status"`
}

// CreateBookmark creates a bookmark for a job.
func (h *TrackerHandler) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	var req createBookmarkRequest
	if err := decodeJSONBody(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	created, err := h.service.CreateBookmark(r.Context(), trackerapp.CreateBookmarkInput{
		UserID: authUser.UserID,
		JobID:  req.JobID,
	})
	if err != nil {
		switch {
		case errors.Is(err, trackerapp.ErrInvalidJobID):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "job_id",
				Code:    errcode.BadRequest,
				Message: "job_id must be 1..100 characters",
			}})
		case errors.Is(err, trackerdomain.ErrBookmarkAlreadyExists):
			response.WriteError(w, http.StatusConflict, "Conflict", requestID, []response.ErrorItem{{
				Code:    errcode.BadRequest,
				Message: "job already bookmarked",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to create bookmark",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, "Bookmark created", requestID, mapBookmark(created))
}

// DeleteBookmark removes a bookmark.
func (h *TrackerHandler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	jobID := strings.TrimSpace(r.PathValue("job_id"))
	if jobID == "" {
		response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
			Field:   "job_id",
			Code:    errcode.BadRequest,
			Message: "job_id is required",
		}})
		return
	}

	if err := h.service.DeleteBookmark(r.Context(), authUser.UserID, jobID); err != nil {
		switch {
		case errors.Is(err, trackerdomain.ErrBookmarkNotFound):
			response.WriteError(w, http.StatusNotFound, "Not found", requestID, []response.ErrorItem{{
				Code:    errcode.NotFound,
				Message: "bookmark not found",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to delete bookmark",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Bookmark deleted", requestID, map[string]any{
		"job_id": jobID,
	})
}

// ListBookmarks returns bookmarks for the authenticated user.
func (h *TrackerHandler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	items, err := h.service.ListBookmarks(r.Context(), authUser.UserID)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to list bookmarks",
		}})
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, mapBookmark(item))
	}
	response.WriteSuccess(w, http.StatusOK, "Bookmarks retrieved", requestID, result)
}

// CreateTrackedApplication creates a tracked application.
func (h *TrackerHandler) CreateTrackedApplication(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	var req createTrackedApplicationRequest
	if err := decodeJSONBody(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	created, err := h.service.CreateTrackedApplication(r.Context(), trackerapp.CreateTrackedApplicationInput{
		UserID: authUser.UserID,
		JobID:  req.JobID,
		Notes:  req.Notes,
	})
	if err != nil {
		switch {
		case errors.Is(err, trackerapp.ErrInvalidJobID):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "job_id",
				Code:    errcode.BadRequest,
				Message: "job_id must be 1..100 characters",
			}})
		case errors.Is(err, trackerapp.ErrInvalidNotes):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "notes",
				Code:    errcode.BadRequest,
				Message: "notes must be <= 2000 characters",
			}})
		case errors.Is(err, trackerdomain.ErrApplicationLimitExceeded):
			response.WriteError(w, http.StatusForbidden, "Limit exceeded", requestID, []response.ErrorItem{{
				Code:    errcode.TrackerLimitExceeded,
				Message: "free tier limit of 5 active tracked applications reached",
			}})
		case errors.Is(err, trackerdomain.ErrApplicationAlreadyExists):
			response.WriteError(w, http.StatusConflict, "Conflict", requestID, []response.ErrorItem{{
				Code:    errcode.BadRequest,
				Message: "application for this job already tracked",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to create tracked application",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, "Tracked application created", requestID, mapTrackedApplication(created))
}

// UpdateApplicationStatus updates the status of a tracked application.
func (h *TrackerHandler) UpdateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	applicationID := strings.TrimSpace(r.PathValue("id"))
	if applicationID == "" {
		response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
			Field:   "id",
			Code:    errcode.BadRequest,
			Message: "application id is required",
		}})
		return
	}

	var req updateApplicationStatusRequest
	if err := decodeJSONBody(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	updated, err := h.service.UpdateApplicationStatus(r.Context(), trackerapp.UpdateApplicationStatusInput{
		UserID:        authUser.UserID,
		ApplicationID: applicationID,
		Status:        req.Status,
	})
	if err != nil {
		switch {
		case errors.Is(err, trackerdomain.ErrInvalidApplicationStatus):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "status",
				Code:    errcode.BadRequest,
				Message: "status must be one of applied, interview, offer, rejected, withdrawn",
			}})
		case errors.Is(err, trackerdomain.ErrApplicationNotFound):
			response.WriteError(w, http.StatusNotFound, "Not found", requestID, []response.ErrorItem{{
				Code:    errcode.NotFound,
				Message: "tracked application not found",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to update application status",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Application status updated", requestID, mapTrackedApplication(updated))
}

// DeleteTrackedApplication removes a tracked application.
func (h *TrackerHandler) DeleteTrackedApplication(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	applicationID := strings.TrimSpace(r.PathValue("id"))
	if applicationID == "" {
		response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
			Field:   "id",
			Code:    errcode.BadRequest,
			Message: "application id is required",
		}})
		return
	}

	if err := h.service.DeleteTrackedApplication(r.Context(), authUser.UserID, applicationID); err != nil {
		switch {
		case errors.Is(err, trackerdomain.ErrApplicationNotFound):
			response.WriteError(w, http.StatusNotFound, "Not found", requestID, []response.ErrorItem{{
				Code:    errcode.NotFound,
				Message: "tracked application not found",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to delete tracked application",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Tracked application deleted", requestID, map[string]any{
		"id": applicationID,
	})
}

// ListTrackedApplications returns tracked applications for the authenticated user.
func (h *TrackerHandler) ListTrackedApplications(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	items, err := h.service.ListTrackedApplications(r.Context(), authUser.UserID)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to list tracked applications",
		}})
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, mapTrackedApplication(item))
	}
	response.WriteSuccess(w, http.StatusOK, "Tracked applications retrieved", requestID, result)
}

func mapBookmark(item trackerdomain.Bookmark) map[string]any {
	return map[string]any{
		"id":         item.ID,
		"job_id":     item.JobID,
		"created_at": item.CreatedAt,
	}
}

func mapTrackedApplication(item trackerdomain.TrackedApplication) map[string]any {
	return map[string]any{
		"id":         item.ID,
		"job_id":     item.JobID,
		"status":     item.Status,
		"notes":      item.Notes,
		"created_at": item.CreatedAt,
		"updated_at": item.UpdatedAt,
	}
}
