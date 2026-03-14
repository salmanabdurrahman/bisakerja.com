package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/billing/mayar"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
)

func TestBillingCheckoutFlow(t *testing.T) {
	var customerCreateCalls int64
	var invoiceCreateCalls int64

	mayarServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hl/v1/customer/create":
			atomic.AddInt64(&customerCreateCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id": "cust_123",
				},
			})
		case "/hl/v1/invoice/create":
			atomic.AddInt64(&invoiceCreateCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":            "inv_123",
					"transactionId": "trx_123",
					"invoiceUrl":    "https://pay.example.com/inv_123",
					"expiredAt":     "2026-03-20T10:00:00Z",
					"amount":        49000,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mayarServer.Close()

	tokenManager, err := auth.NewManager("integration-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)
	authHandler := handler.NewAuthHandler(identityService)
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	transactionRepository := memory.NewBillingRepository()
	mayarClient := mayar.NewClient(mayar.ClientConfig{
		BaseURL:  mayarServer.URL + "/hl/v1",
		APIKey:   "test-key",
		Sleep:    func(time.Duration) {},
		RandIntn: func(int) int { return 0 },
	})
	billingService := billingapp.NewService(identityRepository, transactionRepository, mayarClient, billingapp.Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
		IdempotencyWindow: 15 * time.Minute,
		RateLimitWindow:   10 * time.Second,
	})

	appHandler := router.New(
		logger.New("test"),
		router.Dependencies{
			AuthHandler:    authHandler,
			BillingHandler: handler.NewBillingHandler(billingService),
			AuthMiddleware: authMiddleware,
		},
	)

	registerPayload := map[string]any{
		"email":    "billing-flow@example.com",
		"password": "StrongPass1",
		"name":     "Billing Flow User",
	}
	registerResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/register", registerPayload, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d (%s)", registerResponse.Code, registerResponse.Body.String())
	}

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

	checkoutPayload := map[string]any{
		"plan_code":    "pro_monthly",
		"redirect_url": "https://app.bisakerja.com/billing/success",
	}
	checkoutResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/billing/checkout-session", checkoutPayload, loginResult.Data.AccessToken)
	if checkoutResponse.Code != http.StatusCreated {
		t.Fatalf("expected checkout status 201, got %d (%s)", checkoutResponse.Code, checkoutResponse.Body.String())
	}

	var checkoutResult struct {
		Data struct {
			Provider          string `json:"provider"`
			InvoiceID         string `json:"invoice_id"`
			TransactionID     string `json:"transaction_id"`
			SubscriptionState string `json:"subscription_state"`
			TransactionStatus string `json:"transaction_status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(checkoutResponse.Body.Bytes(), &checkoutResult); err != nil {
		t.Fatalf("decode checkout response: %v", err)
	}
	if checkoutResult.Data.Provider != "mayar" {
		t.Fatalf("expected provider mayar, got %q", checkoutResult.Data.Provider)
	}
	if checkoutResult.Data.InvoiceID == "" || checkoutResult.Data.TransactionID == "" {
		t.Fatalf("expected invoice and transaction id, got %+v", checkoutResult.Data)
	}
	if checkoutResult.Data.SubscriptionState != "pending_payment" {
		t.Fatalf("expected pending_payment state, got %q", checkoutResult.Data.SubscriptionState)
	}
	if checkoutResult.Data.TransactionStatus != "pending" {
		t.Fatalf("expected pending transaction status, got %q", checkoutResult.Data.TransactionStatus)
	}

	if atomic.LoadInt64(&customerCreateCalls) != 1 {
		t.Fatalf("expected customer/create called once, got %d", atomic.LoadInt64(&customerCreateCalls))
	}
	if atomic.LoadInt64(&invoiceCreateCalls) != 1 {
		t.Fatalf("expected invoice/create called once, got %d", atomic.LoadInt64(&invoiceCreateCalls))
	}
}
