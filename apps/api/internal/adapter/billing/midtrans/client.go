package midtrans

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"

	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
)

// SandboxEnv and ProductionEnv are convenience aliases for the Midtrans SDK environment types.
var (
	SandboxEnv    = midtrans.Sandbox    // midtrans.EnvironmentType
	ProductionEnv = midtrans.Production // midtrans.EnvironmentType
)

// ClientConfig stores configuration values for client.
type ClientConfig struct {
	ServerKey string
	ClientKey string
	Env       midtrans.EnvironmentType // midtrans.Sandbox or midtrans.Production
	Logger    *slog.Logger
}

// Client represents midtrans client.
type Client struct {
	snapClient coreapi.Client
	snap       snap.Client
	logger     *slog.Logger
}

// NewClient creates a new midtrans client instance.
func NewClient(config ClientConfig) *Client {
	env := config.Env
	if env != midtrans.Sandbox && env != midtrans.Production {
		env = midtrans.Sandbox
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	coreClient := coreapi.Client{}
	coreClient.New(config.ServerKey, env)

	snapClient := snap.Client{}
	snapClient.New(config.ServerKey, env)

	return &Client{
		snapClient: coreClient,
		snap:       snapClient,
		logger:     logger,
	}
}

// EnsureCustomer is a no-op for Midtrans — customer details are embedded in CreateInvoice.
func (c *Client) EnsureCustomer(
	_ context.Context,
	input billingdomain.EnsureCustomerInput,
) (billingdomain.Customer, error) {
	return billingdomain.Customer{
		ID:    strings.TrimSpace(input.Email),
		Email: strings.TrimSpace(input.Email),
		Name:  strings.TrimSpace(input.Name),
	}, nil
}

// CreateInvoice creates a Midtrans Snap transaction and returns an invoice with snap token and redirect URL.
func (c *Client) CreateInvoice(
	_ context.Context,
	input billingdomain.CreateInvoiceInput,
) (billingdomain.Invoice, error) {
	if strings.TrimSpace(c.snap.ServerKey) == "" {
		return billingdomain.Invoice{}, fmt.Errorf("%w: midtrans server key is empty", billingdomain.ErrProviderUnavailable)
	}

	customerName := strings.TrimSpace(input.CustomerName)
	if customerName == "" {
		customerName = "Bisakerja " + string(input.PlanCode)
	}
	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = customerName
	}

	// Split name for Midtrans first/last name fields.
	firstName, lastName := splitName(customerName)

	// Determine expiry — default 24h.
	expiry := time.Now().UTC().Add(24 * time.Hour)
	if input.ExpiresAt != nil && !input.ExpiresAt.IsZero() {
		expiry = input.ExpiresAt.UTC()
	}

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  strings.TrimSpace(input.ExternalID),
			GrossAmt: input.Amount,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: firstName,
			LName: lastName,
			Email: strings.TrimSpace(input.CustomerEmail),
			Phone: strings.TrimSpace(input.CustomerMobile),
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    string(input.PlanCode),
				Name:  description,
				Price: input.Amount,
				Qty:   1,
			},
		},
		Expiry: &snap.ExpiryDetails{
			StartTime: time.Now().UTC().Format("2006-01-02 15:04:05 +0700"),
			Unit:      "minute",
			Duration:  int64(time.Until(expiry).Minutes()),
		},
	}

	snapResp, snapErr := c.snap.CreateTransaction(req)
	if snapErr != nil {
		return billingdomain.Invoice{}, mapMidtransError(snapErr, "create snap transaction")
	}

	if snapResp == nil || snapResp.Token == "" {
		return billingdomain.Invoice{}, fmt.Errorf("%w: missing snap token in response", billingdomain.ErrProviderUpstream)
	}

	expiresAt := expiry

	return billingdomain.Invoice{
		ID:            snapResp.Token,
		TransactionID: strings.TrimSpace(input.ExternalID),
		CheckoutURL:   snapResp.RedirectURL,
		SnapToken:     snapResp.Token,
		Amount:        input.Amount,
		ExpiresAt:     &expiresAt,
	}, nil
}

