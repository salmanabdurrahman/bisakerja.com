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

// integrationBillingProvider is an in-memory stub implementing billingdomain.Provider.
type integrationBillingProvider struct{}

func (p *integrationBillingProvider) EnsureCustomer(
	_ context.Context,
	_ billingdomain.EnsureCustomerInput,
) (billingdomain.Customer, error) {
	return billingdomain.Customer{ID: "cust_integration"}, nil
}

func (p *integrationBillingProvider) CreateInvoice(
	_ context.Context,
	input billingdomain.CreateInvoiceInput,
) (billingdomain.Invoice, error) {
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	return billingdomain.Invoice{
		ID:            "inv_integration",
		TransactionID: "trx_integration",
		CheckoutURL:   "https://pay.example.com/inv_integration",
		SnapToken:     "snap_token_integration",
		Amount:        input.Amount,
		ExpiresAt:     &expiresAt,
	}, nil
}

func (p *integrationBillingProvider) ValidateCoupon(
	_ context.Context,
	_ billingdomain.ValidateCouponInput,
) (billingdomain.Coupon, error) {
	// Return no coupon — coupons are no longer part of Midtrans flow
	return billingdomain.Coupon{}, billingdomain.ErrCouponInvalid
}

func TestBillingCheckoutFlow(t *testing.T) {
	tokenManager, err := auth.NewManager("integration-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)
	authHandler := handler.NewAuthHandler(identityService)
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	transactionRepository := memory.NewBillingRepository()
	billingService := billingapp.NewService(
		identityRepository,
		transactionRepository,
		&integrationBillingProvider{},
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
			BillingHandler: handler.NewBillingHandler(billingService, ""),
			AuthMiddleware: authMiddleware,
		},
	)

	// 1. Register
	registerPayload := map[string]any{
		"email":    "billing-flow@example.com",
		"password": "StrongPass1",
		"name":     "Billing Flow User",
	}
	registerResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/register", registerPayload, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d (%s)", registerResponse.Code, registerResponse.Body.String())
	}

	// 2. Login
	loginPayload := map[string]any{
		"email":    "billing-flow@example.com",
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

	// 3. Checkout
	checkoutPayload := map[string]any{
		"plan_code":       "pro_monthly",
		"redirect_url":    "https://app.bisakerja.com/billing/success",
		"customer_mobile": "08123456789",
	}
	checkoutResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/billing/checkout-session", checkoutPayload, loginResult.Data.AccessToken)
	if checkoutResponse.Code != http.StatusCreated {
		t.Fatalf("expected checkout status 201, got %d (%s)", checkoutResponse.Code, checkoutResponse.Body.String())
	}

	var checkoutResult struct {
		Data struct {
			Provider          string `json:"provider"`
			PlanCode          string `json:"plan_code"`
			InvoiceID         string `json:"invoice_id"`
			TransactionID     string `json:"transaction_id"`
			OriginalAmount    int64  `json:"original_amount"`
			FinalAmount       int64  `json:"final_amount"`
			SubscriptionState string `json:"subscription_state"`
			TransactionStatus string `json:"transaction_status"`
			SnapToken         string `json:"snap_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(checkoutResponse.Body.Bytes(), &checkoutResult); err != nil {
		t.Fatalf("decode checkout response: %v", err)
	}

	if checkoutResult.Data.Provider != "midtrans" {
		t.Fatalf("expected provider midtrans, got %q", checkoutResult.Data.Provider)
	}
	if checkoutResult.Data.InvoiceID == "" || checkoutResult.Data.TransactionID == "" {
		t.Fatalf("expected invoice and transaction id, got %+v", checkoutResult.Data)
	}
	if checkoutResult.Data.PlanCode != "pro_monthly" {
		t.Fatalf("expected plan code pro_monthly, got %q", checkoutResult.Data.PlanCode)
	}
	if checkoutResult.Data.OriginalAmount != 49_000 || checkoutResult.Data.FinalAmount != 49_000 {
		t.Fatalf("unexpected amount details: %+v", checkoutResult.Data)
	}
	if checkoutResult.Data.SubscriptionState != "pending_payment" {
		t.Fatalf("expected pending_payment state, got %q", checkoutResult.Data.SubscriptionState)
	}
	if checkoutResult.Data.TransactionStatus != "pending" {
		t.Fatalf("expected pending transaction status, got %q", checkoutResult.Data.TransactionStatus)
	}
}
