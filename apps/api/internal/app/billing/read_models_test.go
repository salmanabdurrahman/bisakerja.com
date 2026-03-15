package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

func TestService_GetBillingStatus(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "billing-status@example.com",
		PasswordHash: "hash",
		Name:         "Billing Status",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	transactionRepository := memory.NewBillingRepository()
	now := time.Now().UTC()
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:                user.ID,
		Provider:              billingdomain.PaymentProviderMidtrans,
		PlanCode:              billingdomain.PlanCodeProMonthly,
		ProviderTransactionID: "trx_status_pending",
		InvoiceID:             "inv_status_pending",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, nil, Config{})
	status, err := service.GetBillingStatus(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get billing status: %v", err)
	}
	if status.SubscriptionState != identity.SubscriptionStatePendingPayment {
		t.Fatalf("expected pending_payment state, got %s", status.SubscriptionState)
	}
	if status.LastTransactionStatus != string(billingdomain.TransactionStatusPending) {
		t.Fatalf("expected last transaction status pending, got %s", status.LastTransactionStatus)
	}
}

func TestService_ListBillingTransactions(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "billing-transactions@example.com",
		PasswordHash: "hash",
		Name:         "Billing Transactions",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	transactionRepository := memory.NewBillingRepository()
	now := time.Now().UTC()
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:                user.ID,
		Provider:              billingdomain.PaymentProviderMidtrans,
		PlanCode:              billingdomain.PlanCodeProMonthly,
		ProviderTransactionID: "trx_list_pending",
		InvoiceID:             "inv_list_pending",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}
	_, err = transactionRepository.UpdateStatusByProviderTransactionID(
		context.Background(),
		"trx_list_pending",
		billingdomain.TransactionStatusSuccess,
		map[string]any{"source": "test"},
		now.Add(1*time.Minute),
	)
	if err != nil {
		t.Fatalf("update transaction status: %v", err)
	}
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:                user.ID,
		Provider:              billingdomain.PaymentProviderMidtrans,
		PlanCode:              billingdomain.PlanCodeProMonthly,
		ProviderTransactionID: "trx_list_pending_2",
		InvoiceID:             "inv_list_pending_2",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create second pending transaction: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, nil, Config{})
	page, err := service.ListBillingTransactions(context.Background(), ListTransactionsInput{
		UserID: user.ID,
		Page:   1,
		Limit:  10,
		Status: "pending",
	})
	if err != nil {
		t.Fatalf("list billing transactions: %v", err)
	}
	if page.TotalRecords != 1 || len(page.Data) != 1 {
		t.Fatalf("expected one pending transaction, got total=%d len=%d", page.TotalRecords, len(page.Data))
	}

	_, err = service.ListBillingTransactions(context.Background(), ListTransactionsInput{
		UserID: user.ID,
		Page:   0,
		Limit:  0,
		Status: "unknown",
	})
	if !errors.Is(err, ErrInvalidTransactionStatus) {
		t.Fatalf("expected ErrInvalidTransactionStatus, got %v", err)
	}
}
