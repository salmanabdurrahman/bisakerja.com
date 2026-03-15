package billing

import (
	"context"
	"errors"
	"time"
)

// PlanCode represents plan code.
type PlanCode string

const (
	PlanCodeProMonthly PlanCode = "pro_monthly"
)

// PaymentProvider represents payment provider.
type PaymentProvider string

const (
	PaymentProviderMayar    PaymentProvider = "mayar"
	PaymentProviderMidtrans PaymentProvider = "midtrans"
)

// TransactionStatus describes status details for transaction.
type TransactionStatus string

const (
	TransactionStatusPending  TransactionStatus = "pending"
	TransactionStatusReminder TransactionStatus = "reminder"
	TransactionStatusSuccess  TransactionStatus = "success"
	TransactionStatusFailed   TransactionStatus = "failed"
)

var (
	ErrTransactionNotFound          = errors.New("transaction not found")
	ErrWebhookDeliveryNotFound      = errors.New("webhook delivery not found")
	ErrWebhookDeliveryAlreadyExists = errors.New("webhook delivery already exists")
	ErrCouponInvalid                = errors.New("coupon invalid")
	ErrProviderRateLimited          = errors.New("provider rate limited")
	ErrProviderUpstream             = errors.New("provider upstream error")
	ErrProviderUnavailable          = errors.New("provider unavailable")
)

// Transaction represents transaction.
type Transaction struct {
	ID                    string
	UserID                string
	Provider              PaymentProvider
	PlanCode              PlanCode
	ProviderTransactionID string
	InvoiceID             string
	CheckoutURL           string
	Amount                int64
	Status                TransactionStatus
	IdempotencyKey        string
	ExpiresAt             *time.Time
	Metadata              map[string]any
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// WebhookProcessingStatus describes status details for webhook processing.
type WebhookProcessingStatus string

const (
	WebhookProcessingStatusProcessed        WebhookProcessingStatus = "processed"
	WebhookProcessingStatusIgnoredDuplicate WebhookProcessingStatus = "ignored_duplicate"
	WebhookProcessingStatusRejected         WebhookProcessingStatus = "rejected"
)

// WebhookDelivery represents webhook delivery.
type WebhookDelivery struct {
	ID               string
	Provider         PaymentProvider
	EventType        string
	TransactionID    string
	IdempotencyKey   string
	ProcessingStatus WebhookProcessingStatus
	Payload          map[string]any
	ErrorMessage     string
	ReceivedAt       time.Time
	ProcessedAt      *time.Time
}

// CreatePendingTransactionInput contains input parameters for create pending transaction.
type CreatePendingTransactionInput struct {
	UserID                string
	Provider              PaymentProvider
	PlanCode              PlanCode
	ProviderTransactionID string
	InvoiceID             string
	CheckoutURL           string
	Amount                int64
	IdempotencyKey        string
	ExpiresAt             *time.Time
	Metadata              map[string]any
}

// Repository defines behavior for repository.
type Repository interface {
	CreatePending(ctx context.Context, input CreatePendingTransactionInput) (Transaction, error)
	GetByProviderTransactionID(ctx context.Context, providerTransactionID string) (Transaction, error)
	ListByUser(ctx context.Context, userID string) ([]Transaction, error)
	ListAll(ctx context.Context) ([]Transaction, error)
	UpdateStatusByProviderTransactionID(
		ctx context.Context,
		providerTransactionID string,
		status TransactionStatus,
		metadata map[string]any,
		updatedAt time.Time,
	) (Transaction, error)
	FindPendingByUserAndIdempotencyKey(
		ctx context.Context,
		userID string,
		idempotencyKey string,
		window time.Duration,
		now time.Time,
	) (Transaction, error)
	GetWebhookDeliveryByIdempotencyKey(ctx context.Context, idempotencyKey string) (WebhookDelivery, error)
	RecordWebhookDelivery(ctx context.Context, delivery WebhookDelivery) (WebhookDelivery, error)
}

// EnsureCustomerInput contains input parameters for ensure customer.
type EnsureCustomerInput struct {
	Name   string
	Email  string
	Mobile string
}

// Customer represents customer.
type Customer struct {
	ID    string
	Email string
	Name  string
}

// CreateInvoiceInput contains input parameters for create invoice.
type CreateInvoiceInput struct {
	CustomerID     string
	CustomerName   string
	CustomerEmail  string
	CustomerMobile string
	PlanCode       PlanCode
	Amount         int64
	Description    string
	RedirectURL    string
	ExternalID     string
	ExpiresAt      *time.Time
}

// Invoice represents invoice.
type Invoice struct {
	ID            string
	TransactionID string
	CheckoutURL   string
	SnapToken     string
	Amount        int64
	ExpiresAt     *time.Time
}

// ValidateCouponInput contains input parameters for coupon validation.
type ValidateCouponInput struct {
	Code   string
	Amount int64
}

// Coupon represents normalized coupon validation result.
type Coupon struct {
	Code           string
	DiscountAmount int64
	FinalAmount    int64
}

// Provider defines behavior for provider.
type Provider interface {
	EnsureCustomer(ctx context.Context, input EnsureCustomerInput) (Customer, error)
	CreateInvoice(ctx context.Context, input CreateInvoiceInput) (Invoice, error)
}

// CouponValidator defines behavior for coupon validation provider.
type CouponValidator interface {
	ValidateCoupon(ctx context.Context, input ValidateCouponInput) (Coupon, error)
}

// InvoiceSnapshot represents invoice snapshot.
type InvoiceSnapshot struct {
	InvoiceID         string
	TransactionID     string
	TransactionStatus string
	CustomerEmail     string
	Amount            int64
	UpdatedAt         *time.Time
}

// ReconciliationProvider defines behavior for reconciliation provider.
type ReconciliationProvider interface {
	GetInvoiceByID(ctx context.Context, invoiceID string) (InvoiceSnapshot, error)
}
