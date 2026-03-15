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
	validateCouponErr error
	validateCoupon    billingdomain.Coupon
	lastInvoiceInput  billingdomain.CreateInvoiceInput
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
	s.lastInvoiceInput = input
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

func (s *handlerProviderStub) ValidateCoupon(
	_ context.Context,
	input billingdomain.ValidateCouponInput,
) (billingdomain.Coupon, error) {
	if s.validateCouponErr != nil {
		return billingdomain.Coupon{}, s.validateCouponErr
	}
	if s.validateCoupon.Code != "" || s.validateCoupon.DiscountAmount > 0 || s.validateCoupon.FinalAmount > 0 {
		return s.validateCoupon, nil
	}
	return billingdomain.Coupon{
		Code:           input.Code,
		DiscountAmount: 0,
		FinalAmount:    input.Amount,
	}, nil
}

func TestBillingHandler_CreateCheckoutSession_Success(t *testing.T) {
	provider := &handlerProviderStub{}
	handler, userID := setupBillingHandler(t, false, provider)

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
			PlanCode          string `json:"plan_code"`
			OriginalAmount    int64  `json:"original_amount"`
			DiscountAmount    int64  `json:"discount_amount"`
			FinalAmount       int64  `json:"final_amount"`
			Provider          string `json:"provider"`
			TransactionStatus string `json:"transaction_status"`
			SubscriptionState string `json:"subscription_state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Provider != "midtrans" {
		t.Fatalf("expected provider midtrans, got %q", payload.Data.Provider)
	}
	if payload.Data.PlanCode != "pro_monthly" {
		t.Fatalf("expected plan code pro_monthly, got %q", payload.Data.PlanCode)
	}
	if payload.Data.OriginalAmount != 49_000 || payload.Data.DiscountAmount != 0 || payload.Data.FinalAmount != 49_000 {
		t.Fatalf("unexpected checkout amount details: %+v", payload.Data)
	}
	if payload.Data.TransactionStatus != "pending" {
		t.Fatalf("expected pending status, got %q", payload.Data.TransactionStatus)
	}
	if payload.Data.SubscriptionState != "pending_payment" {
		t.Fatalf("expected pending_payment state, got %q", payload.Data.SubscriptionState)
	}
	if provider.lastInvoiceInput.CustomerMobile != "08123456789" {
		t.Fatalf("expected provider customer mobile 08123456789, got %q", provider.lastInvoiceInput.CustomerMobile)
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
			name:         "invalid coupon code",
			payload:      map[string]any{"plan_code": "pro_monthly", "coupon_code": "SAVE99", "redirect_url": "https://app.bisakerja.com/billing/success"},
			provider:     &handlerProviderStub{validateCouponErr: billingdomain.ErrCouponInvalid},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid customer mobile",
			payload:      map[string]any{"plan_code": "pro_monthly", "customer_mobile": "abc", "redirect_url": "https://app.bisakerja.com/billing/success"},
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
			name:          "reuse pending checkout during rate window",
			repeatRequest: true,
			payload:       map[string]any{"plan_code": "pro_monthly", "redirect_url": "https://app.bisakerja.com/billing/success"},
			provider:      &handlerProviderStub{},
			expectedCode:  http.StatusOK,
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

func TestBillingHandler_CreateCheckoutSession_RateLimitedAfterFailedAttempt(t *testing.T) {
	handler, userID := setupBillingHandler(t, false, &handlerProviderStub{
		ensureCustomerErr: billingdomain.ErrProviderUnavailable,
	})
	requestBody := map[string]any{
		"plan_code":    "pro_monthly",
		"redirect_url": "https://app.bisakerja.com/billing/success",
	}

	first := performCheckoutRequest(t, handler, requestBody, userID, "")
	if first.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected first request status 503, got %d (%s)", first.Code, first.Body.String())
	}

	second := performCheckoutRequest(t, handler, requestBody, userID, "")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status 429, got %d (%s)", second.Code, second.Body.String())
	}
}

func TestBillingHandler_CreateCheckoutSession_WithCoupon(t *testing.T) {
	provider := &handlerProviderStub{
		validateCoupon: billingdomain.Coupon{
			Code:           "SAVE10",
			DiscountAmount: 10_000,
			FinalAmount:    39_000,
		},
	}
	handler, userID := setupBillingHandler(t, false, provider)

	requestBody := map[string]any{
		"plan_code":    "pro_monthly",
		"coupon_code":  "save10",
		"redirect_url": "https://app.bisakerja.com/billing/success",
	}
	recorder := performCheckoutRequest(t, handler, requestBody, userID, "idem-coupon")
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (%s)", recorder.Code, recorder.Body.String())
	}
	if provider.lastInvoiceInput.Amount != 39_000 {
		t.Fatalf("expected invoice amount 39000, got %d", provider.lastInvoiceInput.Amount)
	}

	var payload struct {
		Data struct {
			CouponCode     string `json:"coupon_code"`
			DiscountAmount int64  `json:"discount_amount"`
			FinalAmount    int64  `json:"final_amount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.CouponCode != "SAVE10" {
		t.Fatalf("expected coupon_code SAVE10, got %q", payload.Data.CouponCode)
	}
	if payload.Data.DiscountAmount != 10_000 || payload.Data.FinalAmount != 39_000 {
		t.Fatalf("unexpected discount/final amount: %+v", payload.Data)
	}
}

