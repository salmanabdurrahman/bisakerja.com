package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

// BillingCheckoutService defines behavior for billing checkout service.
type BillingCheckoutService interface {
	CreateCheckoutSession(
		ctx context.Context,
		input billingapp.CreateCheckoutSessionInput,
	) (billingapp.CheckoutSession, error)
	ProcessMidtransWebhook(
		ctx context.Context,
		input billingapp.ProcessMidtransWebhookInput,
	) (billingapp.ProcessMidtransWebhookResult, error)
	GetBillingStatus(ctx context.Context, userID string) (billingapp.BillingStatus, error)
	ListBillingTransactions(
		ctx context.Context,
		input billingapp.ListTransactionsInput,
	) (billingapp.ListTransactionsResult, error)
}

// BillingHandler represents billing handler.
type BillingHandler struct {
	service           BillingCheckoutService
	midtransServerKey string
}

type createCheckoutSessionRequest struct {
	PlanCode       string `json:"plan_code"`
	CouponCode     string `json:"coupon_code"`
	CustomerMobile string `json:"customer_mobile"`
	RedirectURL    string `json:"redirect_url"`
}

type billingTransactionsQuery struct {
	Page   int
	Limit  int
	Status string
}

// NewBillingHandler creates a new billing handler instance.
func NewBillingHandler(service BillingCheckoutService, midtransServerKey ...string) *BillingHandler {
	key := ""
	if len(midtransServerKey) > 0 {
		key = strings.TrimSpace(midtransServerKey[0])
	}
	return &BillingHandler{
		service:           service,
		midtransServerKey: key,
	}
}

// CreateCheckoutSession creates checkout session.
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
		CouponCode:     request.CouponCode,
		CustomerMobile: request.CustomerMobile,
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
				Message: "redirect_url must use an allowlisted host and https (http allowed only for localhost development)",
			}})
		case errors.Is(err, billingapp.ErrInvalidCouponCode):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "coupon_code",
				Code:    errcode.InvalidCouponCode,
				Message: "coupon_code is invalid or not applicable",
			}})
		case errors.Is(err, billingapp.ErrInvalidCustomerMobile):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "customer_mobile",
				Code:    errcode.InvalidCustomerMobile,
				Message: "customer_mobile is required and must be a valid phone number (9-15 digits)",
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
		case errors.Is(err, billingapp.ErrMidtransUpstream):
			response.WriteError(w, http.StatusBadGateway, "Bad gateway", requestID, []response.ErrorItem{{
				Code:    errcode.MidtransUpstreamError,
				Message: "midtrans upstream returned invalid response; verify MIDTRANS_SERVER_KEY and environment settings",
			}})
		case errors.Is(err, billingapp.ErrMidtransRateLimited):
			response.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", requestID, []response.ErrorItem{{
				Code:    errcode.MidtransRateLimited,
				Message: "midtrans rate limit exceeded",
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

	payload := map[string]any{
		"provider":           checkout.Provider,
		"plan_code":          checkout.PlanCode,
		"invoice_id":         checkout.InvoiceID,
		"transaction_id":     checkout.TransactionID,
		"checkout_url":       checkout.CheckoutURL,
		"original_amount":    checkout.OriginalAmount,
		"discount_amount":    checkout.DiscountAmount,
		"final_amount":       checkout.FinalAmount,
		"expired_at":         checkout.ExpiredAt,
		"subscription_state": checkout.SubscriptionState,
		"transaction_status": checkout.TransactionStatus,
		"snap_token":         checkout.SnapToken,
	}
	if checkout.CouponCode != "" {
		payload["coupon_code"] = checkout.CouponCode
	}

	response.WriteSuccess(w, statusCode, message, requestID, payload)
}

