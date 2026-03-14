package mayar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
)

// ClientConfig stores configuration values for client.
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
	HTTPClient *http.Client
	Sleep      func(time.Duration)
	RandIntn   func(int) int
}

// Client represents client.
type Client struct {
	baseURL    string
	apiKey     string
	maxRetries int
	httpClient *http.Client
	sleep      func(time.Duration)
	randIntn   func(int) int
}

// NewClient creates a new client instance.
func NewClient(config ClientConfig) *Client {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	maxRetries := config.MaxRetries
	if maxRetries < 0 {
		maxRetries = 3
	}
	if maxRetries == 0 {
		maxRetries = 3
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	} else if httpClient.Timeout <= 0 {
		httpClient.Timeout = timeout
	}

	sleep := config.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}

	randIntn := config.RandIntn
	if randIntn == nil {
		randIntn = rand.Intn
	}

	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.mayar.id/hl/v1"
	}

	return &Client{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(config.APIKey),
		maxRetries: maxRetries,
		httpClient: httpClient,
		sleep:      sleep,
		randIntn:   randIntn,
	}
}

// EnsureCustomer ensures customer.
func (c *Client) EnsureCustomer(
	ctx context.Context,
	input billingdomain.EnsureCustomerInput,
) (billingdomain.Customer, error) {
	responseBody, err := c.post(
		ctx,
		"/customer/create",
		map[string]any{
			"name":  strings.TrimSpace(input.Name),
			"email": strings.TrimSpace(input.Email),
		},
	)
	if err != nil {
		return billingdomain.Customer{}, err
	}

	customerID := extractString(responseBody,
		"data.id",
		"data.customer.id",
		"data.customer_id",
		"data.customerId",
		"customer.id",
		"customer_id",
		"customerId",
		"id",
	)
	if customerID == "" {
		return billingdomain.Customer{}, fmt.Errorf("%w: missing customer id", billingdomain.ErrProviderUpstream)
	}

	return billingdomain.Customer{
		ID:    customerID,
		Email: strings.TrimSpace(input.Email),
		Name:  strings.TrimSpace(input.Name),
	}, nil
}

// CreateInvoice creates invoice.
func (c *Client) CreateInvoice(
	ctx context.Context,
	input billingdomain.CreateInvoiceInput,
) (billingdomain.Invoice, error) {
	responseBody, err := c.post(
		ctx,
		"/invoice/create",
		map[string]any{
			"customer_id":          strings.TrimSpace(input.CustomerID),
			"customerId":           strings.TrimSpace(input.CustomerID),
			"name":                 "Bisakerja " + string(input.PlanCode),
			"description":          strings.TrimSpace(input.Description),
			"amount":               input.Amount,
			"success_redirect_url": strings.TrimSpace(input.RedirectURL),
			"successRedirectUrl":   strings.TrimSpace(input.RedirectURL),
			"external_id":          strings.TrimSpace(input.ExternalID),
			"externalId":           strings.TrimSpace(input.ExternalID),
		},
	)
	if err != nil {
		return billingdomain.Invoice{}, err
	}

	invoiceID := extractString(responseBody,
		"data.id",
		"data.invoice.id",
		"data.invoice_id",
		"data.invoiceId",
		"invoice.id",
		"invoice_id",
		"invoiceId",
		"id",
	)
	transactionID := extractString(responseBody,
		"data.transactionId",
		"data.transaction_id",
		"transactionId",
		"transaction_id",
	)
	checkoutURL := extractString(responseBody,
		"data.invoiceUrl",
		"data.invoice_url",
		"data.checkoutUrl",
		"data.checkout_url",
		"invoiceUrl",
		"invoice_url",
		"checkoutUrl",
		"checkout_url",
		"url",
	)

	if invoiceID == "" || transactionID == "" || checkoutURL == "" {
		return billingdomain.Invoice{}, fmt.Errorf(
			"%w: missing required invoice fields",
			billingdomain.ErrProviderUpstream,
		)
	}

	var expiredAt *time.Time
	rawExpiredAt := extractString(responseBody,
		"data.expiredAt",
		"data.expired_at",
		"expiredAt",
		"expired_at",
	)
	if rawExpiredAt != "" {
		parsed, parseErr := parseOptionalRFC3339(rawExpiredAt)
		if parseErr != nil {
			return billingdomain.Invoice{}, fmt.Errorf("%w: invalid expired_at", billingdomain.ErrProviderUpstream)
		}
		expiredAt = parsed
	}

	amount := input.Amount
	if value, ok := extractNumber(responseBody,
		"data.amount",
		"amount",
	); ok && value > 0 {
		amount = value
	}

	return billingdomain.Invoice{
		ID:            invoiceID,
		TransactionID: transactionID,
		CheckoutURL:   checkoutURL,
		Amount:        amount,
		ExpiresAt:     expiredAt,
	}, nil
}

