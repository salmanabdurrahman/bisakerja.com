package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
)

// ReconciliationSummary summarizes execution details for reconciliation.
type ReconciliationSummary struct {
	ScannedTransactions int
	ReconciledCount     int
	RetryableFailures   int
	AnomalyCount        int
}

// ReconcileWithMayar handles reconcile with mayar.
func (s *Service) ReconcileWithMayar(ctx context.Context) (ReconciliationSummary, error) {
	if s.identityRepository == nil || s.repository == nil {
		return ReconciliationSummary{}, errors.New("billing service dependency is not fully configured")
	}

	reconciler, ok := s.provider.(billingdomain.ReconciliationProvider)
	if !ok || reconciler == nil {
		return ReconciliationSummary{}, nil
	}

	transactions, err := s.repository.ListAll(ctx)
	if err != nil {
		return ReconciliationSummary{}, fmt.Errorf("list transactions for reconciliation: %w", err)
	}

	now := s.now().UTC()
	summary := ReconciliationSummary{}
	for _, item := range transactions {
		if item.Status != billingdomain.TransactionStatusPending &&
			item.Status != billingdomain.TransactionStatusReminder {
			continue
		}

		summary.ScannedTransactions++
		if now.Sub(item.UpdatedAt) >= 24*time.Hour {
			summary.AnomalyCount++
		}
		if strings.TrimSpace(item.InvoiceID) == "" {
			continue
		}

		invoice, invoiceErr := reconciler.GetInvoiceByID(ctx, item.InvoiceID)
		if invoiceErr != nil {
			if errors.Is(invoiceErr, billingdomain.ErrProviderRateLimited) ||
				errors.Is(invoiceErr, billingdomain.ErrProviderUnavailable) ||
				errors.Is(invoiceErr, billingdomain.ErrProviderUpstream) {
				summary.RetryableFailures++
				continue
			}
			return summary, fmt.Errorf("lookup mayar invoice %s: %w", item.InvoiceID, invoiceErr)
		}

		nextStatus, shouldUpdate := normalizeInvoiceStatus(invoice.TransactionStatus)
		if !shouldUpdate || nextStatus == item.Status {
			continue
		}

		_, updateErr := s.repository.UpdateStatusByMayarTransactionID(
			ctx,
			item.MayarTransactionID,
			nextStatus,
			map[string]any{
				"reconciliation_invoice_id":     invoice.InvoiceID,
				"reconciliation_invoice_status": invoice.TransactionStatus,
				"reconciliation_updated_at":     now,
			},
			now,
		)
		if updateErr != nil {
			return summary, fmt.Errorf("update transaction from reconciliation: %w", updateErr)
		}
		summary.ReconciledCount++

		if nextStatus == billingdomain.TransactionStatusSuccess {
			user, userErr := s.identityRepository.GetUserByID(ctx, item.UserID)
			if userErr != nil {
				return summary, fmt.Errorf("load user for premium activation: %w", userErr)
			}

			newExpiry := nextPremiumExpiry(user.PremiumExpiredAt, now, 30*24*time.Hour)
			if _, premiumErr := s.identityRepository.UpdatePremiumStatus(
				ctx,
				user.ID,
				true,
				&newExpiry,
			); premiumErr != nil {
				return summary, fmt.Errorf("update user premium status: %w", premiumErr)
			}
		}
	}

	return summary, nil
}

func normalizeInvoiceStatus(raw string) (billingdomain.TransactionStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "paid", "success", "completed":
		return billingdomain.TransactionStatusSuccess, true
	case "reminder":
		return billingdomain.TransactionStatusReminder, true
	case "pending", "unpaid", "open", "waiting":
		return billingdomain.TransactionStatusPending, true
	case "failed", "expired", "cancelled", "canceled", "void":
		return billingdomain.TransactionStatusFailed, true
	default:
		return "", false
	}
}