// GetInvoiceByID fetches a Midtrans transaction status by order ID (external ID).
func (c *Client) GetInvoiceByID(
	_ context.Context,
	invoiceID string,
) (billingdomain.InvoiceSnapshot, error) {
	trimmedID := strings.TrimSpace(invoiceID)
	if trimmedID == "" {
		return billingdomain.InvoiceSnapshot{}, fmt.Errorf("%w: invoice id is required", billingdomain.ErrProviderUpstream)
	}

	if strings.TrimSpace(c.snapClient.ServerKey) == "" {
		return billingdomain.InvoiceSnapshot{}, fmt.Errorf("%w: midtrans server key is empty", billingdomain.ErrProviderUnavailable)
	}

	resp, checkErr := c.snapClient.CheckTransaction(trimmedID)
	if checkErr != nil {
		return billingdomain.InvoiceSnapshot{}, mapMidtransError(checkErr, "check transaction")
	}

	if resp == nil {
		return billingdomain.InvoiceSnapshot{}, fmt.Errorf("%w: empty response from midtrans", billingdomain.ErrProviderUpstream)
	}

	// Normalize Midtrans status to our domain status strings.
	normalizedStatus := normalizeMidtransTransactionStatus(resp.TransactionStatus, resp.FraudStatus)

	var updatedAt *time.Time
	if raw := strings.TrimSpace(resp.TransactionTime); raw != "" {
		if parsed, parseErr := time.Parse("2006-01-02 15:04:05", raw); parseErr == nil {
			utc := parsed.UTC()
			updatedAt = &utc
		}
	}

	return billingdomain.InvoiceSnapshot{
		InvoiceID:         resp.OrderID,
		TransactionID:     resp.OrderID,
		TransactionStatus: normalizedStatus,
		CustomerEmail:     "",
		Amount:            parseAmount(resp.GrossAmount),
		UpdatedAt:         updatedAt,
	}, nil
}

// normalizeMidtransTransactionStatus maps Midtrans status+fraud to domain status strings.
func normalizeMidtransTransactionStatus(status, fraudStatus string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	f := strings.ToLower(strings.TrimSpace(fraudStatus))

	switch s {
	case "capture":
		if f == "accept" || f == "" {
			return "success"
		}
		return "failed"
	case "settlement":
		return "success"
	case "pending":
		return "pending"
	case "cancel", "expire", "deny":
		return "failed"
	default:
		return s
	}
}

// parseAmount parses a gross_amount string like "49000.00" to int64.
func parseAmount(raw string) int64 {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0
	}
	// Remove decimal part
	if dotIdx := strings.Index(trimmed, "."); dotIdx >= 0 {
		trimmed = trimmed[:dotIdx]
	}
	var result int64
	for _, ch := range trimmed {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int64(ch-'0')
		}
	}
	return result
}

// splitName splits a full name into first and last name.
func splitName(fullName string) (string, string) {
	trimmed := strings.TrimSpace(fullName)
	if trimmed == "" {
		return "", ""
	}
	idx := strings.LastIndex(trimmed, " ")
	if idx < 0 {
		return trimmed, ""
	}
	return strings.TrimSpace(trimmed[:idx]), strings.TrimSpace(trimmed[idx+1:])
}

// mapMidtransError converts a *midtrans.Error to a domain error.
func mapMidtransError(err *midtrans.Error, operation string) error {
	if err == nil {
		return nil
	}
	switch {
	case err.StatusCode == 429:
		return fmt.Errorf("%w: %s: midtrans returned 429", billingdomain.ErrProviderRateLimited, operation)
	case err.StatusCode >= 500:
		return fmt.Errorf("%w: %s: midtrans returned status %d", billingdomain.ErrProviderUnavailable, operation, err.StatusCode)
	case err.StatusCode == 404:
		return fmt.Errorf("%w: %s: resource not found", billingdomain.ErrProviderUpstream, operation)
	default:
		return fmt.Errorf("%w: %s: midtrans error %d: %s", billingdomain.ErrProviderUpstream, operation, err.StatusCode, err.Message)
	}
}

var _ billingdomain.Provider = (*Client)(nil)
var _ billingdomain.ReconciliationProvider = (*Client)(nil)