// ValidateCoupon validates coupon code.
func (c *Client) ValidateCoupon(
	ctx context.Context,
	input billingdomain.ValidateCouponInput,
) (billingdomain.Coupon, error) {
	couponCode := strings.ToUpper(strings.TrimSpace(input.Code))
	if couponCode == "" {
		return billingdomain.Coupon{}, billingdomain.ErrCouponInvalid
	}
	if input.Amount <= 0 {
		return billingdomain.Coupon{}, fmt.Errorf("%w: coupon validation amount must be > 0", billingdomain.ErrProviderUpstream)
	}

	queryValues := url.Values{}
	queryValues.Set("code", couponCode)
	queryValues.Set("coupon_code", couponCode)
	queryValues.Set("amount", strconv.FormatInt(input.Amount, 10))
	responseBody, err := c.getCouponValidation(ctx, "/coupon/validate?"+queryValues.Encode())
	if err != nil {
		return billingdomain.Coupon{}, err
	}

	isValid, hasValidFlag := extractBool(
		responseBody,
		"data.valid",
		"data.isValid",
		"data.is_valid",
		"valid",
		"isValid",
		"is_valid",
	)
	if hasValidFlag && !isValid {
		return billingdomain.Coupon{}, billingdomain.ErrCouponInvalid
	}

	discountAmount, hasDiscount := extractNonNegativeNumber(
		responseBody,
		"data.discount_amount",
		"data.discountAmount",
		"data.discount",
		"discount_amount",
		"discountAmount",
		"discount",
	)
	finalAmount, hasFinal := extractNonNegativeNumber(
		responseBody,
		"data.final_amount",
		"data.finalAmount",
		"final_amount",
		"finalAmount",
		"data.amount_after_discount",
		"data.amountAfterDiscount",
	)

	if !hasValidFlag && !hasDiscount && !hasFinal {
		return billingdomain.Coupon{}, billingdomain.ErrCouponInvalid
	}
	if hasFinal {
		if finalAmount > input.Amount {
			return billingdomain.Coupon{}, fmt.Errorf("%w: coupon final amount exceeds plan amount", billingdomain.ErrProviderUpstream)
		}
		if !hasDiscount {
			discountAmount = input.Amount - finalAmount
			hasDiscount = true
		}
	}
	if !hasFinal {
		if hasDiscount {
			if discountAmount >= input.Amount {
				return billingdomain.Coupon{}, fmt.Errorf("%w: coupon discount amount out of range", billingdomain.ErrProviderUpstream)
			}
			finalAmount = input.Amount - discountAmount
		} else {
			finalAmount = input.Amount
		}
	}
	if !hasDiscount {
		discountAmount = 0
	}
	if discountAmount < 0 || discountAmount >= input.Amount {
		return billingdomain.Coupon{}, fmt.Errorf("%w: coupon discount amount out of range", billingdomain.ErrProviderUpstream)
	}

	normalizedCode := strings.ToUpper(extractString(
		responseBody,
		"data.code",
		"data.coupon_code",
		"data.couponCode",
		"code",
		"coupon_code",
		"couponCode",
	))
	if normalizedCode == "" {
		normalizedCode = couponCode
	}

	return billingdomain.Coupon{
		Code:           normalizedCode,
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
	}, nil
}

