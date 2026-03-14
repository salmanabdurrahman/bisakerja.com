package memory

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
)

type BillingRepository struct {
	mu                   sync.RWMutex
	transactions         map[string]billing.Transaction
	idempotencyToTxID    map[string]string
	mayarTransactionToID map[string]string
	webhookByIdempotency map[string]billing.WebhookDelivery
}

func NewBillingRepository() *BillingRepository {
	return &BillingRepository{
		transactions:         make(map[string]billing.Transaction),
		idempotencyToTxID:    make(map[string]string),
		mayarTransactionToID: make(map[string]string),
		webhookByIdempotency: make(map[string]billing.WebhookDelivery),
	}
}

func (r *BillingRepository) CreatePending(
	_ context.Context,
	input billing.CreatePendingTransactionInput,
) (billing.Transaction, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return billing.Transaction{}, fmt.Errorf("create pending transaction: user id is required")
	}
	if strings.TrimSpace(input.InvoiceID) == "" {
		return billing.Transaction{}, fmt.Errorf("create pending transaction: invoice id is required")
	}
	if strings.TrimSpace(input.MayarTransactionID) == "" {
		return billing.Transaction{}, fmt.Errorf("create pending transaction: mayar transaction id is required")
	}
	if strings.TrimSpace(input.CheckoutURL) == "" {
		return billing.Transaction{}, fmt.Errorf("create pending transaction: checkout url is required")
	}
	if input.Amount <= 0 {
		return billing.Transaction{}, fmt.Errorf("create pending transaction: amount must be > 0")
	}

	now := time.Now().UTC()
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)

	r.mu.Lock()
	defer r.mu.Unlock()

	if idempotencyKey != "" {
		if existingID, exists := r.idempotencyToTxID[idempotencyCompositeKey(userID, idempotencyKey)]; exists {
			existing, ok := r.transactions[existingID]
			if ok && existing.Status == billing.TransactionStatusPending {
				return cloneBillingTransaction(existing), nil
			}
		}
	}

	transactionID := "txn_" + randomHex(12)
	transaction := billing.Transaction{
		ID:                 transactionID,
		UserID:             userID,
		Provider:           input.Provider,
		PlanCode:           input.PlanCode,
		MayarTransactionID: strings.TrimSpace(input.MayarTransactionID),
		InvoiceID:          strings.TrimSpace(input.InvoiceID),
		CheckoutURL:        strings.TrimSpace(input.CheckoutURL),
		Amount:             input.Amount,
		Status:             billing.TransactionStatusPending,
		IdempotencyKey:     idempotencyKey,
		ExpiresAt:          cloneTime(input.ExpiresAt),
		Metadata:           cloneMap(input.Metadata),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	r.transactions[transactionID] = transaction
	r.mayarTransactionToID[strings.TrimSpace(input.MayarTransactionID)] = transactionID
	if idempotencyKey != "" {
		r.idempotencyToTxID[idempotencyCompositeKey(userID, idempotencyKey)] = transactionID
	}

	return cloneBillingTransaction(transaction), nil
}

func (r *BillingRepository) FindPendingByUserAndIdempotencyKey(
	_ context.Context,
	userID string,
	idempotencyKey string,
	window time.Duration,
	now time.Time,
) (billing.Transaction, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedKey := strings.TrimSpace(idempotencyKey)
	if normalizedUserID == "" || normalizedKey == "" {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	txID, exists := r.idempotencyToTxID[idempotencyCompositeKey(normalizedUserID, normalizedKey)]
	if !exists {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}

	item, exists := r.transactions[txID]
	if !exists {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}
	if item.UserID != normalizedUserID {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}
	if item.Status != billing.TransactionStatusPending {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}
	if window > 0 && now.UTC().After(item.CreatedAt.Add(window)) {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}

	return cloneBillingTransaction(item), nil
}

func (r *BillingRepository) GetByMayarTransactionID(
	_ context.Context,
	mayarTransactionID string,
) (billing.Transaction, error) {
	normalizedID := strings.TrimSpace(mayarTransactionID)
	if normalizedID == "" {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	txID, exists := r.mayarTransactionToID[normalizedID]
	if !exists {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}
	item, exists := r.transactions[txID]
	if !exists {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}
	return cloneBillingTransaction(item), nil
}

func (r *BillingRepository) ListByUser(
	_ context.Context,
	userID string,
) ([]billing.Transaction, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []billing.Transaction{}, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]billing.Transaction, 0)
	for _, item := range r.transactions {
		if item.UserID != normalizedUserID {
			continue
		}
		result = append(result, cloneBillingTransaction(item))
	}
	slices.SortFunc(result, func(left, right billing.Transaction) int {
		if left.CreatedAt.Equal(right.CreatedAt) {
			return strings.Compare(right.ID, left.ID)
		}
		if left.CreatedAt.After(right.CreatedAt) {
			return -1
		}
		return 1
	})
	return result, nil
}

