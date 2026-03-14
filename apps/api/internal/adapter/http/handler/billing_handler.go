package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

type BillingCheckoutService interface {
	CreateCheckoutSession(
		ctx context.Context,
		input billingapp.CreateCheckoutSessionInput,
	) (billingapp.CheckoutSession, error)
}

type BillingHandler struct {
	service BillingCheckoutService
}

type createCheckoutSessionRequest struct {
	PlanCode    string `json:"plan_code"`
	RedirectURL string `json:"redirect_url"`
}

func NewBillingHandler(service BillingCheckoutService) *BillingHandler {
	return &BillingHandler{service: service}
}

func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	var request createCheckoutSessionRequest
	if err := decodeJSONBody(r, &request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	checkout, err := h.service.CreateCheckoutSession(r.Context(), billingapp.CreateCheckoutSessionInput{
		UserID:         authUser.UserID,
		PlanCode:       request.PlanCode,
		RedirectURL:    request.RedirectURL,
		IdempotencyKey: strings.TrimSpace(r.Header.Get("Idempotency-Key")),
	})
	if err != nil {
		switch {
		case errors.Is(err, billingapp.ErrInvalidPlanCode):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "plan_code",
				Code:    errcode.InvalidPlanCode,
				Message: "plan_code must be one of: pro_monthly",
			}})
		case errors.Is(err, billingapp.ErrInvalidRedirectURL):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "redirect_url",
				Code:    errcode.InvalidRedirectURL,
				Message: "redirect_url must be https and in allowlist",
			}})
		case errors.Is(err, billingapp.ErrAlreadyPremium):
			response.WriteError(w, http.StatusConflict, "Conflict", requestID, []response.ErrorItem{{
				Code:    errcode.AlreadyPremium,
				Message: "user already has active premium subscription",
			}})
		case errors.Is(err, billingapp.ErrTooManyRequests):
			response.WriteError(w, http.StatusTooManyRequests, "Too many requests", requestID, []response.ErrorItem{{
				Code:    errcode.TooManyRequests,
				Message: "checkout request rate limit exceeded",
			}})
		case errors.Is(err, billingapp.ErrMayarUpstream):
			response.WriteError(w, http.StatusBadGateway, "Bad gateway", requestID, []response.ErrorItem{{
				Code:    errcode.MayarUpstreamError,
				Message: "mayar upstream returned invalid response",
			}})
		case errors.Is(err, billingapp.ErrMayarRateLimited):
			response.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", requestID, []response.ErrorItem{{
				Code:    errcode.MayarRateLimited,
				Message: "mayar rate limit exceeded",
			}})
		case errors.Is(err, billingapp.ErrServiceUnavailable):
			response.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", requestID, []response.ErrorItem{{
				Code:    errcode.ServiceUnavailable,
				Message: "billing dependency temporarily unavailable",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to create checkout session",
			}})
		}
		return
	}

	statusCode := http.StatusCreated
	message := "Checkout session created"
	if checkout.Reused {
		statusCode = http.StatusOK
		message = "Checkout session reused"
	}

	response.WriteSuccess(w, statusCode, message, requestID, map[string]any{
		"provider":           checkout.Provider,
		"invoice_id":         checkout.InvoiceID,
		"transaction_id":     checkout.TransactionID,
		"checkout_url":       checkout.CheckoutURL,
		"expired_at":         checkout.ExpiredAt,
		"subscription_state": checkout.SubscriptionState,
		"transaction_status": checkout.TransactionStatus,
	})
}
