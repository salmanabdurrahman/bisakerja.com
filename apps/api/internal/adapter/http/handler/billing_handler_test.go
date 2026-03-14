package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

type handlerProviderStub struct {
	ensureCustomerErr error
	createInvoiceErr  error
}

func (s *handlerProviderStub) EnsureCustomer(
	_ context.Context,
	input billingdomain.EnsureCustomerInput,
) (billingdomain.Customer, error) {
	if s.ensureCustomerErr != nil {
		return billingdomain.Customer{}, s.ensureCustomerErr
	}
	return billingdomain.Customer{
		ID:    "cust_1",
		Email: input.Email,
		Name:  input.Name,
	}, nil
}

func (s *handlerProviderStub) CreateInvoice(
	_ context.Context,
	input billingdomain.CreateInvoiceInput,
) (billingdomain.Invoice, error) {
	if s.createInvoiceErr != nil {
		return billingdomain.Invoice{}, s.createInvoiceErr
	}
	expiredAt := time.Now().UTC().Add(24 * time.Hour)
	return billingdomain.Invoice{
		ID:            "inv_1",
		TransactionID: "trx_1",
		CheckoutURL:   "https://pay.example.com/checkout",
		Amount:        input.Amount,
		ExpiresAt:     &expiredAt,
	}, nil
}

