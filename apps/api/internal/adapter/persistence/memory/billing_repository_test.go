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
