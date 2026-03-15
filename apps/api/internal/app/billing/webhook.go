package billing

import (
	"context"
	"crypto/sha512"
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

// ProcessMidtransWebhookInput contains input parameters for process midtrans webhook.
type ProcessMidtransWebhookInput struct {
	Payload   map[string]any
	ServerKey string
}

// ProcessMidtransWebhookResult contains result values for process midtrans webhook.
type ProcessMidtransWebhookResult struct {
	Provider   billingdomain.PaymentProvider
	Processed  bool
	Idempotent bool
}

type parsedMidtransWebhookPayload struct {
	OrderID           string
	TransactionStatus string
	FraudStatus       string
	GrossAmount       string
	StatusCode        string
	SignatureKey      string
	Payload           map[string]any
}

// ProcessMidtransWebhook handles process midtrans webhook.
func (s *Service) ProcessMidtransWebhook(
	ctx context.Context,
	input ProcessMidtransWebhookInput,
) (ProcessMidtransWebhookResult, error) {
	if s.identityRepository == nil || s.repository == nil {
		return ProcessMidtransWebhookResult{}, errors.New("billing service dependency is not fully configured")
	}

	parsed, err := parseMidtransWebhookPayload(input.Payload)
	if err != nil {
		return ProcessMidtransWebhookResult{}, fmt.Errorf("parse webhook payload: %w", ErrInvalidWebhookPayload)
	}

	// Validate Midtrans signature: SHA512(order_id + status_code + gross_amount + server_key)
	if strings.TrimSpace(input.ServerKey) != "" {
		if !validateMidtransSignature(parsed.OrderID, parsed.StatusCode, parsed.GrossAmount, input.ServerKey, parsed.SignatureKey) {
			return ProcessMidtransWebhookResult{}, fmt.Errorf("signature mismatch: %w", ErrInvalidWebhookPayload)
		}
	}

	now := s.now().UTC()
	// Use order_id + transaction_status as idempotency to deduplicate same status events.
	idempotencyKey := webhookIdempotencyKey(parsed.OrderID, parsed.TransactionStatus)
	existingDelivery, err := s.repository.GetWebhookDeliveryByIdempotencyKey(ctx, idempotencyKey)
	if err == nil && existingDelivery.ID != "" {
		return ProcessMidtransWebhookResult{
			Provider:   billingdomain.PaymentProviderMidtrans,
			Processed:  true,
			Idempotent: true,
		}, nil
	}
	if err != nil && !errors.Is(err, billingdomain.ErrWebhookDeliveryNotFound) {
		return ProcessMidtransWebhookResult{}, fmt.Errorf("lookup webhook delivery idempotency: %w", err)
	}

	// Look up transaction by order_id to resolve the owner.
	// The order_id is stored as ProviderTransactionID.
	txn, err := s.repository.GetByProviderTransactionID(ctx, parsed.OrderID)
	if err != nil {
		if errors.Is(err, billingdomain.ErrTransactionNotFound) {
			recordErr := s.recordWebhookDelivery(ctx, billingdomain.WebhookDelivery{
				Provider:         billingdomain.PaymentProviderMidtrans,
				EventType:        parsed.TransactionStatus,
				TransactionID:    parsed.OrderID,
				IdempotencyKey:   idempotencyKey,
				ProcessingStatus: billingdomain.WebhookProcessingStatusRejected,
				Payload:          shallowCloneMap(parsed.Payload),
				ErrorMessage:     "transaction not found for order id",
				ProcessedAt:      &now,
			})
			if recordErr != nil && !errors.Is(recordErr, billingdomain.ErrWebhookDeliveryAlreadyExists) {
				return ProcessMidtransWebhookResult{}, fmt.Errorf("record rejected webhook delivery: %w", recordErr)
			}
			return ProcessMidtransWebhookResult{}, ErrWebhookUserNotFound
		}
		return ProcessMidtransWebhookResult{}, fmt.Errorf("get transaction by order id from webhook: %w", err)
	}
	userID := txn.UserID

	user, err := s.identityRepository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			recordErr := s.recordWebhookDelivery(ctx, billingdomain.WebhookDelivery{
				Provider:         billingdomain.PaymentProviderMidtrans,
				EventType:        parsed.TransactionStatus,
				TransactionID:    parsed.OrderID,
				IdempotencyKey:   idempotencyKey,
				ProcessingStatus: billingdomain.WebhookProcessingStatusRejected,
				Payload:          shallowCloneMap(parsed.Payload),
				ErrorMessage:     "user not found",
				ProcessedAt:      &now,
			})
			if recordErr != nil && !errors.Is(recordErr, billingdomain.ErrWebhookDeliveryAlreadyExists) {
				return ProcessMidtransWebhookResult{}, fmt.Errorf("record rejected webhook delivery: %w", recordErr)
			}
			return ProcessMidtransWebhookResult{}, ErrWebhookUserNotFound
		}
		return ProcessMidtransWebhookResult{}, fmt.Errorf("get user by id from webhook: %w", err)
	}

	targetStatus, shouldUpdateTransaction, shouldActivatePremium := normalizeMidtransWebhookStatus(
		parsed.TransactionStatus,
		parsed.FraudStatus,
	)
	if shouldUpdateTransaction {
		_, updateErr := s.repository.UpdateStatusByProviderTransactionID(
			ctx,
			parsed.OrderID,
			targetStatus,
			map[string]any{
				"webhook_transaction_status": parsed.TransactionStatus,
				"webhook_fraud_status":       parsed.FraudStatus,
				"webhook_payload":            shallowCloneMap(parsed.Payload),
				"customer_user_id":           userID,
			},
			now,
		)
		if updateErr != nil {
			if errors.Is(updateErr, billingdomain.ErrTransactionNotFound) {
				return ProcessMidtransWebhookResult{}, fmt.Errorf("update transaction status: %w", ErrServiceUnavailable)
			}
			return ProcessMidtransWebhookResult{}, fmt.Errorf("update transaction status: %w", updateErr)
		}
	}

	if shouldActivatePremium {
		expiredAt := nextPremiumExpiry(user.PremiumExpiredAt, now, 30*24*time.Hour)
		_, premiumErr := s.identityRepository.UpdatePremiumStatus(ctx, user.ID, true, &expiredAt)
		if premiumErr != nil {
			return ProcessMidtransWebhookResult{}, fmt.Errorf("activate premium from webhook: %w", ErrServiceUnavailable)
		}
	}

	processingStatus := billingdomain.WebhookProcessingStatusProcessed
	errorMessage := ""
	if !shouldUpdateTransaction {
		processingStatus = billingdomain.WebhookProcessingStatusRejected
		errorMessage = "unsupported transaction status"
	}

	recordErr := s.recordWebhookDelivery(ctx, billingdomain.WebhookDelivery{
		Provider:         billingdomain.PaymentProviderMidtrans,
		EventType:        parsed.TransactionStatus,
		TransactionID:    parsed.OrderID,
		IdempotencyKey:   idempotencyKey,
		ProcessingStatus: processingStatus,
		Payload:          shallowCloneMap(parsed.Payload),
		ErrorMessage:     errorMessage,
		ProcessedAt:      &now,
	})
	if recordErr != nil {
		if errors.Is(recordErr, billingdomain.ErrWebhookDeliveryAlreadyExists) {
			return ProcessMidtransWebhookResult{
				Provider:   billingdomain.PaymentProviderMidtrans,
				Processed:  true,
				Idempotent: true,
			}, nil
		}
		return ProcessMidtransWebhookResult{}, fmt.Errorf("record webhook delivery: %w", recordErr)
	}

	return ProcessMidtransWebhookResult{
		Provider:   billingdomain.PaymentProviderMidtrans,
		Processed:  true,
		Idempotent: false,
	}, nil
}

