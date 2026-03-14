package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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

func TestBillingReadAndReconciliationFlow(t *testing.T) {
	var statusMu sync.RWMutex
	invoiceStatus := "pending"

	mayarServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/hl/v1/customer/create":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id": "cust_reconcile_flow",
				},
			})
		case r.URL.Path == "/hl/v1/invoice/create":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":            "inv_reconcile_flow",
					"transactionId": "trx_reconcile_flow",
					"invoiceUrl":    "https://pay.example.com/inv_reconcile_flow",
					"expiredAt":     "2026-03-20T10:00:00Z",
					"amount":        49000,
				},
			})
		case strings.HasPrefix(r.URL.Path, "/hl/v1/invoice/"):
			statusMu.RLock()
			currentStatus := invoiceStatus
			statusMu.RUnlock()
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":                "inv_reconcile_flow",
					"transactionId":     "trx_reconcile_flow",
					"transactionStatus": currentStatus,
					"customerEmail":     "billing-reconcile-flow@example.com",
					"amount":            49000,
					"updatedAt":         "2026-03-20T10:00:00Z",
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

	billingRepository := memory.NewBillingRepository()
	mayarClient := mayar.NewClient(mayar.ClientConfig{
		BaseURL:  mayarServer.URL + "/hl/v1",
		APIKey:   "test-key",
		Sleep:    func(time.Duration) {},
		RandIntn: func(int) int { return 0 },
	})
	billingService := billingapp.NewService(identityRepository, billingRepository, mayarClient, billingapp.Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
		IdempotencyWindow: 15 * time.Minute,
		RateLimitWindow:   10 * time.Second,
	})

	appHandler := router.New(
		logger.New("test"),
		router.Dependencies{
			AuthHandler:    authHandler,
			BillingHandler: handler.NewBillingHandler(billingService, "integration-webhook-token"),
			AuthMiddleware: authMiddleware,
		},
	)

	registerPayload := map[string]any{
		"email":    "billing-reconcile-flow@example.com",
		"password": "StrongPass1",
		"name":     "Billing Reconcile Flow User",
	}
	registerResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/register", registerPayload, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d (%s)", registerResponse.Code, registerResponse.Body.String())
	}

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

	checkoutPayload := map[string]any{
		"plan_code":    "pro_monthly",
		"redirect_url": "https://app.bisakerja.com/billing/success",
	}
	checkoutResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/billing/checkout-session", checkoutPayload, loginResult.Data.AccessToken)
	if checkoutResponse.Code != http.StatusCreated {
		t.Fatalf("expected checkout status 201, got %d (%s)", checkoutResponse.Code, checkoutResponse.Body.String())
	}

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

	transactionsBefore := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/billing/transactions?page=1&limit=10&status=pending", nil, loginResult.Data.AccessToken)
	if transactionsBefore.Code != http.StatusOK {
		t.Fatalf("expected pending transactions before reconcile 200, got %d (%s)", transactionsBefore.Code, transactionsBefore.Body.String())
	}

	statusMu.Lock()
	invoiceStatus = "paid"
	statusMu.Unlock()

	summary, err := billingService.ReconcileWithMayar(context.Background())
	if err != nil {
		t.Fatalf("reconcile with mayar: %v", err)
	}
	if summary.ReconciledCount < 1 {
		t.Fatalf("expected reconciled count >= 1, got %+v", summary)
	}

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

	transactionsAfter := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/billing/transactions?page=1&limit=10&status=success", nil, loginResult.Data.AccessToken)
	if transactionsAfter.Code != http.StatusOK {
		t.Fatalf("expected success transactions after reconcile 200, got %d (%s)", transactionsAfter.Code, transactionsAfter.Body.String())
	}
}
