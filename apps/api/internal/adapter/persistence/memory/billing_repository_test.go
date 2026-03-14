package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
)

func TestBillingRepository_CreatePendingAndFindByIdempotencyKey(t *testing.T) {
	repository := NewBillingRepository()
	now := time.Now().UTC()

	created, err := repository.CreatePending(context.Background(), billing.CreatePendingTransactionInput{
		UserID:             "usr_1",
		Provider:           billing.PaymentProviderMayar,
		PlanCode:           billing.PlanCodeProMonthly,
		MayarTransactionID: "trx_1",
		InvoiceID:          "inv_1",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		IdempotencyKey:     "idem_1",
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	found, err := repository.FindPendingByUserAndIdempotencyKey(
		context.Background(),
		"usr_1",
		"idem_1",
		15*time.Minute,
		now.Add(5*time.Minute),
	)
	if err != nil {
		t.Fatalf("find pending transaction: %v", err)
	}
	if found.ID != created.ID {
		t.Fatalf("expected found ID %q, got %q", created.ID, found.ID)
	}
}

func TestBillingRepository_FindByIdempotencyKey_WindowExpired(t *testing.T) {
	repository := NewBillingRepository()
	now := time.Now().UTC()

	_, err := repository.CreatePending(context.Background(), billing.CreatePendingTransactionInput{
		UserID:             "usr_1",
		Provider:           billing.PaymentProviderMayar,
		PlanCode:           billing.PlanCodeProMonthly,
		MayarTransactionID: "trx_2",
		InvoiceID:          "inv_2",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		IdempotencyKey:     "idem_2",
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	_, err = repository.FindPendingByUserAndIdempotencyKey(
		context.Background(),
		"usr_1",
		"idem_2",
		15*time.Minute,
		now.Add(16*time.Minute),
	)
	if !errors.Is(err, billing.ErrTransactionNotFound) {
		t.Fatalf("expected ErrTransactionNotFound, got %v", err)
	}
}

func TestBillingRepository_UpdateStatusByMayarTransactionID(t *testing.T) {
	repository := NewBillingRepository()
	now := time.Now().UTC()

	_, err := repository.CreatePending(context.Background(), billing.CreatePendingTransactionInput{
		UserID:             "usr_1",
		Provider:           billing.PaymentProviderMayar,
		PlanCode:           billing.PlanCodeProMonthly,
		MayarTransactionID: "trx_update",
		InvoiceID:          "inv_update",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	updated, err := repository.UpdateStatusByMayarTransactionID(
		context.Background(),
		"trx_update",
		billing.TransactionStatusSuccess,
		map[string]any{"event": "payment.received"},
		now.Add(2*time.Minute),
	)
	if err != nil {
		t.Fatalf("update status by mayar transaction id: %v", err)
	}
	if updated.Status != billing.TransactionStatusSuccess {
		t.Fatalf("expected status success, got %s", updated.Status)
	}
	if updated.Metadata["event"] != "payment.received" {
		t.Fatalf("expected metadata to be stored, got %#v", updated.Metadata)
	}
}

func TestBillingRepository_RecordWebhookDelivery_Duplicate(t *testing.T) {
	repository := NewBillingRepository()
	now := time.Now().UTC()

	_, err := repository.RecordWebhookDelivery(context.Background(), billing.WebhookDelivery{
		Provider:         billing.PaymentProviderMayar,
		EventType:        "payment.received",
		TransactionID:    "trx_1",
		IdempotencyKey:   "mayar:payment.received:trx_1",
		ProcessingStatus: billing.WebhookProcessingStatusProcessed,
		Payload:          map[string]any{"event": "payment.received"},
		ProcessedAt:      &now,
	})
	if err != nil {
		t.Fatalf("record webhook delivery: %v", err)
	}

	_, err = repository.RecordWebhookDelivery(context.Background(), billing.WebhookDelivery{
		Provider:         billing.PaymentProviderMayar,
		EventType:        "payment.received",
		TransactionID:    "trx_1",
		IdempotencyKey:   "mayar:payment.received:trx_1",
		ProcessingStatus: billing.WebhookProcessingStatusProcessed,
		Payload:          map[string]any{"event": "payment.received"},
		ProcessedAt:      &now,
	})
	if !errors.Is(err, billing.ErrWebhookDeliveryAlreadyExists) {
		t.Fatalf("expected ErrWebhookDeliveryAlreadyExists, got %v", err)
	}
}
