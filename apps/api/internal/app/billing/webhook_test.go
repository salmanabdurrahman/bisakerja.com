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

// orderID builds a test order_id string used as ProviderTransactionID.
func orderID(_, key string) string {
	return "pay-test-" + key
}

func TestService_ProcessMidtransWebhook_PaymentSettledAndDuplicate(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "webhook-success@example.com",
		PasswordHash: "hash",
		Name:         "Webhook Success",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	transactionRepository := memory.NewBillingRepository()
	now := time.Now().UTC()
	oid := orderID(user.ID, "trx_webhook_success")
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:                user.ID,
		Provider:              billingdomain.PaymentProviderMidtrans,
		PlanCode:              billingdomain.PlanCodeProMonthly,
		ProviderTransactionID: oid,
		InvoiceID:             "inv_webhook_success",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, nil, Config{})
	fixedNow := time.Date(2026, time.March, 14, 0, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }

	result, err := service.ProcessMidtransWebhook(context.Background(), ProcessMidtransWebhookInput{
		Payload: map[string]any{
			"order_id":           oid,
			"transaction_status": "settlement",
			"fraud_status":       "accept",
			"gross_amount":       "49000.00",
			"status_code":        "200",
			"signature_key":      "",
		},
		ServerKey: "",
	})
	if err != nil {
		t.Fatalf("process webhook: %v", err)
	}
	if !result.Processed || result.Idempotent {
		t.Fatalf("unexpected result: %+v", result)
	}

	updatedTransaction, err := transactionRepository.GetByProviderTransactionID(context.Background(), oid)
	if err != nil {
		t.Fatalf("get transaction by provider id: %v", err)
	}
	if updatedTransaction.Status != billingdomain.TransactionStatusSuccess {
		t.Fatalf("expected transaction status success, got %s", updatedTransaction.Status)
	}

	updatedUser, err := identityRepository.GetUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if !updatedUser.IsPremium {
		t.Fatal("expected user premium to be active")
	}
	expectedExpiry := fixedNow.Add(30 * 24 * time.Hour)
	if updatedUser.PremiumExpiredAt == nil || !updatedUser.PremiumExpiredAt.Equal(expectedExpiry) {
		t.Fatalf("expected premium_expired_at %v, got %v", expectedExpiry, updatedUser.PremiumExpiredAt)
	}

	// Duplicate — should be idempotent.
	duplicate, err := service.ProcessMidtransWebhook(context.Background(), ProcessMidtransWebhookInput{
		Payload: map[string]any{
			"order_id":           oid,
			"transaction_status": "settlement",
			"fraud_status":       "accept",
			"gross_amount":       "49000.00",
			"status_code":        "200",
			"signature_key":      "",
		},
		ServerKey: "",
	})
	if err != nil {
		t.Fatalf("process duplicate webhook: %v", err)
	}
	if !duplicate.Idempotent {
		t.Fatalf("expected duplicate webhook idempotent=true, got %+v", duplicate)
	}
}

func TestService_ProcessMidtransWebhook_PendingReminder(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "webhook-reminder@example.com",
		PasswordHash: "hash",
		Name:         "Webhook Reminder",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	transactionRepository := memory.NewBillingRepository()
	now := time.Now().UTC()
	oid := orderID(user.ID, "trx_webhook_reminder")
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:                user.ID,
		Provider:              billingdomain.PaymentProviderMidtrans,
		PlanCode:              billingdomain.PlanCodeProMonthly,
		ProviderTransactionID: oid,
		InvoiceID:             "inv_webhook_reminder",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, nil, Config{})
	_, err = service.ProcessMidtransWebhook(context.Background(), ProcessMidtransWebhookInput{
		Payload: map[string]any{
			"order_id":           oid,
			"transaction_status": "pending",
			"gross_amount":       "49000.00",
			"status_code":        "201",
			"signature_key":      "",
		},
		ServerKey: "",
	})
	if err != nil {
		t.Fatalf("process webhook reminder: %v", err)
	}

	updatedTransaction, err := transactionRepository.GetByProviderTransactionID(context.Background(), oid)
	if err != nil {
		t.Fatalf("get transaction by provider id: %v", err)
	}
	if updatedTransaction.Status != billingdomain.TransactionStatusPending {
		t.Fatalf("expected transaction status pending, got %s", updatedTransaction.Status)
	}

	updatedUser, err := identityRepository.GetUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if updatedUser.IsPremium {
		t.Fatal("expected reminder webhook not activating premium")
	}
}

