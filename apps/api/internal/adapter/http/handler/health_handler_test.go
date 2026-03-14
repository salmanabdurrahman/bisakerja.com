package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

func TestHealthz_ReturnsSuccessEnvelope(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req = req.WithContext(observability.WithRequestID(req.Context(), "req_test"))
	res := httptest.NewRecorder()

	Healthz(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	body := res.Body.String()
	if !strings.Contains(body, `"status":"success"`) {
		t.Fatalf("expected success envelope, got %s", body)
	}

	if !strings.Contains(body, `"request_id":"req_test"`) {
		t.Fatalf("expected request_id in envelope, got %s", body)
	}
}