func (r *BillingRepository) ListAll(_ context.Context) ([]billing.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]billing.Transaction, 0, len(r.transactions))
	for _, item := range r.transactions {
		result = append(result, cloneBillingTransaction(item))
	}
	slices.SortFunc(result, func(left, right billing.Transaction) int {
		if left.CreatedAt.Equal(right.CreatedAt) {
			return strings.Compare(right.ID, left.ID)
		}
		if left.CreatedAt.After(right.CreatedAt) {
			return -1
		}
		return 1
	})
	return result, nil
}

func (r *BillingRepository) UpdateStatusByMayarTransactionID(
	_ context.Context,
	mayarTransactionID string,
	status billing.TransactionStatus,
	metadata map[string]any,
	updatedAt time.Time,
) (billing.Transaction, error) {
	normalizedID := strings.TrimSpace(mayarTransactionID)
	if normalizedID == "" {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	txID, exists := r.mayarTransactionToID[normalizedID]
	if !exists {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}
	item, exists := r.transactions[txID]
	if !exists {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}

	item.Status = status
	item.Metadata = cloneMap(metadata)
	item.UpdatedAt = updatedAt.UTC()
	r.transactions[txID] = item

	return cloneBillingTransaction(item), nil
}

func (r *BillingRepository) GetWebhookDeliveryByIdempotencyKey(
	_ context.Context,
	idempotencyKey string,
) (billing.WebhookDelivery, error) {
	normalizedKey := strings.TrimSpace(idempotencyKey)
	if normalizedKey == "" {
		return billing.WebhookDelivery{}, billing.ErrWebhookDeliveryNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	delivery, exists := r.webhookByIdempotency[normalizedKey]
	if !exists {
		return billing.WebhookDelivery{}, billing.ErrWebhookDeliveryNotFound
	}
	return cloneWebhookDelivery(delivery), nil
}

func (r *BillingRepository) RecordWebhookDelivery(
	_ context.Context,
	delivery billing.WebhookDelivery,
) (billing.WebhookDelivery, error) {
	idempotencyKey := strings.TrimSpace(delivery.IdempotencyKey)
	if idempotencyKey == "" {
		return billing.WebhookDelivery{}, fmt.Errorf("record webhook delivery: idempotency key is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.webhookByIdempotency[idempotencyKey]; exists {
		return billing.WebhookDelivery{}, billing.ErrWebhookDeliveryAlreadyExists
	}

	now := time.Now().UTC()
	stored := billing.WebhookDelivery{
		ID:               "whd_" + randomHex(12),
		Provider:         delivery.Provider,
		EventType:        strings.TrimSpace(delivery.EventType),
		TransactionID:    strings.TrimSpace(delivery.TransactionID),
		IdempotencyKey:   idempotencyKey,
		ProcessingStatus: delivery.ProcessingStatus,
		Payload:          cloneMap(delivery.Payload),
		ErrorMessage:     strings.TrimSpace(delivery.ErrorMessage),
		ReceivedAt:       now,
		ProcessedAt:      cloneTime(delivery.ProcessedAt),
	}
	if stored.ProcessedAt == nil {
		processedAt := now
		stored.ProcessedAt = &processedAt
	}

	r.webhookByIdempotency[idempotencyKey] = stored
	return cloneWebhookDelivery(stored), nil
}

func idempotencyCompositeKey(userID, idempotencyKey string) string {
	return strings.ToLower(strings.TrimSpace(userID)) + "|" + strings.TrimSpace(idempotencyKey)
}

func cloneBillingTransaction(value billing.Transaction) billing.Transaction {
	result := value
	result.ExpiresAt = cloneTime(value.ExpiresAt)
	result.Metadata = cloneMap(value.Metadata)
	return result
}

func cloneWebhookDelivery(value billing.WebhookDelivery) billing.WebhookDelivery {
	result := value
	result.Payload = cloneMap(value.Payload)
	result.ProcessedAt = cloneTime(value.ProcessedAt)
	return result
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}

func cloneMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}