// HandleMidtransWebhook handles midtrans webhook.
func (h *BillingHandler) HandleMidtransWebhook(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())

	payload := map[string]any{}
	if err := decodeJSONBody(r, &payload); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.InvalidWebhookPayload,
			Message: "webhook payload must be valid JSON object",
		}})
		return
	}

	result, err := h.service.ProcessMidtransWebhook(r.Context(), billingapp.ProcessMidtransWebhookInput{
		Payload:   payload,
		ServerKey: h.midtransServerKey,
	})
	if err != nil {
		switch {
		case errors.Is(err, billingapp.ErrInvalidWebhookPayload):
			response.WriteError(w, http.StatusBadRequest, "Bad request", requestID, []response.ErrorItem{{
				Code:    errcode.InvalidWebhookPayload,
				Message: "webhook payload is invalid",
			}})
		case errors.Is(err, billingapp.ErrWebhookUserNotFound):
			response.WriteError(w, http.StatusUnprocessableEntity, "Unprocessable entity", requestID, []response.ErrorItem{{
				Code:    errcode.WebhookUserNotFound,
				Message: "webhook user not found",
			}})
		case errors.Is(err, billingapp.ErrMidtransUpstream):
			response.WriteError(w, http.StatusBadGateway, "Bad gateway", requestID, []response.ErrorItem{{
				Code:    errcode.MidtransUpstreamError,
				Message: "midtrans upstream returned invalid response",
			}})
		case errors.Is(err, billingapp.ErrMidtransRateLimited):
			response.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", requestID, []response.ErrorItem{{
				Code:    errcode.MidtransRateLimited,
				Message: "midtrans rate limit exceeded",
			}})
		case errors.Is(err, billingapp.ErrServiceUnavailable):
			response.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", requestID, []response.ErrorItem{{
				Code:    errcode.ServiceUnavailable,
				Message: "webhook processing dependency unavailable",
			}})
		default:
			response.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", requestID, []response.ErrorItem{{
				Code:    errcode.ServiceUnavailable,
				Message: "failed to process webhook",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Webhook processed", requestID, map[string]any{
		"provider":   result.Provider,
		"processed":  result.Processed,
		"idempotent": result.Idempotent,
	})
}

// GetBillingStatus returns billing status.
func (h *BillingHandler) GetBillingStatus(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	status, err := h.service.GetBillingStatus(r.Context(), authUser.UserID)
	if err != nil {
		switch {
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to load billing status",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Billing status retrieved", requestID, map[string]any{
		"plan_code":               status.PlanCode,
		"subscription_state":      status.SubscriptionState,
		"is_premium":              status.IsPremium,
		"premium_expired_at":      status.PremiumExpiredAt,
		"last_transaction_status": status.LastTransactionStatus,
	})
}

// GetBillingTransactions returns billing transactions.
func (h *BillingHandler) GetBillingTransactions(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	query, err := parseBillingTransactionsQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: err.Error(),
		}})
		return
	}

	result, serviceErr := h.service.ListBillingTransactions(r.Context(), billingapp.ListTransactionsInput{
		UserID: authUser.UserID,
		Page:   query.Page,
		Limit:  query.Limit,
		Status: query.Status,
	})
	if serviceErr != nil {
		switch {
		case errors.Is(serviceErr, billingapp.ErrInvalidPage):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "page",
				Code:    errcode.InvalidPage,
				Message: "page must be an integer >= 1",
			}})
		case errors.Is(serviceErr, billingapp.ErrInvalidLimit):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "limit",
				Code:    errcode.InvalidLimit,
				Message: "limit must be between 1 and 100",
			}})
		case errors.Is(serviceErr, billingapp.ErrInvalidTransactionStatus):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "status",
				Code:    errcode.BadRequest,
				Message: "status must be one of pending, reminder, success, failed",
			}})
		case errors.Is(serviceErr, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to load billing transactions",
			}})
		}
		return
	}

	transactions := make([]map[string]any, 0, len(result.Data))
	for _, item := range result.Data {
		transactions = append(transactions, map[string]any{
			"id":                      item.ID,
			"provider":                item.Provider,
			"provider_transaction_id": item.ProviderTransactionID,
			"amount":                  item.Amount,
			"status":                  item.Status,
			"created_at":              item.CreatedAt,
		})
	}

	response.WriteSuccessWithPagination(
		w,
		http.StatusOK,
		"Transactions retrieved",
		requestID,
		transactions,
		response.Pagination{
			Page:         result.Page,
			Limit:        result.Limit,
			TotalPages:   result.TotalPages,
			TotalRecords: result.TotalRecords,
		},
	)
}

func parseBillingTransactionsQuery(r *http.Request) (billingTransactionsQuery, error) {
	values := r.URL.Query()
	result := billingTransactionsQuery{
		Page:   1,
		Limit:  20,
		Status: strings.ToLower(strings.TrimSpace(values.Get("status"))),
	}

	if rawPage := strings.TrimSpace(values.Get("page")); rawPage != "" {
		page, err := strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			return billingTransactionsQuery{}, errors.New("page must be an integer >= 1")
		}
		result.Page = page
	}

	if rawLimit := strings.TrimSpace(values.Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > 100 {
			return billingTransactionsQuery{}, errors.New("limit must be between 1 and 100")
		}
		result.Limit = limit
	}

	if result.Status != "" {
		switch result.Status {
		case "pending", "reminder", "success", "failed":
		default:
			return billingTransactionsQuery{}, errors.New("status must be one of pending, reminder, success, failed")
		}
	}

	return result, nil
}
