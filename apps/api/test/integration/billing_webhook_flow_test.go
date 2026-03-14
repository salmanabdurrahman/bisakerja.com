package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestBillingWebhookFlow_DuplicateAndPremiumActivation(t *testing.T) {
	tokenManager, err := auth.NewManager("integration-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)
	authHandler := handler.NewAuthHandler(identityService)
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	billingRepository := memory.NewBillingRepository()
	billingService := billingapp.NewService(identityRepository, billingRepository, nil, billingapp.Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
	})
	billingHandler := handler.NewBillingHandler(billingService, "integration-webhook-token")

	appHandler := router.New(
		logger.New("test"),
		router.Dependencies{
			AuthHandler:    authHandler,
			BillingHandler: billingHandler,
			AuthMiddleware: authMiddleware,
		},
	)

	registerPayload := map[string]any{
		"email":    "billing-webhook-flow@example.com",
		"password": "StrongPass1",
		"name":     "Billing Webhook Flow User",
	}
	registerResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/register", registerPayload, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d (%s)", registerResponse.Code, registerResponse.Body.String())
	}

	var registerResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(registerResponse.Body.Bytes(), &registerResult); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registerResult.Data.ID == "" {
		t.Fatalf("expected register response with user id, got %s", registerResponse.Body.String())
	}

	loginPayload := map[string]any{
		"email":    "billing-webhook-flow@example.com",
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

	now := time.Now().UTC()
	_, err = billingRepository.CreatePending(httptest.NewRequest(http.MethodGet, "/", nil).Context(), billingdomain.CreatePendingTransactionInput{
		UserID:             registerResult.Data.ID,
		Provider:           billingdomain.PaymentProviderMayar,
		PlanCode:           billingdomain.PlanCodeProMonthly,
		MayarTransactionID: "trx_integration_webhook",
		InvoiceID:          "inv_integration_webhook",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("seed pending transaction: %v", err)
	}

	firstWebhook := performWebhookJSONRequest(t, appHandler, map[string]any{
		"event": "payment.received",
		"data": map[string]any{
			"transactionId":     "trx_integration_webhook",
			"transactionStatus": "paid",
			"customerEmail":     "billing-webhook-flow@example.com",
		},
	}, "integration-webhook-token")
	if firstWebhook.Code != http.StatusOK {
		t.Fatalf("expected first webhook status 200, got %d (%s)", firstWebhook.Code, firstWebhook.Body.String())
	}

	secondWebhook := performWebhookJSONRequest(t, appHandler, map[string]any{
		"event": "payment.received",
		"data": map[string]any{
			"transactionId":     "trx_integration_webhook",
			"transactionStatus": "paid",
			"customerEmail":     "billing-webhook-flow@example.com",
		},
	}, "integration-webhook-token")
	if secondWebhook.Code != http.StatusOK {
		t.Fatalf("expected duplicate webhook status 200, got %d (%s)", secondWebhook.Code, secondWebhook.Body.String())
	}

	var duplicateResult struct {
		Data struct {
			Idempotent bool `json:"idempotent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(secondWebhook.Body.Bytes(), &duplicateResult); err != nil {
		t.Fatalf("decode duplicate webhook response: %v", err)
	}
	if !duplicateResult.Data.Idempotent {
		t.Fatalf("expected duplicate webhook idempotent=true, got %s", secondWebhook.Body.String())
	}

	meResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/auth/me", nil, loginResult.Data.AccessToken)
	if meResponse.Code != http.StatusOK {
		t.Fatalf("expected /auth/me status 200, got %d (%s)", meResponse.Code, meResponse.Body.String())
	}
	var meResult struct {
		Data struct {
			IsPremium         bool   `json:"is_premium"`
			SubscriptionState string `json:"subscription_state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(meResponse.Body.Bytes(), &meResult); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if !meResult.Data.IsPremium || meResult.Data.SubscriptionState != "premium_active" {
		t.Fatalf("expected premium_active profile after webhook success, got %+v", meResult.Data)
	}
}

func performWebhookJSONRequest(
	t *testing.T,
	appHandler http.Handler,
	payload map[string]any,
	webhookToken string,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		t.Fatalf("encode webhook payload: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhook/mayar", &body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Bisakerja-Webhook-Token", webhookToken)

	response := httptest.NewRecorder()
	appHandler.ServeHTTP(response, request)
	return response
}