func TestBillingHandler_CreateCheckoutSession_Success(t *testing.T) {
	handler, userID := setupBillingHandler(t, false, &handlerProviderStub{})

	requestBody := map[string]any{
		"plan_code":    "pro_monthly",
		"redirect_url": "https://app.bisakerja.com/billing/success",
	}
	recorder := performCheckoutRequest(t, handler, requestBody, userID, "idem-success")

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (%s)", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Data struct {
			Provider          string `json:"provider"`
			TransactionStatus string `json:"transaction_status"`
			SubscriptionState string `json:"subscription_state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Provider != "mayar" {
		t.Fatalf("expected provider mayar, got %q", payload.Data.Provider)
	}
	if payload.Data.TransactionStatus != "pending" {
		t.Fatalf("expected pending status, got %q", payload.Data.TransactionStatus)
	}
	if payload.Data.SubscriptionState != "pending_payment" {
		t.Fatalf("expected pending_payment state, got %q", payload.Data.SubscriptionState)
	}
}

func TestBillingHandler_CreateCheckoutSession_Unauthorized(t *testing.T) {
	handler, _ := setupBillingHandler(t, false, &handlerProviderStub{})

	requestBody := map[string]any{
		"plan_code":    "pro_monthly",
		"redirect_url": "https://app.bisakerja.com/billing/success",
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(requestBody); err != nil {
		t.Fatalf("encode request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout-session", &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_checkout_unauthorized"))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.CreateCheckoutSession(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestBillingHandler_CreateCheckoutSession_ErrorMatrix(t *testing.T) {
	testCases := []struct {
		name          string
		isPremiumUser bool
		provider      *handlerProviderStub
		payload       map[string]any
		repeatRequest bool
		expectedCode  int
	}{
		{
			name:         "invalid plan",
			payload:      map[string]any{"plan_code": "unknown", "redirect_url": "https://app.bisakerja.com/billing/success"},
			provider:     &handlerProviderStub{},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid redirect url",
			payload:      map[string]any{"plan_code": "pro_monthly", "redirect_url": "http://app.bisakerja.com/billing/success"},
			provider:     &handlerProviderStub{},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:          "already premium",
			isPremiumUser: true,
			payload:       map[string]any{"plan_code": "pro_monthly", "redirect_url": "https://app.bisakerja.com/billing/success"},
			provider:      &handlerProviderStub{},
			expectedCode:  http.StatusConflict,
		},
		{
			name:          "rate limited",
			repeatRequest: true,
			payload:       map[string]any{"plan_code": "pro_monthly", "redirect_url": "https://app.bisakerja.com/billing/success"},
			provider:      &handlerProviderStub{},
			expectedCode:  http.StatusTooManyRequests,
		},
		{
			name:         "mayar upstream",
			payload:      map[string]any{"plan_code": "pro_monthly", "redirect_url": "https://app.bisakerja.com/billing/success"},
			provider:     &handlerProviderStub{ensureCustomerErr: billingdomain.ErrProviderUpstream},
			expectedCode: http.StatusBadGateway,
		},
		{
			name:         "mayar rate limited",
			payload:      map[string]any{"plan_code": "pro_monthly", "redirect_url": "https://app.bisakerja.com/billing/success"},
			provider:     &handlerProviderStub{ensureCustomerErr: billingdomain.ErrProviderRateLimited},
			expectedCode: http.StatusServiceUnavailable,
		},
		{
			name:         "dependency unavailable",
			payload:      map[string]any{"plan_code": "pro_monthly", "redirect_url": "https://app.bisakerja.com/billing/success"},
			provider:     &handlerProviderStub{ensureCustomerErr: billingdomain.ErrProviderUnavailable},
			expectedCode: http.StatusServiceUnavailable,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler, userID := setupBillingHandler(t, testCase.isPremiumUser, testCase.provider)
			if testCase.repeatRequest {
				first := performCheckoutRequest(t, handler, testCase.payload, userID, "")
				if first.Code != http.StatusCreated {
					t.Fatalf("expected first request status 201, got %d (%s)", first.Code, first.Body.String())
				}
			}

			recorder := performCheckoutRequest(t, handler, testCase.payload, userID, "")
			if recorder.Code != testCase.expectedCode {
				t.Fatalf("expected status %d, got %d (%s)", testCase.expectedCode, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestBillingHandler_HandleMayarWebhook(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "webhook-handler@example.com",
		PasswordHash: "hashed-password",
		Name:         "Webhook Handler User",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	transactionRepository := memory.NewBillingRepository()
	now := time.Now().UTC()
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:             user.ID,
		Provider:           billingdomain.PaymentProviderMayar,
		PlanCode:           billingdomain.PlanCodeProMonthly,
		MayarTransactionID: "trx_handler_webhook",
		InvoiceID:          "inv_handler_webhook",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("seed transaction: %v", err)
	}

	service := billingapp.NewService(identityRepository, transactionRepository, &handlerProviderStub{}, billingapp.Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
	})
	handler := NewBillingHandler(service, "webhook-secret")

	unauthorized := performWebhookRequest(t, handler, map[string]any{
		"event": "payment.received",
		"data": map[string]any{
			"transactionId": "trx_handler_webhook",
			"customerEmail": "webhook-handler@example.com",
		},
	}, "")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected webhook unauthorized 401, got %d", unauthorized.Code)
	}

	invalidPayload := performWebhookRequest(t, handler, map[string]any{
		"event": "payment.received",
	}, "webhook-secret")
	if invalidPayload.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid payload 400, got %d", invalidPayload.Code)
	}

	userNotFound := performWebhookRequest(t, handler, map[string]any{
		"event": "payment.received",
		"data": map[string]any{
			"transactionId": "trx_user_missing",
			"customerEmail": "missing-user@example.com",
		},
	}, "webhook-secret")
	if userNotFound.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected user not found 422, got %d", userNotFound.Code)
	}

	success := performWebhookRequest(t, handler, map[string]any{
		"event": "payment.received",
		"data": map[string]any{
			"transactionId":     "trx_handler_webhook",
			"transactionStatus": "paid",
			"customerEmail":     "webhook-handler@example.com",
		},
	}, "webhook-secret")
	if success.Code != http.StatusOK {
		t.Fatalf("expected webhook success 200, got %d (%s)", success.Code, success.Body.String())
	}

	duplicate := performWebhookRequest(t, handler, map[string]any{
		"event": "payment.received",
		"data": map[string]any{
			"transactionId":     "trx_handler_webhook",
			"transactionStatus": "paid",
			"customerEmail":     "webhook-handler@example.com",
		},
	}, "webhook-secret")
	if duplicate.Code != http.StatusOK {
		t.Fatalf("expected webhook duplicate 200, got %d (%s)", duplicate.Code, duplicate.Body.String())
	}

	serviceUnavailable := performWebhookRequest(t, handler, map[string]any{
		"event": "payment.received",
		"data": map[string]any{
			"transactionId":     "trx_missing_transaction",
			"transactionStatus": "paid",
			"customerEmail":     "webhook-handler@example.com",
		},
	}, "webhook-secret")
	if serviceUnavailable.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected webhook service unavailable 503, got %d", serviceUnavailable.Code)
	}
}

func TestBillingHandler_GetBillingStatusAndTransactions(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "billing-read@example.com",
		PasswordHash: "hashed-password",
		Name:         "Billing Read User",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	transactionRepository := memory.NewBillingRepository()
	now := time.Now().UTC()
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:             user.ID,
		Provider:           billingdomain.PaymentProviderMayar,
		PlanCode:           billingdomain.PlanCodeProMonthly,
		MayarTransactionID: "trx_read_1",
		InvoiceID:          "inv_read_1",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("seed transaction: %v", err)
	}

	service := billingapp.NewService(identityRepository, transactionRepository, &handlerProviderStub{}, billingapp.Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
	})
	handler := NewBillingHandler(service, "webhook-secret")

	statusReq := httptest.NewRequest(http.MethodGet, "/api/v1/billing/status", nil)
	statusReq = statusReq.WithContext(observability.WithRequestID(statusReq.Context(), "req_billing_status"))
	statusReq = statusReq.WithContext(middleware.WithAuthUser(statusReq.Context(), middleware.AuthUser{
		UserID: user.ID,
		Role:   identity.RoleUser,
	}))
	statusResp := httptest.NewRecorder()
	handler.GetBillingStatus(statusResp, statusReq)
	if statusResp.Code != http.StatusOK {
		t.Fatalf("expected billing status 200, got %d (%s)", statusResp.Code, statusResp.Body.String())
	}

	transactionsReq := httptest.NewRequest(http.MethodGet, "/api/v1/billing/transactions?page=1&limit=1&status=pending", nil)
	transactionsReq = transactionsReq.WithContext(observability.WithRequestID(transactionsReq.Context(), "req_billing_transactions"))
	transactionsReq = transactionsReq.WithContext(middleware.WithAuthUser(transactionsReq.Context(), middleware.AuthUser{
		UserID: user.ID,
		Role:   identity.RoleUser,
	}))
	transactionsResp := httptest.NewRecorder()
	handler.GetBillingTransactions(transactionsResp, transactionsReq)
	if transactionsResp.Code != http.StatusOK {
		t.Fatalf("expected billing transactions 200, got %d (%s)", transactionsResp.Code, transactionsResp.Body.String())
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/v1/billing/transactions?status=unknown", nil)
	invalidReq = invalidReq.WithContext(observability.WithRequestID(invalidReq.Context(), "req_billing_transactions_invalid"))
	invalidReq = invalidReq.WithContext(middleware.WithAuthUser(invalidReq.Context(), middleware.AuthUser{
		UserID: user.ID,
		Role:   identity.RoleUser,
	}))
	invalidResp := httptest.NewRecorder()
	handler.GetBillingTransactions(invalidResp, invalidReq)
	if invalidResp.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid transactions query 400, got %d", invalidResp.Code)
	}
}

func setupBillingHandler(
	t *testing.T,
	isPremiumUser bool,
	provider billingdomain.Provider,
	webhookToken ...string,
) (*BillingHandler, string) {
	t.Helper()

	identityRepository := memory.NewIdentityRepository()
	premiumExpiredAt := (*time.Time)(nil)
	if isPremiumUser {
		activeUntil := time.Now().UTC().Add(48 * time.Hour)
		premiumExpiredAt = &activeUntil
	}

	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:            "billing-handler-" + time.Now().UTC().Format("20060102150405.000000000") + "@example.com",
		PasswordHash:     "hashed-password",
		Name:             "Billing Handler User",
		Role:             identity.RoleUser,
		IsPremium:        isPremiumUser,
		PremiumExpiredAt: premiumExpiredAt,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	transactionRepository := memory.NewBillingRepository()
	service := billingapp.NewService(identityRepository, transactionRepository, provider, billingapp.Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
		RateLimitWindow:   10 * time.Second,
	})

	return NewBillingHandler(service, webhookToken...), user.ID
}

func performCheckoutRequest(
	t *testing.T,
	handler *BillingHandler,
	payload map[string]any,
	userID string,
	idempotencyKey string,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		t.Fatalf("encode request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout-session", &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_checkout"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	request.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}

	recorder := httptest.NewRecorder()
	handler.CreateCheckoutSession(recorder, request)
	return recorder
}

func performWebhookRequest(
	t *testing.T,
	handler *BillingHandler,
	payload map[string]any,
	webhookToken string,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		t.Fatalf("encode request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhook/mayar", &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_webhook"))
	request.Header.Set("Content-Type", "application/json")
	if webhookToken != "" {
		request.Header.Set("X-Bisakerja-Webhook-Token", webhookToken)
	}

	recorder := httptest.NewRecorder()
	handler.HandleMayarWebhook(recorder, request)
	return recorder
}
