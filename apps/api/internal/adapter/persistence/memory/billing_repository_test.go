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
		UserID:                "usr_1",
		Provider:              billing.PaymentProviderMidtrans,
		PlanCode:              billing.PlanCodeProMonthly,
		ProviderTransactionID: "trx_1",
		InvoiceID:             "inv_1",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		IdempotencyKey:        "idem_1",
		ExpiresAt:             &now,
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
		UserID:                "usr_1",
		Provider:              billing.PaymentProviderMidtrans,
		PlanCode:              billing.PlanCodeProMonthly,
		ProviderTransactionID: "trx_2",
		InvoiceID:             "inv_2",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		IdempotencyKey:        "idem_2",
		ExpiresAt:             &now,
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

func TestBillingRepository_UpdateStatusByProviderTransactionID(t *testing.T) {
	repository := NewBillingRepository()
	now := time.Now().UTC()

	_, err := repository.CreatePending(context.Background(), billing.CreatePendingTransactionInput{
		UserID:                "usr_1",
		Provider:              billing.PaymentProviderMidtrans,
		PlanCode:              billing.PlanCodeProMonthly,
		ProviderTransactionID: "trx_update",
		InvoiceID:             "inv_update",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	updated, err := repository.UpdateStatusByProviderTransactionID(
		context.Background(),
		"trx_update",
		billing.TransactionStatusSuccess,
		map[string]any{"event": "payment.received"},
		now.Add(2*time.Minute),
	)
	if err != nil {
		t.Fatalf("update status by provider transaction id: %v", err)
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
		Provider:         billing.PaymentProviderMidtrans,
		EventType:        "payment.received",
		TransactionID:    "trx_1",
		IdempotencyKey:   "midtrans:payment.received:trx_1",
		ProcessingStatus: billing.WebhookProcessingStatusProcessed,
		Payload:          map[string]any{"event": "payment.received"},
		ProcessedAt:      &now,
	})
	if err != nil {
		t.Fatalf("record webhook delivery: %v", err)
	}

	_, err = repository.RecordWebhookDelivery(context.Background(), billing.WebhookDelivery{
		Provider:         billing.PaymentProviderMidtrans,
		EventType:        "payment.received",
		TransactionID:    "trx_1",
		IdempotencyKey:   "midtrans:payment.received:trx_1",
		ProcessingStatus: billing.WebhookProcessingStatusProcessed,
		Payload:          map[string]any{"event": "payment.received"},
		ProcessedAt:      &now,
	})
	if !errors.Is(err, billing.ErrWebhookDeliveryAlreadyExists) {
		t.Fatalf("expected ErrWebhookDeliveryAlreadyExists, got %v", err)
	}
}

func TestBillingRepository_ListByUserAndListAll(t *testing.T) {
	repository := NewBillingRepository()
	now := time.Now().UTC()

	_, err := repository.CreatePending(context.Background(), billing.CreatePendingTransactionInput{
		UserID:                "usr_1",
		Provider:              billing.PaymentProviderMidtrans,
		PlanCode:              billing.PlanCodeProMonthly,
		ProviderTransactionID: "trx_list_1",
		InvoiceID:             "inv_list_1",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create first transaction: %v", err)
	}
	_, err = repository.CreatePending(context.Background(), billing.CreatePendingTransactionInput{
		UserID:                "usr_2",
		Provider:              billing.PaymentProviderMidtrans,
		PlanCode:              billing.PlanCodeProMonthly,
		ProviderTransactionID: "trx_list_2",
		InvoiceID:             "inv_list_2",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create second transaction: %v", err)
	}

	userTransactions, err := repository.ListByUser(context.Background(), "usr_1")
	if err != nil {
		t.Fatalf("list by user: %v", err)
	}
	if len(userTransactions) != 1 {
		t.Fatalf("expected one transaction for usr_1, got %d", len(userTransactions))
	}

	allTransactions, err := repository.ListAll(context.Background())
	if err != nil {
		t.Fatalf("list all transactions: %v", err)
	}
	if len(allTransactions) != 2 {
		t.Fatalf("expected two transactions in list all, got %d", len(allTransactions))
	}
}
