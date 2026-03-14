package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
)

type BillingRepository struct {
	pool *pgxpool.Pool
}

func NewBillingRepository(pool *pgxpool.Pool) *BillingRepository {
	return &BillingRepository{pool: pool}
}

func (r *BillingRepository) CreatePending(
	ctx context.Context,
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

	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if idempotencyKey != "" {
		existing, err := r.FindPendingByUserAndIdempotencyKey(ctx, userID, idempotencyKey, 0, time.Now().UTC())
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, billing.ErrTransactionNotFound) {
			return billing.Transaction{}, fmt.Errorf("lookup existing idempotent transaction: %w", err)
		}
	}

	metadata, err := encodeJSON(input.Metadata)
	if err != nil {
		return billing.Transaction{}, err
	}

	insertQuery := `
WITH selected_user AS (
  SELECT id FROM users WHERE id::text = $1
)
INSERT INTO transactions (
  user_id,
  provider,
  plan_code,
  mayar_transaction_id,
  invoice_id,
  checkout_url,
  amount,
  status,
  idempotency_key,
  expires_at,
  metadata,
  created_at,
  updated_at
)
SELECT
  selected_user.id,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11::jsonb,
  now(),
  now()
FROM selected_user
ON CONFLICT (provider, mayar_transaction_id) DO NOTHING
RETURNING id::text, user_id::text, provider, plan_code, mayar_transaction_id, invoice_id, checkout_url, amount, status,
          idempotency_key, expires_at, metadata, created_at, updated_at
`

	created, scanErr := scanTransaction(
		r.pool.QueryRow(
			ctx,
			insertQuery,
			userID,
			string(input.Provider),
			string(input.PlanCode),
			strings.TrimSpace(input.MayarTransactionID),
			strings.TrimSpace(input.InvoiceID),
			strings.TrimSpace(input.CheckoutURL),
			input.Amount,
			string(billing.TransactionStatusPending),
			idempotencyKey,
			nullableTime(input.ExpiresAt),
			metadata,
		),
	)
	if scanErr == nil {
		return created, nil
	}
	if !errors.Is(scanErr, pgx.ErrNoRows) {
		return billing.Transaction{}, fmt.Errorf("create pending transaction: %w", scanErr)
	}

	// Conflict path (same provider + transaction ID): return existing row.
	existing, err := r.GetByMayarTransactionID(ctx, input.MayarTransactionID)
	if err == nil {
		return existing, nil
	}

	if errors.Is(err, billing.ErrTransactionNotFound) {
		return billing.Transaction{}, fmt.Errorf("create pending transaction: user not found")
	}
	return billing.Transaction{}, err
}

func (r *BillingRepository) FindPendingByUserAndIdempotencyKey(
	ctx context.Context,
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

	conditions := []string{
		"user_id::text = $1",
		"idempotency_key = $2",
		"status = $3",
	}
	args := []any{normalizedUserID, normalizedKey, string(billing.TransactionStatusPending)}

	if window > 0 {
		cutoff := now.UTC().Add(-window)
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)+1))
		args = append(args, cutoff)
	}

	query := `
SELECT id::text, user_id::text, provider, plan_code, mayar_transaction_id, invoice_id, checkout_url, amount, status,
       idempotency_key, expires_at, metadata, created_at, updated_at
FROM transactions
WHERE ` + strings.Join(conditions, " AND ") + `
ORDER BY created_at DESC, id DESC
LIMIT 1
`

	transaction, err := scanTransaction(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return billing.Transaction{}, billing.ErrTransactionNotFound
		}
		return billing.Transaction{}, fmt.Errorf("find pending transaction by idempotency key: %w", err)
	}
	return transaction, nil
}

func (r *BillingRepository) GetByMayarTransactionID(
	ctx context.Context,
	mayarTransactionID string,
) (billing.Transaction, error) {
	normalizedID := strings.TrimSpace(mayarTransactionID)
	if normalizedID == "" {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}

	query := `
SELECT id::text, user_id::text, provider, plan_code, mayar_transaction_id, invoice_id, checkout_url, amount, status,
       idempotency_key, expires_at, metadata, created_at, updated_at
FROM transactions
WHERE mayar_transaction_id = $1
`

	item, err := scanTransaction(r.pool.QueryRow(ctx, query, normalizedID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return billing.Transaction{}, billing.ErrTransactionNotFound
		}
		return billing.Transaction{}, fmt.Errorf("get transaction by mayar transaction id: %w", err)
	}
	return item, nil
}

