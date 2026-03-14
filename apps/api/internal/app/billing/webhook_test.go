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

func TestService_ProcessMayarWebhook_PaymentReceivedAndDuplicate(t *testing.T) {
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
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:             user.ID,
		Provider:           billingdomain.PaymentProviderMayar,
		PlanCode:           billingdomain.PlanCodeProMonthly,
		MayarTransactionID: "trx_webhook_success",
		InvoiceID:          "inv_webhook_success",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, nil, Config{})
	fixedNow := time.Date(2026, time.March, 14, 0, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }

	result, err := service.ProcessMayarWebhook(context.Background(), ProcessMayarWebhookInput{
		Payload: map[string]any{
			"event": "payment.received",
			"data": map[string]any{
				"transactionId":     "trx_webhook_success",
				"transactionStatus": "paid",
				"customerEmail":     "webhook-success@example.com",
			},
		},
	})
	if err != nil {
		t.Fatalf("process webhook: %v", err)
	}
	if !result.Processed || result.Idempotent {
		t.Fatalf("unexpected result: %+v", result)
	}

	updatedTransaction, err := transactionRepository.GetByMayarTransactionID(context.Background(), "trx_webhook_success")
	if err != nil {
		t.Fatalf("get transaction by mayar id: %v", err)
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

	duplicate, err := service.ProcessMayarWebhook(context.Background(), ProcessMayarWebhookInput{
		Payload: map[string]any{
			"event": "payment.received",
			"data": map[string]any{
				"transactionId":     "trx_webhook_success",
				"transactionStatus": "paid",
				"customerEmail":     "webhook-success@example.com",
			},
		},
	})
	if err != nil {
		t.Fatalf("process duplicate webhook: %v", err)
	}
	if !duplicate.Idempotent {
		t.Fatalf("expected duplicate webhook idempotent=true, got %+v", duplicate)
	}
}

func TestService_ProcessMayarWebhook_PaymentReminder(t *testing.T) {
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
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:             user.ID,
		Provider:           billingdomain.PaymentProviderMayar,
		PlanCode:           billingdomain.PlanCodeProMonthly,
		MayarTransactionID: "trx_webhook_reminder",
		InvoiceID:          "inv_webhook_reminder",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, nil, Config{})
	_, err = service.ProcessMayarWebhook(context.Background(), ProcessMayarWebhookInput{
		Payload: map[string]any{
			"event": "payment.reminder",
			"data": map[string]any{
				"transactionId": "trx_webhook_reminder",
				"customerEmail": "webhook-reminder@example.com",
			},
		},
	})
	if err != nil {
		t.Fatalf("process webhook reminder: %v", err)
	}

	updatedTransaction, err := transactionRepository.GetByMayarTransactionID(context.Background(), "trx_webhook_reminder")
	if err != nil {
		t.Fatalf("get transaction by mayar id: %v", err)
	}
	if updatedTransaction.Status != billingdomain.TransactionStatusReminder {
		t.Fatalf("expected transaction status reminder, got %s", updatedTransaction.Status)
	}

	updatedUser, err := identityRepository.GetUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if updatedUser.IsPremium {
		t.Fatal("expected reminder webhook not activating premium")
	}
}

func TestService_ProcessMayarWebhook_ErrorCases(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	transactionRepository := memory.NewBillingRepository()
	service := NewService(identityRepository, transactionRepository, nil, Config{})

	_, err := service.ProcessMayarWebhook(context.Background(), ProcessMayarWebhookInput{
		Payload: map[string]any{
			"event": "payment.received",
		},
	})
	if !errors.Is(err, ErrInvalidWebhookPayload) {
		t.Fatalf("expected ErrInvalidWebhookPayload, got %v", err)
	}

	now := time.Now().UTC()
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:             "usr_missing",
		Provider:           billingdomain.PaymentProviderMayar,
		PlanCode:           billingdomain.PlanCodeProMonthly,
		MayarTransactionID: "trx_user_missing",
		InvoiceID:          "inv_user_missing",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	_, err = service.ProcessMayarWebhook(context.Background(), ProcessMayarWebhookInput{
		Payload: map[string]any{
			"event": "payment.received",
			"data": map[string]any{
				"transactionId":     "trx_user_missing",
				"transactionStatus": "paid",
				"customerEmail":     "missing-user@example.com",
			},
		},
	})
	if !errors.Is(err, ErrWebhookUserNotFound) {
		t.Fatalf("expected ErrWebhookUserNotFound, got %v", err)
	}

	delivery, err := transactionRepository.GetWebhookDeliveryByIdempotencyKey(
		context.Background(),
		"mayar:payment.received:trx_user_missing",
	)
	if err != nil {
		t.Fatalf("get webhook delivery by idempotency key: %v", err)
	}
	if delivery.ProcessingStatus != billingdomain.WebhookProcessingStatusRejected {
		t.Fatalf("expected rejected webhook delivery, got %s", delivery.ProcessingStatus)
	}
}

func TestService_ProcessMayarWebhook_UnsupportedEventRecorded(t *testing.T) {
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
	_, err = transactionRepository.CreatePending(context.Background(), billingdomain.CreatePendingTransactionInput{
		UserID:             user.ID,
		Provider:           billingdomain.PaymentProviderMayar,
		PlanCode:           billingdomain.PlanCodeProMonthly,
		MayarTransactionID: "trx_ignored",
		InvoiceID:          "inv_ignored",
		CheckoutURL:        "https://pay.example.com/checkout",
		Amount:             49_000,
		ExpiresAt:          &now,
	})
	if err != nil {
		t.Fatalf("create pending transaction: %v", err)
	}

	service := NewService(identityRepository, transactionRepository, nil, Config{})
	result, err := service.ProcessMayarWebhook(context.Background(), ProcessMayarWebhookInput{
		Payload: map[string]any{
			"event": "membership.memberExpired",
			"data": map[string]any{
				"transactionId": "trx_ignored",
				"customerEmail": "webhook-ignored@example.com",
			},
		},
	})
	if err != nil {
		t.Fatalf("process unsupported event: %v", err)
	}
	if !result.Processed || result.Idempotent {
		t.Fatalf("unexpected webhook result: %+v", result)
	}

	delivery, err := transactionRepository.GetWebhookDeliveryByIdempotencyKey(
		context.Background(),
		"mayar:membership.memberexpired:trx_ignored",
	)
	if err != nil {
		t.Fatalf("get webhook delivery by idempotency key: %v", err)
	}
	if delivery.ProcessingStatus != billingdomain.WebhookProcessingStatusRejected {
		t.Fatalf("expected rejected processing status for unsupported event, got %s", delivery.ProcessingStatus)
	}
}