func webhookIdempotencyKey(orderID, transactionStatus string) string {
	return "midtrans:" + strings.TrimSpace(orderID) + ":" + strings.ToLower(strings.TrimSpace(transactionStatus))
}

func parseMidtransWebhookPayload(raw map[string]any) (parsedMidtransWebhookPayload, error) {
	if len(raw) == 0 {
		return parsedMidtransWebhookPayload{}, ErrInvalidWebhookPayload
	}

	orderID := strings.TrimSpace(extractStringFromMap(raw, "order_id"))
	transactionStatus := strings.ToLower(strings.TrimSpace(extractStringFromMap(raw, "transaction_status")))
	fraudStatus := strings.ToLower(strings.TrimSpace(extractStringFromMap(raw, "fraud_status")))
	grossAmount := strings.TrimSpace(extractStringFromMap(raw, "gross_amount"))
	statusCode := strings.TrimSpace(extractStringFromMap(raw, "status_code"))
	signatureKey := strings.TrimSpace(extractStringFromMap(raw, "signature_key"))

	if orderID == "" || transactionStatus == "" {
		return parsedMidtransWebhookPayload{}, ErrInvalidWebhookPayload
	}

	return parsedMidtransWebhookPayload{
		OrderID:           orderID,
		TransactionStatus: transactionStatus,
		FraudStatus:       fraudStatus,
		GrossAmount:       grossAmount,
		StatusCode:        statusCode,
		SignatureKey:      signatureKey,
		Payload:           shallowCloneMap(raw),
	}, nil
}

// normalizeMidtransWebhookStatus maps Midtrans transaction_status + fraud_status to our domain status.
// Returns: (targetStatus, shouldUpdateTransaction, shouldActivatePremium).
func normalizeMidtransWebhookStatus(
	transactionStatus string,
	fraudStatus string,
) (billingdomain.TransactionStatus, bool, bool) {
	s := strings.ToLower(strings.TrimSpace(transactionStatus))
	f := strings.ToLower(strings.TrimSpace(fraudStatus))

	switch s {
	case "capture":
		if f == "accept" || f == "" {
			return billingdomain.TransactionStatusSuccess, true, true
		}
		return billingdomain.TransactionStatusFailed, true, false
	case "settlement":
		return billingdomain.TransactionStatusSuccess, true, true
	case "pending":
		return billingdomain.TransactionStatusPending, true, false
	case "cancel", "expire", "deny":
		return billingdomain.TransactionStatusFailed, true, false
	default:
		return billingdomain.TransactionStatusPending, false, false
	}
}

// validateMidtransSignature checks SHA512(order_id + status_code + gross_amount + server_key).
func validateMidtransSignature(orderID, statusCode, grossAmount, serverKey, expected string) bool {
	raw := orderID + statusCode + grossAmount + serverKey
	sum := sha512.Sum512([]byte(raw))
	computed := fmt.Sprintf("%x", sum)
	return strings.EqualFold(computed, strings.TrimSpace(expected))
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
