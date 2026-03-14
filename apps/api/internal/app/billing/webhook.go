package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

var (
	ErrInvalidWebhookPayload = errors.New("invalid webhook payload")
	ErrWebhookUserNotFound   = errors.New("webhook user not found")
)

// ProcessMayarWebhookInput contains input parameters for process mayar webhook.
type ProcessMayarWebhookInput struct {
	Payload map[string]any
}

// ProcessMayarWebhookResult contains result values for process mayar webhook.
type ProcessMayarWebhookResult struct {
	Provider   billingdomain.PaymentProvider
	Processed  bool
	Idempotent bool
}

type parsedMayarWebhookPayload struct {
	Event             string
	TransactionID     string
	TransactionStatus string
	CustomerEmail     string
	Payload           map[string]any
}

// ProcessMayarWebhook handles process mayar webhook.
func (s *Service) ProcessMayarWebhook(
	ctx context.Context,
	input ProcessMayarWebhookInput,
) (ProcessMayarWebhookResult, error) {
	if s.identityRepository == nil || s.repository == nil {
		return ProcessMayarWebhookResult{}, errors.New("billing service dependency is not fully configured")
	}

	parsed, err := parseMayarWebhookPayload(input.Payload)
	if err != nil {
		return ProcessMayarWebhookResult{}, fmt.Errorf("parse webhook payload: %w", ErrInvalidWebhookPayload)
	}

	now := s.now().UTC()
	idempotencyKey := webhookIdempotencyKey(parsed.Event, parsed.TransactionID)
	existingDelivery, err := s.repository.GetWebhookDeliveryByIdempotencyKey(ctx, idempotencyKey)
	if err == nil && existingDelivery.ID != "" {
		return ProcessMayarWebhookResult{
			Provider:   billingdomain.PaymentProviderMayar,
			Processed:  true,
			Idempotent: true,
		}, nil
	}
	if err != nil && !errors.Is(err, billingdomain.ErrWebhookDeliveryNotFound) {
		return ProcessMayarWebhookResult{}, fmt.Errorf("lookup webhook delivery idempotency: %w", err)
	}

	user, err := s.identityRepository.GetUserByEmail(ctx, parsed.CustomerEmail)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			recordErr := s.recordWebhookDelivery(ctx, billingdomain.WebhookDelivery{
				Provider:         billingdomain.PaymentProviderMayar,
				EventType:        parsed.Event,
				TransactionID:    parsed.TransactionID,
				IdempotencyKey:   idempotencyKey,
				ProcessingStatus: billingdomain.WebhookProcessingStatusRejected,
				Payload:          shallowCloneMap(parsed.Payload),
				ErrorMessage:     "user not found",
				ProcessedAt:      &now,
			})
			if recordErr != nil && !errors.Is(recordErr, billingdomain.ErrWebhookDeliveryAlreadyExists) {
				return ProcessMayarWebhookResult{}, fmt.Errorf("record rejected webhook delivery: %w", recordErr)
			}
			return ProcessMayarWebhookResult{}, ErrWebhookUserNotFound
		}
		return ProcessMayarWebhookResult{}, fmt.Errorf("get user by webhook email: %w", err)
	}

	targetStatus, shouldUpdateTransaction, shouldActivatePremium := normalizeWebhookTransactionStatus(
		parsed.Event,
		parsed.TransactionStatus,
	)
	if shouldUpdateTransaction {
		_, updateErr := s.repository.UpdateStatusByMayarTransactionID(
			ctx,
			parsed.TransactionID,
			targetStatus,
			map[string]any{
				"webhook_event":              parsed.Event,
				"webhook_transaction_status": parsed.TransactionStatus,
				"webhook_payload":            shallowCloneMap(parsed.Payload),
				"customer_email":             parsed.CustomerEmail,
			},
			now,
		)
		if updateErr != nil {
			if errors.Is(updateErr, billingdomain.ErrTransactionNotFound) {
				return ProcessMayarWebhookResult{}, fmt.Errorf("update transaction status: %w", ErrServiceUnavailable)
			}
			return ProcessMayarWebhookResult{}, fmt.Errorf("update transaction status: %w", updateErr)
		}
	}

	if shouldActivatePremium {
		expiredAt := nextPremiumExpiry(user.PremiumExpiredAt, now, 30*24*time.Hour)
		_, premiumErr := s.identityRepository.UpdatePremiumStatus(ctx, user.ID, true, &expiredAt)
		if premiumErr != nil {
			return ProcessMayarWebhookResult{}, fmt.Errorf("activate premium from webhook: %w", ErrServiceUnavailable)
		}
	}

	processingStatus := billingdomain.WebhookProcessingStatusProcessed
	errorMessage := ""
	if !shouldUpdateTransaction {
		processingStatus = billingdomain.WebhookProcessingStatusRejected
		errorMessage = "unsupported event"
	}

	recordErr := s.recordWebhookDelivery(ctx, billingdomain.WebhookDelivery{
		Provider:         billingdomain.PaymentProviderMayar,
		EventType:        parsed.Event,
		TransactionID:    parsed.TransactionID,
		IdempotencyKey:   idempotencyKey,
		ProcessingStatus: processingStatus,
		Payload:          shallowCloneMap(parsed.Payload),
		ErrorMessage:     errorMessage,
		ProcessedAt:      &now,
	})
	if recordErr != nil {
		if errors.Is(recordErr, billingdomain.ErrWebhookDeliveryAlreadyExists) {
			return ProcessMayarWebhookResult{
				Provider:   billingdomain.PaymentProviderMayar,
				Processed:  true,
				Idempotent: true,
			}, nil
		}
		return ProcessMayarWebhookResult{}, fmt.Errorf("record webhook delivery: %w", recordErr)
	}

	return ProcessMayarWebhookResult{
		Provider:   billingdomain.PaymentProviderMayar,
		Processed:  true,
		Idempotent: false,
	}, nil
}

