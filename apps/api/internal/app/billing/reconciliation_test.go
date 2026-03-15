package billing

import (
	"context"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

type fakeReconciliationProvider struct {
	getInvoiceFn func(context.Context, string) (billingdomain.InvoiceSnapshot, error)
}

func (f *fakeReconciliationProvider) EnsureCustomer(
	_ context.Context,
	input billingdomain.EnsureCustomerInput,
) (billingdomain.Customer, error) {
	return billingdomain.Customer{ID: "cust_reconcile", Email: input.Email, Name: input.Name}, nil
}

func (f *fakeReconciliationProvider) CreateInvoice(
	_ context.Context,
	input billingdomain.CreateInvoiceInput,
) (billingdomain.Invoice, error) {
	expiredAt := time.Now().UTC().Add(24 * time.Hour)
	return billingdomain.Invoice{
		ID:            "inv_reconcile",
		TransactionID: "trx_reconcile",
		CheckoutURL:   "https://pay.example.com/checkout",
		Amount:        input.Amount,
		ExpiresAt:     &expiredAt,
	}, nil
}

func (f *fakeReconciliationProvider) GetInvoiceByID(
	ctx context.Context,
	invoiceID string,
) (billingdomain.InvoiceSnapshot, error) {
	if f.getInvoiceFn != nil {
		return f.getInvoiceFn(ctx, invoiceID)
	}
	return billingdomain.InvoiceSnapshot{
		InvoiceID:         invoiceID,
		TransactionID:     "trx_reconcile",
		TransactionStatus: "paid",
	}, nil
}

func TestService_ReconcileWithMidtrans_UpdatesSuccessAndPremium(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "reconcile-success@example.com",
		PasswordHash: "hash",
		Name:         "Reconcile Success",
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
		ProviderTransactionID: "trx_reconcile_success",
		InvoiceID:             "inv_reconcile_success",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, &fakeReconciliationProvider{
		getInvoiceFn: func(_ context.Context, invoiceID string) (billingdomain.InvoiceSnapshot, error) {
			return billingdomain.InvoiceSnapshot{
				InvoiceID:         invoiceID,
				TransactionID:     "trx_reconcile_success",
				TransactionStatus: "paid",
			}, nil
		},
	}, Config{})

	summary, err := service.ReconcileWithMidtrans(context.Background())
	if err != nil {
		t.Fatalf("reconcile with midtrans: %v", err)
	}
	if summary.ScannedTransactions != 1 || summary.ReconciledCount != 1 {
		t.Fatalf("unexpected reconciliation summary: %+v", summary)
	}

	transaction, err := transactionRepository.GetByProviderTransactionID(context.Background(), "trx_reconcile_success")
	if err != nil {
		t.Fatalf("get transaction by provider id: %v", err)
	}
	if transaction.Status != billingdomain.TransactionStatusSuccess {
		t.Fatalf("expected success status after reconciliation, got %s", transaction.Status)
	}

	updatedUser, err := identityRepository.GetUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if !updatedUser.IsPremium {
		t.Fatal("expected user premium active after successful reconciliation")
	}
}

func TestService_ReconcileWithMidtrans_RetryableFailureAndAnomaly(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "reconcile-retry@example.com",
		PasswordHash: "hash",
		Name:         "Reconcile Retry",
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
		ProviderTransactionID: "trx_reconcile_retry",
		InvoiceID:             "inv_reconcile_retry",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}
	_, err = transactionRepository.UpdateStatusByProviderTransactionID(
		context.Background(),
		"trx_reconcile_retry",
		billingdomain.TransactionStatusPending,
		map[string]any{"seed": true},
		now.Add(-25*time.Hour),
	)
	if err != nil {
		t.Fatalf("set old updated_at for anomaly seed: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, &fakeReconciliationProvider{
		getInvoiceFn: func(_ context.Context, _ string) (billingdomain.InvoiceSnapshot, error) {
			return billingdomain.InvoiceSnapshot{}, billingdomain.ErrProviderUnavailable
		},
	}, Config{})

	summary, err := service.ReconcileWithMidtrans(context.Background())
	if err != nil {
		t.Fatalf("reconcile with midtrans: %v", err)
	}
	if summary.RetryableFailures != 1 {
		t.Fatalf("expected one retryable failure, got %+v", summary)
	}
	if summary.AnomalyCount != 1 {
		t.Fatalf("expected one anomaly, got %+v", summary)
	}
}
