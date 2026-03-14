package mayar

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
)

func TestClient_CreateInvoice_RetryUntilSuccess(t *testing.T) {
	var attempts int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hl/v1/invoice/create" {
			http.NotFound(w, r)
			return
		}
		current := atomic.AddInt64(&attempts, 1)
		if current < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id":            "inv_123",
				"transactionId": "trx_123",
				"invoiceUrl":    "https://pay.example.com/inv_123",
				"expiredAt":     "2026-03-20T10:00:00Z",
				"amount":        49000,
			},
		})
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL:    server.URL + "/hl/v1",
		APIKey:     "test-key",
		MaxRetries: 3,
		Sleep:      func(time.Duration) {},
		RandIntn:   func(int) int { return 0 },
	})

	invoice, err := client.CreateInvoice(context.Background(), billingdomain.CreateInvoiceInput{
		CustomerID:  "cust_1",
		PlanCode:    billingdomain.PlanCodeProMonthly,
		Amount:      49_000,
		Description: "test plan",
		RedirectURL: "https://app.bisakerja.com/billing/success",
		ExternalID:  "checkout:usr_1:idem_1",
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if invoice.ID != "inv_123" || invoice.TransactionID != "trx_123" {
		t.Fatalf("unexpected invoice data: %+v", invoice)
	}
	if atomic.LoadInt64(&attempts) != 3 {
		t.Fatalf("expected 3 attempts, got %d", atomic.LoadInt64(&attempts))
	}
}

func TestClient_CreateInvoice_ExhaustedRateLimit(t *testing.T) {
	var attempts int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hl/v1/invoice/create" {
			http.NotFound(w, r)
			return
		}
		atomic.AddInt64(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL:    server.URL + "/hl/v1",
		APIKey:     "test-key",
		MaxRetries: 2,
		Sleep:      func(time.Duration) {},
		RandIntn:   func(int) int { return 0 },
	})

	_, err := client.CreateInvoice(context.Background(), billingdomain.CreateInvoiceInput{
		CustomerID:  "cust_1",
		PlanCode:    billingdomain.PlanCodeProMonthly,
		Amount:      49_000,
		Description: "test plan",
		RedirectURL: "https://app.bisakerja.com/billing/success",
		ExternalID:  "checkout:usr_1:idem_1",
	})
	if !errors.Is(err, billingdomain.ErrProviderRateLimited) {
		t.Fatalf("expected ErrProviderRateLimited, got %v", err)
	}
	if atomic.LoadInt64(&attempts) != 3 {
		t.Fatalf("expected 3 attempts with maxRetries=2, got %d", atomic.LoadInt64(&attempts))
	}
}

func TestClient_EnsureCustomer_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hl/v1/customer/create" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{},
		})
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL:    server.URL + "/hl/v1",
		APIKey:     "test-key",
		MaxRetries: 1,
		Sleep:      func(time.Duration) {},
	})

	_, err := client.EnsureCustomer(context.Background(), billingdomain.EnsureCustomerInput{
		Name:  "Budi",
		Email: "user@example.com",
	})
	if !errors.Is(err, billingdomain.ErrProviderUpstream) {
		t.Fatalf("expected ErrProviderUpstream, got %v", err)
	}
}

func TestClient_GetInvoiceByID_RetryAndParse(t *testing.T) {
	var attempts int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hl/v1/invoice/inv_123" {
			http.NotFound(w, r)
			return
		}
		current := atomic.AddInt64(&attempts, 1)
		if current == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id":                "inv_123",
				"transactionId":     "trx_123",
				"transactionStatus": "paid",
				"customerEmail":     "user@example.com",
				"amount":            49000,
				"updatedAt":         "2026-03-20T10:00:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL:    server.URL + "/hl/v1",
		APIKey:     "test-key",
		MaxRetries: 2,
		Sleep:      func(time.Duration) {},
		RandIntn:   func(int) int { return 0 },
	})

	snapshot, err := client.GetInvoiceByID(context.Background(), "inv_123")
	if err != nil {
		t.Fatalf("get invoice by id: %v", err)
	}
	if snapshot.InvoiceID != "inv_123" || snapshot.TransactionID != "trx_123" {
		t.Fatalf("unexpected invoice snapshot: %+v", snapshot)
	}
	if snapshot.TransactionStatus != "paid" {
		t.Fatalf("expected paid invoice status, got %s", snapshot.TransactionStatus)
	}
	if atomic.LoadInt64(&attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", atomic.LoadInt64(&attempts))
	}
}

func TestClient_GetInvoiceByID_ExhaustedRateLimit(t *testing.T) {
	var attempts int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hl/v1/invoice/inv_rate" {
			http.NotFound(w, r)
			return
		}
		atomic.AddInt64(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL:    server.URL + "/hl/v1",
		APIKey:     "test-key",
		MaxRetries: 2,
		Sleep:      func(time.Duration) {},
		RandIntn:   func(int) int { return 0 },
	})

	_, err := client.GetInvoiceByID(context.Background(), "inv_rate")
	if !errors.Is(err, billingdomain.ErrProviderRateLimited) {
		t.Fatalf("expected ErrProviderRateLimited, got %v", err)
	}
	if atomic.LoadInt64(&attempts) != 3 {
		t.Fatalf("expected 3 attempts with maxRetries=2, got %d", atomic.LoadInt64(&attempts))
	}
}
