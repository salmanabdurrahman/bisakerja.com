package handler

import (
	"errors"
	"net/http"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	domain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

type PreferencesHandler struct {
	service *identityapp.Service
}

type updatePreferencesRequest struct {
	Keywords  *[]string `json:"keywords"`
	Locations *[]string `json:"locations"`
	JobTypes  *[]string `json:"job_types"`
	SalaryMin *int64    `json:"salary_min"`
}

type updateNotificationPreferencesRequest struct {
	AlertMode  *string `json:"alert_mode"`
	DigestHour *int    `json:"digest_hour"`
}

func NewPreferencesHandler(service *identityapp.Service) *PreferencesHandler {
	return &PreferencesHandler{service: service}
}

func (h *PreferencesHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	preferences, err := h.service.GetPreferences(r.Context(), authUser.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to load preferences",
		}})
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Preferences retrieved", requestID, map[string]any{
		"user_id":     preferences.UserID,
		"keywords":    preferences.Keywords,
		"locations":   preferences.Locations,
		"job_types":   preferences.JobTypes,
		"salary_min":  preferences.SalaryMin,
		"alert_mode":  preferences.AlertMode,
		"digest_hour": preferences.DigestHour,
		"updated_at":  preferences.UpdatedAt,
	})
}

func (h *PreferencesHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	var request updatePreferencesRequest
	if err := decodeJSONBody(r, &request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	input := identityapp.UpdatePreferencesInput{}
	if request.Keywords != nil {
		input.Keywords = append([]string(nil), (*request.Keywords)...)
		input.KeywordsSet = true
	}
	if request.Locations != nil {
		input.Locations = append([]string(nil), (*request.Locations)...)
		input.LocationsSet = true
	}
	if request.JobTypes != nil {
		input.JobTypes = append([]string(nil), (*request.JobTypes)...)
		input.JobTypesSet = true
	}
	if request.SalaryMin != nil {
		input.SalaryMin = *request.SalaryMin
		input.SalaryMinSet = true
	}

	preferences, err := h.service.UpdatePreferences(r.Context(), authUser.UserID, input)
	if err != nil {
		switch {
		case errors.Is(err, identityapp.ErrKeywordsRequired):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "keywords",
				Code:    errcode.BadRequest,
				Message: "keywords is required",
			}})
		case errors.Is(err, identityapp.ErrInvalidKeyword):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "keywords",
				Code:    errcode.BadRequest,
				Message: "keywords must contain 1..10 items with 2..50 chars",
			}})
		case errors.Is(err, identityapp.ErrInvalidLocation):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "locations",
				Code:    errcode.BadRequest,
				Message: "locations must contain up to 5 items with 2..100 chars",
			}})
		case errors.Is(err, identityapp.ErrInvalidJobType):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "job_types",
				Code:    errcode.InvalidJobType,
				Message: "job_types must be one of fulltime, parttime, contract, internship",
			}})
		case errors.Is(err, identityapp.ErrInvalidSalaryMin):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "salary_min",
				Code:    errcode.InvalidSalaryMin,
				Message: "salary_min must be between 0 and 999000000",
			}})
		case errors.Is(err, domain.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to update preferences",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Preferences updated", requestID, map[string]any{
		"user_id":     preferences.UserID,
		"keywords":    preferences.Keywords,
		"locations":   preferences.Locations,
		"job_types":   preferences.JobTypes,
		"salary_min":  preferences.SalaryMin,
		"alert_mode":  preferences.AlertMode,
		"digest_hour": preferences.DigestHour,
		"updated_at":  preferences.UpdatedAt,
	})
}

func (h *PreferencesHandler) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	var request updateNotificationPreferencesRequest
	if err := decodeJSONBody(r, &request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	input := identityapp.UpdateNotificationPreferencesInput{}
	if request.AlertMode != nil {
		input.AlertMode = *request.AlertMode
		input.AlertModeSet = true
	}
	if request.DigestHour != nil {
		input.DigestHour = *request.DigestHour
		input.DigestHourSet = true
	}

	preferences, err := h.service.UpdateNotificationPreferences(r.Context(), authUser.UserID, input)
	if err != nil {
		switch {
		case errors.Is(err, identityapp.ErrInvalidAlertMode):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "alert_mode",
				Code:    errcode.BadRequest,
				Message: "alert_mode must be one of instant, daily_digest, weekly_digest",
			}})
		case errors.Is(err, identityapp.ErrInvalidDigestHour):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "digest_hour",
				Code:    errcode.BadRequest,
				Message: "digest_hour must be 0..23 and only set for digest mode",
			}})
		case errors.Is(err, domain.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to update notification preferences",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Notification preferences updated", requestID, map[string]any{
		"user_id":     preferences.UserID,
		"alert_mode":  preferences.AlertMode,
		"digest_hour": preferences.DigestHour,
		"updated_at":  preferences.UpdatedAt,
	})
}