// GetInvoiceByID returns invoice by id.
func (c *Client) GetInvoiceByID(
	ctx context.Context,
	invoiceID string,
) (billingdomain.InvoiceSnapshot, error) {
	trimmedInvoiceID := strings.TrimSpace(invoiceID)
	if trimmedInvoiceID == "" {
		return billingdomain.InvoiceSnapshot{}, fmt.Errorf("%w: invoice id is required", billingdomain.ErrProviderUpstream)
	}

	responseBody, err := c.get(ctx, "/invoice/"+url.PathEscape(trimmedInvoiceID))
	if err != nil {
		return billingdomain.InvoiceSnapshot{}, err
	}

	transactionID := extractString(responseBody,
		"data.transactionId",
		"data.transaction_id",
		"transactionId",
		"transaction_id",
	)
	transactionStatus := extractString(responseBody,
		"data.transactionStatus",
		"data.transaction_status",
		"data.status",
		"transactionStatus",
		"transaction_status",
		"status",
	)
	customerEmail := extractString(responseBody,
		"data.customerEmail",
		"data.customer_email",
		"customerEmail",
		"customer_email",
	)
	parsedInvoiceID := extractString(responseBody,
		"data.id",
		"data.invoice.id",
		"data.invoice_id",
		"data.invoiceId",
		"id",
	)
	if parsedInvoiceID == "" {
		parsedInvoiceID = trimmedInvoiceID
	}

	amount := int64(0)
	if value, ok := extractNumber(responseBody, "data.amount", "amount"); ok {
		amount = value
	}

	var updatedAt *time.Time
	rawUpdatedAt := extractString(responseBody,
		"data.updatedAt",
		"data.updated_at",
		"updatedAt",
		"updated_at",
	)
	if rawUpdatedAt != "" {
		parsedUpdatedAt, parseErr := parseOptionalRFC3339(rawUpdatedAt)
		if parseErr != nil {
			return billingdomain.InvoiceSnapshot{}, fmt.Errorf("%w: invalid updated_at", billingdomain.ErrProviderUpstream)
		}
		updatedAt = parsedUpdatedAt
	}

	return billingdomain.InvoiceSnapshot{
		InvoiceID:         parsedInvoiceID,
		TransactionID:     transactionID,
		TransactionStatus: strings.ToLower(strings.TrimSpace(transactionStatus)),
		CustomerEmail:     strings.ToLower(strings.TrimSpace(customerEmail)),
		Amount:            amount,
		UpdatedAt:         updatedAt,
	}, nil
}

func (c *Client) post(
	ctx context.Context,
	path string,
	payload map[string]any,
) (map[string]any, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("%w: mayar api key is empty", billingdomain.ErrProviderUnavailable)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: encode request payload: %v", billingdomain.ErrProviderUpstream, err)
	}

	endpoint := c.baseURL + path
	for attempt := 0; ; attempt++ {
		request, reqErr := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			endpoint,
			bytes.NewReader(body),
		)
		if reqErr != nil {
			return nil, fmt.Errorf("%w: build request: %v", billingdomain.ErrProviderUpstream, reqErr)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+c.apiKey)

		response, doErr := c.httpClient.Do(request)
		if doErr != nil {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("%w: request failed: %v", billingdomain.ErrProviderUnavailable, doErr)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderUnavailable)
			}
			continue
		}

		responseBody, readErr := readResponseBody(response.Body)
		if readErr != nil {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("%w: read response body: %v", billingdomain.ErrProviderUnavailable, readErr)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderUnavailable)
			}
			continue
		}

		if response.StatusCode == http.StatusTooManyRequests {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("%w: mayar returned 429", billingdomain.ErrProviderRateLimited)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderRateLimited)
			}
			continue
		}
		if response.StatusCode >= http.StatusInternalServerError {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf(
					"%w: mayar returned status %d",
					billingdomain.ErrProviderUnavailable,
					response.StatusCode,
				)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderUnavailable)
			}
			continue
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf(
				"%w: mayar returned status %d",
				billingdomain.ErrProviderUpstream,
				response.StatusCode,
			)
		}

		decoded := map[string]any{}
		if len(responseBody) > 0 {
			if decodeErr := json.Unmarshal(responseBody, &decoded); decodeErr != nil {
				return nil, fmt.Errorf("%w: invalid response JSON", billingdomain.ErrProviderUpstream)
			}
		}
		return decoded, nil
	}
}

