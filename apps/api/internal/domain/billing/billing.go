package billing

import (
	"context"
	"errors"
	"time"
)

type PlanCode string

const (
	PlanCodeProMonthly PlanCode = "pro_monthly"
)

type PaymentProvider string

const (
	PaymentProviderMayar PaymentProvider = "mayar"
)

type TransactionStatus string

const (
	TransactionStatusPending  TransactionStatus = "pending"
	TransactionStatusReminder TransactionStatus = "reminder"
	TransactionStatusSuccess  TransactionStatus = "success"
	TransactionStatusFailed   TransactionStatus = "failed"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrProviderRateLimited = errors.New("provider rate limited")
	ErrProviderUpstream    = errors.New("provider upstream error")
	ErrProviderUnavailable = errors.New("provider unavailable")
)

type Transaction struct {
	ID                 string
	UserID             string
	Provider           PaymentProvider
	PlanCode           PlanCode
	MayarTransactionID string
	InvoiceID          string
	CheckoutURL        string
	Amount             int64
	Status             TransactionStatus
	IdempotencyKey     string
	ExpiresAt          *time.Time
	Metadata           map[string]any
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreatePendingTransactionInput struct {
	UserID             string
	Provider           PaymentProvider
	PlanCode           PlanCode
	MayarTransactionID string
	InvoiceID          string
	CheckoutURL        string
	Amount             int64
	IdempotencyKey     string
	ExpiresAt          *time.Time
	Metadata           map[string]any
}

type Repository interface {
	CreatePending(ctx context.Context, input CreatePendingTransactionInput) (Transaction, error)
	FindPendingByUserAndIdempotencyKey(
		ctx context.Context,
		userID string,
		idempotencyKey string,
		window time.Duration,
		now time.Time,
	) (Transaction, error)
}

type EnsureCustomerInput struct {
	Name  string
	Email string
}

type Customer struct {
	ID    string
	Email string
	Name  string
}

type CreateInvoiceInput struct {
	CustomerID  string
	PlanCode    PlanCode
	Amount      int64
	Description string
	RedirectURL string
	ExternalID  string
}

type Invoice struct {
	ID            string
	TransactionID string
	CheckoutURL   string
	Amount        int64
	ExpiresAt     *time.Time
}

type Provider interface {
	EnsureCustomer(ctx context.Context, input EnsureCustomerInput) (Customer, error)
	CreateInvoice(ctx context.Context, input CreateInvoiceInput) (Invoice, error)
}
