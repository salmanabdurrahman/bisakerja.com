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
	aiapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/ai"
	billingapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/billing"
	growthapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/growth"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/jobs"
	notificationapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/notification"
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
	preferencesHandler := handler.NewPreferencesHandler(identityService, slog.New(slog.NewTextHandler(io.Discard, nil)))
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

	statusRequest := httptest.NewRequest(http.MethodGet, "/api/v1/billing/status", nil)
	statusResponse := httptest.NewRecorder()
	appHandler.ServeHTTP(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected billing status route to be protected with 401, got %d", statusResponse.Code)
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
		"/api/v1/webhook/midtrans",
		strings.NewReader(`{"order_id":"checkout:usr_router:key","transaction_status":"settlement","gross_amount":"49000.00","status_code":"200","signature_key":"badsig"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	appHandler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected webhook route to return 400 (registered + signature rejected), got %d", response.Code)
	}
}

func TestNew_RegistersGrowthAndNotificationRoutesWhenDependencyProvided(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokenManager, err := platformauth.NewManager("router-growth-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}

	identityRepository := memory.NewIdentityRepository()
	growthService := growthapp.NewService(identityRepository, memory.NewGrowthRepository())
	notificationService := notificationapp.NewCenterService(identityRepository, memory.NewNotificationRepository())
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	appHandler := New(logger, Dependencies{
		GrowthHandler:       handler.NewGrowthHandler(growthService),
		NotificationHandler: handler.NewNotificationHandler(notificationService),
		AuthMiddleware:      authMiddleware,
	})

	savedSearchRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/saved-searches",
		strings.NewReader(`{"query":"golang backend","frequency":"instant"}`),
	)
	savedSearchRequest.Header.Set("Content-Type", "application/json")
	savedSearchResponse := httptest.NewRecorder()
	appHandler.ServeHTTP(savedSearchResponse, savedSearchRequest)
	if savedSearchResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected saved-searches route to be protected with 401, got %d", savedSearchResponse.Code)
	}

	notificationsRequest := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	notificationsResponse := httptest.NewRecorder()
	appHandler.ServeHTTP(notificationsResponse, notificationsRequest)
	if notificationsResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected notifications route to be protected with 401, got %d", notificationsResponse.Code)
	}
}

func TestNew_RegistersAIRoutesWhenDependencyProvided(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokenManager, err := platformauth.NewManager("router-ai-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}

	authMiddleware := middleware.NewAuthenticator(tokenManager)
	appHandler := New(logger, Dependencies{
		AIHandler:      handler.NewAIHandler(&routerTestAIService{}),
		AuthMiddleware: authMiddleware,
	})

	searchRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/ai/search-assistant",
		strings.NewReader(`{"prompt":"golang backend jobs"}`),
	)
	searchRequest.Header.Set("Content-Type", "application/json")
	searchResponse := httptest.NewRecorder()
	appHandler.ServeHTTP(searchResponse, searchRequest)
	if searchResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected ai search route to be protected with 401, got %d", searchResponse.Code)
	}

	jobFitRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/ai/job-fit-summary",
		strings.NewReader(`{"job_id":"job_1"}`),
	)
	jobFitRequest.Header.Set("Content-Type", "application/json")
	jobFitResponse := httptest.NewRecorder()
	appHandler.ServeHTTP(jobFitResponse, jobFitRequest)
	if jobFitResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected ai job-fit route to be protected with 401, got %d", jobFitResponse.Code)
	}

	coverLetterRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/ai/cover-letter-draft",
		strings.NewReader(`{"job_id":"job_1","tone":"professional"}`),
	)
	coverLetterRequest.Header.Set("Content-Type", "application/json")
	coverLetterResponse := httptest.NewRecorder()
	appHandler.ServeHTTP(coverLetterResponse, coverLetterRequest)
	if coverLetterResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected ai cover-letter route to be protected with 401, got %d", coverLetterResponse.Code)
	}

	usageRequest := httptest.NewRequest(http.MethodGet, "/api/v1/ai/usage", nil)
	usageResponse := httptest.NewRecorder()
	appHandler.ServeHTTP(usageResponse, usageRequest)
	if usageResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected ai usage route to be protected with 401, got %d", usageResponse.Code)
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

type routerTestAIService struct{}

func (s *routerTestAIService) GenerateSearchAssistant(
	_ context.Context,
	_ aiapp.GenerateSearchAssistantInput,
) (aiapp.SearchAssistantResult, error) {
	return aiapp.SearchAssistantResult{}, nil
}

func (s *routerTestAIService) GenerateJobFitSummary(
	_ context.Context,
	_ aiapp.GenerateJobFitSummaryInput,
) (aiapp.JobFitSummaryResult, error) {
	return aiapp.JobFitSummaryResult{}, nil
}

func (s *routerTestAIService) GenerateCoverLetterDraft(
	_ context.Context,
	_ aiapp.GenerateCoverLetterDraftInput,
) (aiapp.CoverLetterDraftResult, error) {
	return aiapp.CoverLetterDraftResult{}, nil
}

func (s *routerTestAIService) GetUsage(
	_ context.Context,
	_ aiapp.GetUsageInput,
) (aiapp.UsageSnapshot, error) {
	return aiapp.UsageSnapshot{}, nil
}
