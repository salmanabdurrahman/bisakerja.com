package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
)

// reconcileProvider is a stub Provider that starts with "pending" invoices and
// transitions to "paid" when reconcileReady is set.
type reconcileProvider struct {
	ready bool
}

func (p *reconcileProvider) EnsureCustomer(
	_ context.Context,
	_ billingdomain.EnsureCustomerInput,
) (billingdomain.Customer, error) {
	return billingdomain.Customer{ID: "cust_reconcile"}, nil
}

func (p *reconcileProvider) CreateInvoice(
	_ context.Context,
	input billingdomain.CreateInvoiceInput,
) (billingdomain.Invoice, error) {
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	return billingdomain.Invoice{
		ID:            "inv_reconcile_flow",
		TransactionID: "trx_reconcile_integration",
		CheckoutURL:   "https://pay.example.com/inv_reconcile_flow",
		SnapToken:     "snap_token_reconcile",
		Amount:        input.Amount,
		ExpiresAt:     &expiresAt,
	}, nil
}

func (p *reconcileProvider) ValidateCoupon(
	_ context.Context,
	_ billingdomain.ValidateCouponInput,
) (billingdomain.Coupon, error) {
	return billingdomain.Coupon{}, billingdomain.ErrCouponInvalid
}

func (p *reconcileProvider) GetInvoiceByID(
	_ context.Context,
	invoiceID string,
) (billingdomain.InvoiceSnapshot, error) {
	status := "pending"
	if p.ready {
		status = "paid"
	}
	return billingdomain.InvoiceSnapshot{
		InvoiceID:         invoiceID,
		TransactionID:     "trx_reconcile_integration",
		TransactionStatus: status,
		Amount:            49_000,
	}, nil
}

func TestBillingReadAndReconciliationFlow(t *testing.T) {
	tokenManager, err := auth.NewManager("integration-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)
	authHandler := handler.NewAuthHandler(identityService)
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	billingRepository := memory.NewBillingRepository()
	provider := &reconcileProvider{}
	billingService := billingapp.NewService(
		identityRepository,
		billingRepository,
		provider,
		billingapp.Config{
			RedirectAllowlist: []string{"app.bisakerja.com"},
			IdempotencyWindow: 15 * time.Minute,
			RateLimitWindow:   10 * time.Second,
		},
	)

	appHandler := router.New(
		logger.New("test"),
		router.Dependencies{
			AuthHandler:    authHandler,
			BillingHandler: handler.NewBillingHandler(billingService, "integration-webhook-token"),
			AuthMiddleware: authMiddleware,
		},
	)

	// 1. Register
	registerPayload := map[string]any{
		"email":    "billing-reconcile-flow@example.com",
		"password": "StrongPass1",
		"name":     "Billing Reconcile Flow User",
	}
	registerResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/register", registerPayload, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d (%s)", registerResponse.Code, registerResponse.Body.String())
	}

	// 2. Login
	loginPayload := map[string]any{
		"email":    "billing-reconcile-flow@example.com",
		"password": "StrongPass1",
	}
	loginResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/login", loginPayload, "")
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d (%s)", loginResponse.Code, loginResponse.Body.String())
	}
	var loginResult struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &loginResult); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	// 3. Checkout — creates pending transaction
	checkoutPayload := map[string]any{
		"plan_code":       "pro_monthly",
		"redirect_url":    "https://app.bisakerja.com/billing/success",
		"customer_mobile": "08123456789",
	}
	checkoutResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/billing/checkout-session", checkoutPayload, loginResult.Data.AccessToken)
	if checkoutResponse.Code != http.StatusCreated {
		t.Fatalf("expected checkout status 201, got %d (%s)", checkoutResponse.Code, checkoutResponse.Body.String())
	}

	// 4. Billing status before reconcile
	statusBefore := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/billing/status", nil, loginResult.Data.AccessToken)
	if statusBefore.Code != http.StatusOK {
		t.Fatalf("expected billing status before reconcile 200, got %d (%s)", statusBefore.Code, statusBefore.Body.String())
	}
	var statusBeforePayload struct {
		Data struct {
			SubscriptionState string `json:"subscription_state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(statusBefore.Body.Bytes(), &statusBeforePayload); err != nil {
		t.Fatalf("decode status before response: %v", err)
	}
	if statusBeforePayload.Data.SubscriptionState != "pending_payment" {
		t.Fatalf("expected pending_payment before reconciliation, got %q", statusBeforePayload.Data.SubscriptionState)
	}

	// 5. Transitions: provider now returns "paid"
	provider.ready = true

	// 6. Reconcile
	summary, err := billingService.ReconcileWithMidtrans(context.Background())
	if err != nil {
		t.Fatalf("reconcile with midtrans: %v", err)
	}
	if summary.ReconciledCount < 1 {
		t.Fatalf("expected reconciled count >= 1, got %+v", summary)
	}

	// 7. Billing status after reconcile
	statusAfter := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/billing/status", nil, loginResult.Data.AccessToken)
	if statusAfter.Code != http.StatusOK {
		t.Fatalf("expected billing status after reconcile 200, got %d (%s)", statusAfter.Code, statusAfter.Body.String())
	}
	var statusAfterPayload struct {
		Data struct {
			SubscriptionState     string `json:"subscription_state"`
			LastTransactionStatus string `json:"last_transaction_status"`
			IsPremium             bool   `json:"is_premium"`
		} `json:"data"`
	}
	if err := json.Unmarshal(statusAfter.Body.Bytes(), &statusAfterPayload); err != nil {
		t.Fatalf("decode status after response: %v", err)
	}
	if statusAfterPayload.Data.SubscriptionState != "premium_active" {
		t.Fatalf("expected premium_active after reconciliation, got %q", statusAfterPayload.Data.SubscriptionState)
	}
	if statusAfterPayload.Data.LastTransactionStatus != "success" {
		t.Fatalf("expected last transaction status success, got %q", statusAfterPayload.Data.LastTransactionStatus)
	}
	if !statusAfterPayload.Data.IsPremium {
		t.Fatal("expected user premium flag true after reconciliation")
	}

	// 8. Success transactions are visible
	transactionsAfter := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/billing/transactions?page=1&limit=10&status=success", nil, loginResult.Data.AccessToken)
	if transactionsAfter.Code != http.StatusOK {
		t.Fatalf("expected success transactions after reconcile 200, got %d (%s)", transactionsAfter.Code, transactionsAfter.Body.String())
	}
}
