package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
)

func TestHealthRoute_RespondsOK(t *testing.T) {
	handler := router.New(logger.New("test"))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", response.Code)
	}
}
