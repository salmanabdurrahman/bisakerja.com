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

func setupBillingHandler(
	t *testing.T,
	isPremiumUser bool,
	provider billingdomain.Provider,
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

	return NewBillingHandler(service), user.ID
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