func (c *Client) get(ctx context.Context, path string) (map[string]any, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("%w: mayar api key is empty", billingdomain.ErrProviderUnavailable)
	}

	endpoint := c.baseURL + path
	for attempt := 0; ; attempt++ {
		request, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("%w: build request: %v", billingdomain.ErrProviderUpstream, reqErr)
		}
		request.Header.Set("Authorization", "Bearer "+c.apiKey)

		response, doErr := c.httpClient.Do(request)
		if doErr != nil {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("%w: request failed: %v", billingdomain.ErrProviderUnavailable, doErr)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderUnavailable)
			}
			continue
		}

		responseBody, readErr := readResponseBody(response.Body)
		if readErr != nil {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("%w: read response body: %v", billingdomain.ErrProviderUnavailable, readErr)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderUnavailable)
			}
			continue
		}

		if response.StatusCode == http.StatusTooManyRequests {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("%w: mayar returned 429", billingdomain.ErrProviderRateLimited)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderRateLimited)
			}
			continue
		}
		if response.StatusCode >= http.StatusInternalServerError {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf(
					"%w: mayar returned status %d",
					billingdomain.ErrProviderUnavailable,
					response.StatusCode,
				)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderUnavailable)
			}
			continue
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf(
				"%w: mayar returned status %d",
				billingdomain.ErrProviderUpstream,
				response.StatusCode,
			)
		}

		decoded := map[string]any{}
		if len(responseBody) > 0 {
			if decodeErr := json.Unmarshal(responseBody, &decoded); decodeErr != nil {
				return nil, fmt.Errorf("%w: invalid response JSON", billingdomain.ErrProviderUpstream)
			}
		}
		return decoded, nil
	}
}

func (c *Client) getCouponValidation(ctx context.Context, path string) (map[string]any, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("%w: mayar api key is empty", billingdomain.ErrProviderUnavailable)
	}

	endpoint := c.baseURL + path
	for attempt := 0; ; attempt++ {
		request, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("%w: build request: %v", billingdomain.ErrProviderUpstream, reqErr)
		}
		request.Header.Set("Authorization", "Bearer "+c.apiKey)

		response, doErr := c.httpClient.Do(request)
		if doErr != nil {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("%w: request failed: %v", billingdomain.ErrProviderUnavailable, doErr)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderUnavailable)
			}
			continue
		}

		responseBody, readErr := readResponseBody(response.Body)
		if readErr != nil {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("%w: read response body: %v", billingdomain.ErrProviderUnavailable, readErr)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderUnavailable)
			}
			continue
		}

		if response.StatusCode == http.StatusTooManyRequests {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf("%w: mayar returned 429", billingdomain.ErrProviderRateLimited)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderRateLimited)
			}
			continue
		}
		if response.StatusCode >= http.StatusInternalServerError {
			if attempt >= c.maxRetries {
				return nil, fmt.Errorf(
					"%w: mayar returned status %d",
					billingdomain.ErrProviderUnavailable,
					response.StatusCode,
				)
			}
			if !c.waitForRetry(attempt + 1) {
				return nil, fmt.Errorf("%w: request canceled while retrying", billingdomain.ErrProviderUnavailable)
			}
			continue
		}
		if response.StatusCode == http.StatusBadRequest || response.StatusCode == http.StatusNotFound {
			return nil, billingdomain.ErrCouponInvalid
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf(
				"%w: mayar returned status %d",
				billingdomain.ErrProviderUpstream,
				response.StatusCode,
			)
		}

		decoded := map[string]any{}
		if len(responseBody) > 0 {
			if decodeErr := json.Unmarshal(responseBody, &decoded); decodeErr != nil {
				return nil, fmt.Errorf("%w: invalid response JSON", billingdomain.ErrProviderUpstream)
			}
		}
		return decoded, nil
	}
}

