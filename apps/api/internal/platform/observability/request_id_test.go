package observability

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID_UsesIncomingHeaderAndContext(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, RequestIDFromContext(r.Context()))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "req_external_123")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if got := res.Header().Get("X-Request-Id"); got != "req_external_123" {
		t.Fatalf("expected response request id req_external_123, got %q", got)
	}

	if got := strings.TrimSpace(res.Body.String()); got != "req_external_123" {
		t.Fatalf("expected request id in context req_external_123, got %q", got)
	}
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, RequestIDFromContext(r.Context()))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	requestID := strings.TrimSpace(res.Body.String())
	if requestID == "" {
		t.Fatal("expected generated request id, got empty")
	}

	if len(requestID) < 20 {
		t.Fatalf("expected generated request id length >= 20, got %d", len(requestID))
	}

	if headerID := res.Header().Get("X-Request-Id"); headerID != requestID {
		t.Fatalf("expected header request id %q, got %q", requestID, headerID)
	}
}

func TestRequestIDFromContext_EmptyWhenMissing(t *testing.T) {
	if got := RequestIDFromContext(context.Background()); got != "" {
		t.Fatalf("expected empty request id from invalid context value, got %q", got)
	}
}