func (r *BillingRepository) ListByUser(ctx context.Context, userID string) ([]billing.Transaction, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return []billing.Transaction{}, nil
	}

	query := `
SELECT id::text, user_id::text, provider, plan_code, mayar_transaction_id, invoice_id, checkout_url, amount, status,
       idempotency_key, expires_at, metadata, created_at, updated_at
FROM transactions
WHERE user_id::text = $1
ORDER BY created_at DESC, id DESC
`

	rows, err := r.pool.Query(ctx, query, normalizedUserID)
	if err != nil {
		return nil, fmt.Errorf("list transactions by user: %w", err)
	}
	defer rows.Close()

	result := make([]billing.Transaction, 0)
	for rows.Next() {
		item, scanErr := scanTransaction(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list transactions by user rows: %w", err)
	}

	return result, nil
}

func (r *BillingRepository) ListAll(ctx context.Context) ([]billing.Transaction, error) {
	query := `
SELECT id::text, user_id::text, provider, plan_code, mayar_transaction_id, invoice_id, checkout_url, amount, status,
       idempotency_key, expires_at, metadata, created_at, updated_at
FROM transactions
ORDER BY created_at DESC, id DESC
`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all transactions: %w", err)
	}
	defer rows.Close()

	result := make([]billing.Transaction, 0)
	for rows.Next() {
		item, scanErr := scanTransaction(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list all transactions rows: %w", err)
	}
	return result, nil
}

func (r *BillingRepository) UpdateStatusByMayarTransactionID(
	ctx context.Context,
	mayarTransactionID string,
	status billing.TransactionStatus,
	metadata map[string]any,
	updatedAt time.Time,
) (billing.Transaction, error) {
	normalizedID := strings.TrimSpace(mayarTransactionID)
	if normalizedID == "" {
		return billing.Transaction{}, billing.ErrTransactionNotFound
	}

	encodedMetadata, err := encodeJSON(metadata)
	if err != nil {
		return billing.Transaction{}, err
	}

	query := `
UPDATE transactions
SET status = $2, metadata = $3::jsonb, updated_at = $4
WHERE mayar_transaction_id = $1
RETURNING id::text, user_id::text, provider, plan_code, mayar_transaction_id, invoice_id, checkout_url, amount, status,
          idempotency_key, expires_at, metadata, created_at, updated_at
`

	updated, scanErr := scanTransaction(
		r.pool.QueryRow(ctx, query, normalizedID, string(status), encodedMetadata, updatedAt.UTC()),
	)
	if scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return billing.Transaction{}, billing.ErrTransactionNotFound
		}
		return billing.Transaction{}, fmt.Errorf("update transaction status: %w", scanErr)
	}

	return updated, nil
}

func (r *BillingRepository) GetWebhookDeliveryByIdempotencyKey(
	ctx context.Context,
	idempotencyKey string,
) (billing.WebhookDelivery, error) {
	normalizedKey := strings.TrimSpace(idempotencyKey)
	if normalizedKey == "" {
		return billing.WebhookDelivery{}, billing.ErrWebhookDeliveryNotFound
	}

	query := `
SELECT id::text, provider, event_type, COALESCE(transaction_id, ''), idempotency_key, processing_status, payload,
       COALESCE(error_message, ''), received_at, processed_at
FROM webhook_deliveries
WHERE idempotency_key = $1
`

	item, err := scanWebhookDelivery(r.pool.QueryRow(ctx, query, normalizedKey))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return billing.WebhookDelivery{}, billing.ErrWebhookDeliveryNotFound
		}
		return billing.WebhookDelivery{}, fmt.Errorf("get webhook delivery by idempotency key: %w", err)
	}
	return item, nil
}