func (c *Client) waitForRetry(retryNumber int) bool {
	delay := retryDelay(retryNumber, c.randIntn)
	if delay <= 0 {
		return true
	}
	c.sleep(delay)
	return true
}

func retryDelay(retryNumber int, randIntn func(int) int) time.Duration {
	backoffSchedule := []time.Duration{
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	index := retryNumber - 1
	if index < 0 {
		index = 0
	}
	if index >= len(backoffSchedule) {
		index = len(backoffSchedule) - 1
	}

	jitter := 0
	if randIntn != nil {
		jitter = randIntn(100)
	}
	return backoffSchedule[index] + (time.Duration(jitter) * time.Millisecond)
}

func readResponseBody(body io.ReadCloser) ([]byte, error) {
	defer func() {
		_ = body.Close()
	}()
	return io.ReadAll(body)
}

func extractString(payload map[string]any, paths ...string) string {
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

func extractNumber(payload map[string]any, paths ...string) (int64, bool) {
	for _, path := range paths {
		value, ok := lookupPath(payload, path)
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			if typed <= 0 {
				return 0, false
			}
			return int64(typed), true
		case int64:
			if typed <= 0 {
				return 0, false
			}
			return typed, true
		case int:
			if typed <= 0 {
				return 0, false
			}
			return int64(typed), true
		case json.Number:
			parsed, err := typed.Int64()
			if err == nil && parsed > 0 {
				return parsed, true
			}
		}
	}
	return 0, false
}

func extractNonNegativeNumber(payload map[string]any, paths ...string) (int64, bool) {
	for _, path := range paths {
		value, ok := lookupPath(payload, path)
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			if typed < 0 {
				return 0, false
			}
			return int64(typed), true
		case int64:
			if typed < 0 {
				return 0, false
			}
			return typed, true
		case int:
			if typed < 0 {
				return 0, false
			}
			return int64(typed), true
		case json.Number:
			parsed, err := typed.Int64()
			if err == nil && parsed >= 0 {
				return parsed, true
			}
		case string:
			parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
			if err == nil && parsed >= 0 {
				return parsed, true
			}
		}
	}
	return 0, false
}

func extractBool(payload map[string]any, paths ...string) (bool, bool) {
	for _, path := range paths {
		value, ok := lookupPath(payload, path)
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case bool:
			return typed, true
		case string:
			normalized := strings.TrimSpace(strings.ToLower(typed))
			switch normalized {
			case "true", "1", "yes", "valid", "applied":
				return true, true
			case "false", "0", "no", "invalid", "not_found", "not-found":
				return false, true
			}
		case float64:
			if typed == 1 {
				return true, true
			}
			if typed == 0 {
				return false, true
			}
		case int:
			if typed == 1 {
				return true, true
			}
			if typed == 0 {
				return false, true
			}
		}
	}
	return false, false
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

func parseOptionalRFC3339(raw string) (*time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, err
	}
	utc := parsed.UTC()
	return &utc, nil
}

var _ billingdomain.Provider = (*Client)(nil)
var _ billingdomain.ReconciliationProvider = (*Client)(nil)
var _ billingdomain.CouponValidator = (*Client)(nil)