func TestBillingHandler_HandleMidtransWebhook(t *testing.T) {
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
		UserID:                user.ID,
		Provider:              billingdomain.PaymentProviderMidtrans,
		PlanCode:              billingdomain.PlanCodeProMonthly,
		ProviderTransactionID: "checkout:" + user.ID + ":trx_handler_webhook",
		InvoiceID:             "inv_handler_webhook",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("seed transaction: %v", err)
	}

	// Handler without server key — signature check is skipped
	service := billingapp.NewService(identityRepository, transactionRepository, &handlerProviderStub{}, billingapp.Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
	})
	handlerNoKey := NewBillingHandler(service)

	// Handler with server key — signature check is enforced
	handlerWithKey := NewBillingHandler(service, "webhook-secret")

	// Bad signature: handler enforces serverKey, payload has wrong signature_key → 400
	badSig := performWebhookRequest(t, handlerWithKey, map[string]any{
		"order_id":           "checkout:" + user.ID + ":trx_handler_webhook",
		"transaction_status": "settlement",
		"gross_amount":       "49000.00",
		"status_code":        "200",
		"signature_key":      "badsig",
	}, "")
	if badSig.Code != http.StatusBadRequest {
		t.Fatalf("expected webhook bad signature 400, got %d", badSig.Code)
	}

	// Missing required fields (no order_id) → 400
	invalidPayload := performWebhookRequest(t, handlerNoKey, map[string]any{
		"transaction_status": "settlement",
	}, "")
	if invalidPayload.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid payload 400, got %d", invalidPayload.Code)
	}

	// order_id with empty userID segment → 422
	userNotFound := performWebhookRequest(t, handlerNoKey, map[string]any{
		"order_id":           "checkout::badkey",
		"transaction_status": "settlement",
		"gross_amount":       "49000.00",
		"status_code":        "200",
	}, "")
	if userNotFound.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected user not found 422, got %d", userNotFound.Code)
	}

	// Valid settlement for seeded transaction → 200
	success := performWebhookRequest(t, handlerNoKey, map[string]any{
		"order_id":           "checkout:" + user.ID + ":trx_handler_webhook",
		"transaction_status": "settlement",
		"gross_amount":       "49000.00",
		"status_code":        "200",
	}, "")
	if success.Code != http.StatusOK {
		t.Fatalf("expected webhook success 200, got %d (%s)", success.Code, success.Body.String())
	}

	// Duplicate event → 200 (idempotent)
	duplicate := performWebhookRequest(t, handlerNoKey, map[string]any{
		"order_id":           "checkout:" + user.ID + ":trx_handler_webhook",
		"transaction_status": "settlement",
		"gross_amount":       "49000.00",
		"status_code":        "200",
	}, "")
	if duplicate.Code != http.StatusOK {
		t.Fatalf("expected webhook duplicate 200, got %d (%s)", duplicate.Code, duplicate.Body.String())
	}

	// Transaction not found in DB → 422 (order_id not recognized)
	serviceUnavailable := performWebhookRequest(t, handlerNoKey, map[string]any{
		"order_id":           "pay-unknown-order",
		"transaction_status": "settlement",
		"gross_amount":       "49000.00",
		"status_code":        "200",
	}, "")
	if serviceUnavailable.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected webhook user not found 422, got %d", serviceUnavailable.Code)
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
		UserID:                user.ID,
		Provider:              billingdomain.PaymentProviderMidtrans,
		PlanCode:              billingdomain.PlanCodeProMonthly,
		ProviderTransactionID: "trx_read_1",
		InvoiceID:             "inv_read_1",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
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
	if _, exists := payload["customer_mobile"]; !exists {
		payload["customer_mobile"] = "08123456789"
	}

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
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhook/midtrans", &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_webhook"))
	request.Header.Set("Content-Type", "application/json")
	_ = webhookToken // Midtrans auth is signature-based in payload, not header

	recorder := httptest.NewRecorder()
	handler.HandleMidtransWebhook(recorder, request)
	return recorder
}
