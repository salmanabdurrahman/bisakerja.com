package router

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
