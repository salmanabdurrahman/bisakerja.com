package router

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/jobs"
	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
	platformauth "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

func TestWithRecovery_RecoversPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	panicHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})

	handler := observability.RequestID(withRecovery(logger, panicHandler))
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.Code)
	}

	body := response.Body.String()
	if !strings.Contains(body, "INTERNAL_SERVER_ERROR") {
		t.Fatalf("expected error code in response body, got %s", body)
	}

	if !strings.Contains(body, "request_id") {
		t.Fatalf("expected request_id in response body, got %s", body)
	}
}

func TestNew_RegistersHealthRoutes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := New(logger)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/readyz", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	if requestID := response.Header().Get("X-Request-Id"); requestID == "" {
		t.Fatal("expected X-Request-Id header to be set")
	}
}

func TestNew_RegistersJobsRoutesWhenDependencyProvided(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repository := memory.NewJobsRepository()
	_, err := repository.UpsertMany(context.Background(), job.SourceGlints, []job.UpsertInput{{
		OriginalJobID: "g-1",
		Title:         "Backend Engineer",
		Company:       "Acme",
		URL:           "https://example.com/jobs/g-1",
	}})
	if err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	jobsHandler := handler.NewJobsHandler(jobs.NewService(repository))
	appHandler := New(logger, Dependencies{JobsHandler: jobsHandler})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/jobs?source=glints", nil)
	response := httptest.NewRecorder()
	appHandler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	var payload struct {
		Meta struct {
			Pagination *struct {
				TotalRecords int `json:"total_records"`
			} `json:"pagination"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if payload.Meta.Pagination == nil || payload.Meta.Pagination.TotalRecords != 1 {
		t.Fatalf("unexpected pagination payload: %+v", payload.Meta.Pagination)
	}
}

func TestNew_RegistersAuthRoutesWhenDependencyProvided(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokenManager, err := platformauth.NewManager("router-test-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}

	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)
	authHandler := handler.NewAuthHandler(identityService)
	preferencesHandler := handler.NewPreferencesHandler(identityService)
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	appHandler := New(logger, Dependencies{
		AuthHandler:        authHandler,
		PreferencesHandler: preferencesHandler,
		AuthMiddleware:     authMiddleware,
	})

	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"user@example.com","password":"StrongPass1","name":"Budi"}`))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	appHandler.ServeHTTP(registerResp, registerReq)
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d", registerResp.Code)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meResp := httptest.NewRecorder()
	appHandler.ServeHTTP(meResp, meReq)
	if meResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected /auth/me status 401 without bearer token, got %d", meResp.Code)
	}
}

func TestNew_RegistersBillingRouteWhenDependencyProvided(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokenManager, err := platformauth.NewManager("router-billing-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}

	identityRepository := memory.NewIdentityRepository()
	billingService := billingapp.NewService(
		identityRepository,
		memory.NewBillingRepository(),
		&routerTestBillingProvider{},
		billingapp.Config{RedirectAllowlist: []string{"app.bisakerja.com"}},
	)
	billingHandler := handler.NewBillingHandler(billingService, "router-webhook-token")
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	appHandler := New(logger, Dependencies{
		BillingHandler: billingHandler,
		AuthMiddleware: authMiddleware,
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/billing/checkout-session",
		strings.NewReader(`{"plan_code":"pro_monthly","redirect_url":"https://app.bisakerja.com/billing/success"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	appHandler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected billing route to be protected with 401, got %d", response.Code)
	}
}

func TestNew_RegistersWebhookRouteWhenBillingDependencyProvided(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	identityRepository := memory.NewIdentityRepository()
	billingService := billingapp.NewService(
		identityRepository,
		memory.NewBillingRepository(),
		&routerTestBillingProvider{},
		billingapp.Config{RedirectAllowlist: []string{"app.bisakerja.com"}},
	)
	billingHandler := handler.NewBillingHandler(billingService, "router-webhook-token")

	appHandler := New(logger, Dependencies{
		BillingHandler: billingHandler,
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/webhook/mayar",
		strings.NewReader(`{"event":"payment.received","data":{"transactionId":"trx_router","customerEmail":"user@example.com"}}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	appHandler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected webhook route to return 401 (registered + token protected), got %d", response.Code)
	}
}

type routerTestBillingProvider struct{}

func (p *routerTestBillingProvider) EnsureCustomer(
	_ context.Context,
	_ billingdomain.EnsureCustomerInput,
) (billingdomain.Customer, error) {
	return billingdomain.Customer{ID: "cust_router"}, nil
}

func (p *routerTestBillingProvider) CreateInvoice(
	_ context.Context,
	input billingdomain.CreateInvoiceInput,
) (billingdomain.Invoice, error) {
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	return billingdomain.Invoice{
		ID:            "inv_router",
		TransactionID: "trx_router",
		CheckoutURL:   "https://pay.example.com/router",
		Amount:        input.Amount,
		ExpiresAt:     &expiresAt,
	}, nil
}
