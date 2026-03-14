package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
)

type BillingRepository struct {
	mu                sync.RWMutex
	transactions      map[string]billing.Transaction
	idempotencyToTxID map[string]string
}

func NewBillingRepository() *BillingRepository {
	return &BillingRepository{
		transactions:      make(map[string]billing.Transaction),
		idempotencyToTxID: make(map[string]string),
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

func idempotencyCompositeKey(userID, idempotencyKey string) string {
	return strings.ToLower(strings.TrimSpace(userID)) + "|" + strings.TrimSpace(idempotencyKey)
}

func cloneBillingTransaction(value billing.Transaction) billing.Transaction {
	result := value
	result.ExpiresAt = cloneTime(value.ExpiresAt)
	result.Metadata = cloneMap(value.Metadata)
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