func TestService_ProcessMidtransWebhook_ErrorCases(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	transactionRepository := memory.NewBillingRepository()
	service := NewService(identityRepository, transactionRepository, nil, Config{})

	// Missing order_id → ErrInvalidWebhookPayload
	_, err := service.ProcessMidtransWebhook(context.Background(), ProcessMidtransWebhookInput{
		Payload: map[string]any{
			"transaction_status": "settlement",
		},
		ServerKey: "",
	})
	if !errors.Is(err, ErrInvalidWebhookPayload) {
		t.Fatalf("expected ErrInvalidWebhookPayload, got %v", err)
	}

	// Transaction exists but user doesn't → ErrWebhookUserNotFound
	now := time.Now().UTC()
	oid := orderID("usr_missing", "trx_user_missing")
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:                "usr_missing",
		Provider:              billingdomain.PaymentProviderMidtrans,
		PlanCode:              billingdomain.PlanCodeProMonthly,
		ProviderTransactionID: oid,
		InvoiceID:             "inv_user_missing",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	_, err = service.ProcessMidtransWebhook(context.Background(), ProcessMidtransWebhookInput{
		Payload: map[string]any{
			"order_id":           oid,
			"transaction_status": "settlement",
			"gross_amount":       "49000.00",
			"status_code":        "200",
			"signature_key":      "",
		},
		ServerKey: "",
	})
	if !errors.Is(err, ErrWebhookUserNotFound) {
		t.Fatalf("expected ErrWebhookUserNotFound, got %v", err)
	}

	delivery, err := transactionRepository.GetWebhookDeliveryByIdempotencyKey(
		context.Background(),
		"midtrans:"+oid+":settlement",
	)
	if err != nil {
		t.Fatalf("get webhook delivery by idempotency key: %v", err)
	}
	if delivery.ProcessingStatus != billingdomain.WebhookProcessingStatusRejected {
		t.Fatalf("expected rejected webhook delivery, got %s", delivery.ProcessingStatus)
	}
}

func TestService_ProcessMidtransWebhook_UnsupportedStatusRecorded(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "webhook-ignored@example.com",
		PasswordHash: "hash",
		Name:         "Webhook Ignored",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	transactionRepository := memory.NewBillingRepository()
	now := time.Now().UTC()
	oid := orderID(user.ID, "trx_ignored")
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:                user.ID,
		Provider:              billingdomain.PaymentProviderMidtrans,
		PlanCode:              billingdomain.PlanCodeProMonthly,
		ProviderTransactionID: oid,
		InvoiceID:             "inv_ignored",
		CheckoutURL:           "https://pay.example.com/checkout",
		Amount:                49_000,
		ExpiresAt:             &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, nil, Config{})
	result, err := service.ProcessMidtransWebhook(context.Background(), ProcessMidtransWebhookInput{
		Payload: map[string]any{
			"order_id":           oid,
			"transaction_status": "refund",
			"gross_amount":       "49000.00",
			"status_code":        "206",
			"signature_key":      "",
		},
		ServerKey: "",
	})
	if err != nil {
		t.Fatalf("process unsupported event: %v", err)
	}
	if !result.Processed || result.Idempotent {
		t.Fatalf("unexpected webhook result: %+v", result)
	}

	delivery, err := transactionRepository.GetWebhookDeliveryByIdempotencyKey(
		context.Background(),
		"midtrans:"+oid+":refund",
	)
	if err != nil {
		t.Fatalf("get webhook delivery by idempotency key: %v", err)
	}
	if delivery.ProcessingStatus != billingdomain.WebhookProcessingStatusRejected {
		t.Fatalf("expected rejected processing status for unsupported event, got %s", delivery.ProcessingStatus)
	}
}