func (r *BillingRepository) RecordWebhookDelivery(
	ctx context.Context,
	delivery billing.WebhookDelivery,
) (billing.WebhookDelivery, error) {
	idempotencyKey := strings.TrimSpace(delivery.IdempotencyKey)
	if idempotencyKey == "" {
		return billing.WebhookDelivery{}, fmt.Errorf("record webhook delivery: idempotency key is required")
	}

	payload, err := encodeJSON(delivery.Payload)
	if err != nil {
		return billing.WebhookDelivery{}, err
	}

	receivedAt := delivery.ReceivedAt.UTC()
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}
	processedAt := cloneTime(delivery.ProcessedAt)
	if processedAt == nil {
		defaultProcessedAt := receivedAt
		processedAt = &defaultProcessedAt
	}

	query := `
INSERT INTO webhook_deliveries (
  provider,
  event_type,
  transaction_id,
  idempotency_key,
  processing_status,
  payload,
  error_message,
  received_at,
  processed_at
)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9)
RETURNING id::text, provider, event_type, COALESCE(transaction_id, ''), idempotency_key, processing_status, payload,
          COALESCE(error_message, ''), received_at, processed_at
`

	stored, scanErr := scanWebhookDelivery(
		r.pool.QueryRow(
			ctx,
			query,
			string(delivery.Provider),
			strings.TrimSpace(delivery.EventType),
			strings.TrimSpace(delivery.TransactionID),
			idempotencyKey,
			string(delivery.ProcessingStatus),
			payload,
			strings.TrimSpace(delivery.ErrorMessage),
			receivedAt,
			nullableTime(processedAt),
		),
	)
	if scanErr != nil {
		if isUniqueViolation(scanErr) {
			return billing.WebhookDelivery{}, billing.ErrWebhookDeliveryAlreadyExists
		}
		return billing.WebhookDelivery{}, fmt.Errorf("record webhook delivery: %w", scanErr)
	}

	return stored, nil
}

type transactionScanner interface {
	Scan(dest ...any) error
}

func scanTransaction(scanner transactionScanner) (billing.Transaction, error) {
	var (
		item             billing.Transaction
		provider         string
		planCode         sql.NullString
		status           string
		mayarTxID        sql.NullString
		invoiceID        sql.NullString
		checkoutURL      sql.NullString
		idempotencyKey   sql.NullString
		expiresAt        sql.NullTime
		metadataRawValue []byte
	)

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&provider,
		&planCode,
		&mayarTxID,
		&invoiceID,
		&checkoutURL,
		&item.Amount,
		&status,
		&idempotencyKey,
		&expiresAt,
		&metadataRawValue,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return billing.Transaction{}, err
	}

	item.Provider = billing.PaymentProvider(provider)
	if planCode.Valid {
		item.PlanCode = billing.PlanCode(planCode.String)
	}
	if mayarTxID.Valid {
		item.MayarTransactionID = mayarTxID.String
	}
	if invoiceID.Valid {
		item.InvoiceID = invoiceID.String
	}
	if checkoutURL.Valid {
		item.CheckoutURL = checkoutURL.String
	}
	if idempotencyKey.Valid {
		item.IdempotencyKey = idempotencyKey.String
	}
	if expiresAt.Valid {
		value := expiresAt.Time.UTC()
		item.ExpiresAt = &value
	}
	item.Status = billing.TransactionStatus(status)
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()

	metadata, err := decodeJSON(metadataRawValue)
	if err != nil {
		return billing.Transaction{}, err
	}
	item.Metadata = metadata
	return item, nil
}

type webhookScanner interface {
	Scan(dest ...any) error
}

func scanWebhookDelivery(scanner webhookScanner) (billing.WebhookDelivery, error) {
	var (
		item            billing.WebhookDelivery
		provider        string
		processing      string
		payloadRawValue []byte
		processedAt     sql.NullTime
	)

	err := scanner.Scan(
		&item.ID,
		&provider,
		&item.EventType,
		&item.TransactionID,
		&item.IdempotencyKey,
		&processing,
		&payloadRawValue,
		&item.ErrorMessage,
		&item.ReceivedAt,
		&processedAt,
	)
	if err != nil {
		return billing.WebhookDelivery{}, err
	}

	item.Provider = billing.PaymentProvider(provider)
	item.ProcessingStatus = billing.WebhookProcessingStatus(processing)
	item.ReceivedAt = item.ReceivedAt.UTC()
	if processedAt.Valid {
		value := processedAt.Time.UTC()
		item.ProcessedAt = &value
	}

	payload, err := decodeJSON(payloadRawValue)
	if err != nil {
		return billing.WebhookDelivery{}, err
	}
	item.Payload = payload

	return item, nil
}

var _ billing.Repository = (*BillingRepository)(nil)
