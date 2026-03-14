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
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/jobs"
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