func webhookIdempotencyKey(event, transactionID string) string {
	return "mayar:" + strings.ToLower(strings.TrimSpace(event)) + ":" + strings.TrimSpace(transactionID)
}

func parseMayarWebhookPayload(raw map[string]any) (parsedMayarWebhookPayload, error) {
	if len(raw) == 0 {
		return parsedMayarWebhookPayload{}, ErrInvalidWebhookPayload
	}

	event := strings.ToLower(strings.TrimSpace(extractStringFromMap(raw,
		"event",
	)))
	transactionID := strings.TrimSpace(extractStringFromMap(raw,
		"data.transactionId",
		"data.transaction_id",
	))
	customerEmail := identity.NormalizeEmail(extractStringFromMap(raw,
		"data.customerEmail",
		"data.customer_email",
	))
	transactionStatus := strings.ToLower(strings.TrimSpace(extractStringFromMap(raw,
		"data.transactionStatus",
		"data.transaction_status",
	)))

	if event == "" || transactionID == "" || customerEmail == "" {
		return parsedMayarWebhookPayload{}, ErrInvalidWebhookPayload
	}

	return parsedMayarWebhookPayload{
		Event:             event,
		TransactionID:     transactionID,
		TransactionStatus: transactionStatus,
		CustomerEmail:     customerEmail,
		Payload:           shallowCloneMap(raw),
	}, nil
}

func normalizeWebhookTransactionStatus(
	event string,
	transactionStatus string,
) (billingdomain.TransactionStatus, bool, bool) {
	normalizedEvent := strings.ToLower(strings.TrimSpace(event))
	normalizedStatus := strings.ToLower(strings.TrimSpace(transactionStatus))

	switch normalizedEvent {
	case "payment.reminder":
		return billingdomain.TransactionStatusReminder, true, false
	case "payment.received":
		switch normalizedStatus {
		case "", "paid", "success", "completed":
			return billingdomain.TransactionStatusSuccess, true, true
		case "pending", "unpaid", "reminder":
			return billingdomain.TransactionStatusReminder, true, false
		case "failed", "expired", "cancelled", "canceled":
			return billingdomain.TransactionStatusFailed, true, false
		default:
			return billingdomain.TransactionStatusSuccess, true, true
		}
	default:
		return billingdomain.TransactionStatusPending, false, false
	}
}

func nextPremiumExpiry(current *time.Time, now time.Time, extension time.Duration) time.Time {
	base := now.UTC()
	if current != nil && current.After(base) {
		base = current.UTC()
	}
	return base.Add(extension).UTC()
}

func (s *Service) recordWebhookDelivery(ctx context.Context, delivery billingdomain.WebhookDelivery) error {
	_, err := s.repository.RecordWebhookDelivery(ctx, delivery)
	if err != nil {
		return err
	}
	return nil
}

func extractStringFromMap(payload map[string]any, paths ...string) string {
	for _, path := range paths {
		value, ok := lookupPath(payload, path)
		if !ok {
			continue
		}
		raw, ok := value.(string)
		if !ok {
			continue
		}
		trimmed := strings.TrimSpace(raw)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func lookupPath(payload map[string]any, path string) (any, bool) {
	current := any(payload)
	segments := strings.Split(path, ".")
	for _, segment := range segments {
		asMap, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		next, exists := asMap[segment]
		if !exists {
			return nil, false
		}
		current = next
	}
	return current, true
}

func shallowCloneMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}
