package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	aiapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/ai"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

// AIService defines behavior for AI service.
type AIService interface {
	GenerateSearchAssistant(ctx context.Context, input aiapp.GenerateSearchAssistantInput) (aiapp.SearchAssistantResult, error)
	GetUsage(ctx context.Context, input aiapp.GetUsageInput) (aiapp.UsageSnapshot, error)
}

// AIHandler represents AI handler.
type AIHandler struct {
	service AIService
}

type aiSearchAssistantRequest struct {
	Prompt  string                   `json:"prompt"`
	Context aiSearchAssistantContext `json:"context"`
}

type aiSearchAssistantContext struct {
	Location  string   `json:"location"`
	JobTypes  []string `json:"job_types"`
	SalaryMin *int64   `json:"salary_min"`
}

// NewAIHandler creates a new AI handler instance.
func NewAIHandler(service AIService) *AIHandler {
	return &AIHandler{service: service}
}

// GenerateSearchAssistant generates AI search assistant suggestions.
func (h *AIHandler) GenerateSearchAssistant(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	var request aiSearchAssistantRequest
	if err := decodeJSONBody(r, &request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	result, err := h.service.GenerateSearchAssistant(r.Context(), aiapp.GenerateSearchAssistantInput{
		UserID: authUser.UserID,
		Prompt: request.Prompt,
		Context: aiapp.SearchAssistantContext{
			Location:  request.Context.Location,
			JobTypes:  request.Context.JobTypes,
			SalaryMin: request.Context.SalaryMin,
		},
	})
	if err != nil {
		h.writeServiceError(w, requestID, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, "AI search assistant generated", requestID, map[string]any{
		"feature":         result.Feature,
		"prompt":          result.Prompt,
		"suggested_query": result.SuggestedQuery,
		"suggested_filters": map[string]any{
			"locations":  result.SuggestedLocations,
			"job_types":  result.SuggestedJobTypes,
			"salary_min": result.SuggestedSalaryMin,
		},
		"summary":         result.Summary,
		"tier":            result.Tier,
		"provider":        result.Provider,
		"model":           result.Model,
		"daily_quota":     result.Quota.DailyQuota,
		"used_today":      result.Quota.Used,
		"quota_remaining": result.Quota.Remaining,
		"reset_at":        result.Quota.ResetAt,
	})
}

// GetUsage returns AI usage state for authenticated user.
func (h *AIHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	feature := strings.TrimSpace(r.URL.Query().Get("feature"))
	usage, err := h.service.GetUsage(r.Context(), aiapp.GetUsageInput{
		UserID:  authUser.UserID,
		Feature: feature,
	})
	if err != nil {
		h.writeServiceError(w, requestID, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, "AI usage retrieved", requestID, map[string]any{
		"feature":     usage.Feature,
		"tier":        usage.Tier,
		"daily_quota": usage.Quota.DailyQuota,
		"used":        usage.Quota.Used,
		"remaining":   usage.Quota.Remaining,
		"reset_at":    usage.Quota.ResetAt,
	})
}

func (h *AIHandler) writeServiceError(w http.ResponseWriter, requestID string, err error) {
	switch {
	case errors.Is(err, aiapp.ErrPromptRequired),
		errors.Is(err, aiapp.ErrPromptTooShort),
		errors.Is(err, aiapp.ErrPromptTooLong):
		response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
			Field:   "prompt",
			Code:    errcode.InvalidAIPrompt,
			Message: promptValidationMessage(err),
		}})
	case errors.Is(err, aiapp.ErrInvalidFeature):
		response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
			Field:   "feature",
			Code:    errcode.InvalidAIFeature,
			Message: "feature must be search_assistant",
		}})
	case errors.Is(err, aiapp.ErrQuotaExceeded):
		response.WriteError(w, http.StatusTooManyRequests, "Quota exceeded", requestID, []response.ErrorItem{{
			Code:    errcode.AIQuotaExceeded,
			Message: "ai quota exceeded for current period",
		}})
	case errors.Is(err, aiapp.ErrProviderRateLimited):
		response.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", requestID, []response.ErrorItem{{
			Code:    errcode.AIProviderRateLimited,
			Message: "ai provider rate limit exceeded",
		}})
	case errors.Is(err, aiapp.ErrProviderUpstream):
		response.WriteError(w, http.StatusBadGateway, "Bad gateway", requestID, []response.ErrorItem{{
			Code:    errcode.AIProviderUpstreamError,
			Message: "ai provider returned invalid response",
		}})
	case errors.Is(err, aiapp.ErrServiceUnavailable):
		response.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", requestID, []response.ErrorItem{{
			Code:    errcode.AIProviderUnavailable,
			Message: "ai provider unavailable",
		}})
	case errors.Is(err, identity.ErrUserNotFound):
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "user not found",
		}})
	default:
		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to process ai request",
		}})
	}
}

func promptValidationMessage(err error) string {
	switch {
	case errors.Is(err, aiapp.ErrPromptRequired):
		return "prompt is required"
	case errors.Is(err, aiapp.ErrPromptTooShort):
		return "prompt must be at least 5 characters"
	case errors.Is(err, aiapp.ErrPromptTooLong):
		return "prompt must be <= 500 characters"
	default:
		return "prompt is invalid"
	}
}
