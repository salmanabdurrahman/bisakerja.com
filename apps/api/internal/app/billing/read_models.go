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
	ErrInvalidPage              = errors.New("invalid page")
	ErrInvalidLimit             = errors.New("invalid limit")
	ErrInvalidTransactionStatus = errors.New("invalid transaction status")
)

// BillingStatus describes status details for billing.
type BillingStatus struct {
	PlanCode              string
	SubscriptionState     identity.SubscriptionState
	IsPremium             bool
	PremiumExpiredAt      *time.Time
	LastTransactionStatus string
}

// ListTransactionsInput contains input parameters for list transactions.
type ListTransactionsInput struct {
	UserID string
	Page   int
	Limit  int
	Status string
}

// TransactionRecord represents a persisted record for transaction.
type TransactionRecord struct {
	ID                 string
	Provider           billingdomain.PaymentProvider
	MayarTransactionID string
	Amount             int64
	Status             billingdomain.TransactionStatus
	CreatedAt          time.Time
}

// ListTransactionsResult contains result values for list transactions.
type ListTransactionsResult struct {
	Data         []TransactionRecord
	Page         int
	Limit        int
	TotalPages   int
	TotalRecords int
}

// GetBillingStatus returns billing status.
func (s *Service) GetBillingStatus(ctx context.Context, userID string) (BillingStatus, error) {
	if s.identityRepository == nil || s.repository == nil {
		return BillingStatus{}, errors.New("billing service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return BillingStatus{}, identity.ErrUserNotFound
	}

	user, err := s.identityRepository.GetUserByID(ctx, normalizedUserID)
	if err != nil {
		return BillingStatus{}, fmt.Errorf("get user profile: %w", err)
	}

	transactions, err := s.repository.ListByUser(ctx, normalizedUserID)
	if err != nil {
		return BillingStatus{}, fmt.Errorf("list user transactions: %w", err)
	}

	var latest *billingdomain.Transaction
	if len(transactions) > 0 {
		latest = &transactions[0]
	}

	now := s.now().UTC()
	subscriptionState := deriveBillingSubscriptionState(user, latest, now)
	planCode := string(billingdomain.PlanCodeProMonthly)
	lastTransactionStatus := ""

	if latest != nil {
		if latest.PlanCode != "" {
			planCode = string(latest.PlanCode)
		}
		lastTransactionStatus = string(latest.Status)
	}

	return BillingStatus{
		PlanCode:              planCode,
		SubscriptionState:     subscriptionState,
		IsPremium:             user.IsPremium,
		PremiumExpiredAt:      user.PremiumExpiredAt,
		LastTransactionStatus: lastTransactionStatus,
	}, nil
}

// ListBillingTransactions returns a list of billing transactions.
func (s *Service) ListBillingTransactions(
	ctx context.Context,
	input ListTransactionsInput,
) (ListTransactionsResult, error) {
	if s.identityRepository == nil || s.repository == nil {
		return ListTransactionsResult{}, errors.New("billing service dependency is not fully configured")
	}

	normalizedUserID := strings.TrimSpace(input.UserID)
	if normalizedUserID == "" {
		return ListTransactionsResult{}, identity.ErrUserNotFound
	}

	if _, err := s.identityRepository.GetUserByID(ctx, normalizedUserID); err != nil {
		return ListTransactionsResult{}, fmt.Errorf("get user profile: %w", err)
	}

	page := input.Page
	if page == 0 {
		page = 1
	}
	if page < 1 {
		return ListTransactionsResult{}, ErrInvalidPage
	}

	limit := input.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return ListTransactionsResult{}, ErrInvalidLimit
	}

	statusFilter := strings.ToLower(strings.TrimSpace(input.Status))
	filteredStatus, hasStatusFilter := parseTransactionStatus(statusFilter)
	if statusFilter != "" && !hasStatusFilter {
		return ListTransactionsResult{}, ErrInvalidTransactionStatus
	}

	transactions, err := s.repository.ListByUser(ctx, normalizedUserID)
	if err != nil {
		return ListTransactionsResult{}, fmt.Errorf("list user transactions: %w", err)
	}

	filtered := make([]billingdomain.Transaction, 0, len(transactions))
	for _, item := range transactions {
		if hasStatusFilter && item.Status != filteredStatus {
			continue
		}
		filtered = append(filtered, item)
	}

	totalRecords := len(filtered)
	totalPages := 0
	if totalRecords > 0 {
		totalPages = (totalRecords + limit - 1) / limit
	}

	start := (page - 1) * limit
	if start >= totalRecords {
		return ListTransactionsResult{
			Data:         []TransactionRecord{},
			Page:         page,
			Limit:        limit,
			TotalPages:   totalPages,
			TotalRecords: totalRecords,
		}, nil
	}

	end := start + limit
	if end > totalRecords {
		end = totalRecords
	}

	result := make([]TransactionRecord, 0, end-start)
	for _, item := range filtered[start:end] {
		result = append(result, TransactionRecord{
			ID:                 item.ID,
			Provider:           item.Provider,
			MayarTransactionID: item.MayarTransactionID,
			Amount:             item.Amount,
			Status:             item.Status,
			CreatedAt:          item.CreatedAt,
		})
	}

	return ListTransactionsResult{
		Data:         result,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
	}, nil
}

func deriveBillingSubscriptionState(
	user identity.User,
	latestTransaction *billingdomain.Transaction,
	now time.Time,
) identity.SubscriptionState {
	if user.IsPremium {
		if user.PremiumExpiredAt == nil || user.PremiumExpiredAt.After(now) {
			return identity.SubscriptionStatePremiumActive
		}
		return identity.SubscriptionStatePremiumExpired
	}

	if user.PremiumExpiredAt != nil && !user.PremiumExpiredAt.After(now) {
		return identity.SubscriptionStatePremiumExpired
	}

	if latestTransaction != nil {
		switch latestTransaction.Status {
		case billingdomain.TransactionStatusPending, billingdomain.TransactionStatusReminder:
			return identity.SubscriptionStatePendingPayment
		}
	}

	return identity.SubscriptionStateFree
}

func parseTransactionStatus(raw string) (billingdomain.TransactionStatus, bool) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(billingdomain.TransactionStatusPending):
		return billingdomain.TransactionStatusPending, true
	case string(billingdomain.TransactionStatusReminder):
		return billingdomain.TransactionStatusReminder, true
	case string(billingdomain.TransactionStatusSuccess):
		return billingdomain.TransactionStatusSuccess, true
	case string(billingdomain.TransactionStatusFailed):
		return billingdomain.TransactionStatusFailed, true
	default:
		return "", false
	}
}
