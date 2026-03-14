package billing

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

var (
	ErrInvalidPlanCode    = errors.New("invalid plan code")
	ErrInvalidCouponCode  = errors.New("invalid coupon code")
	ErrInvalidRedirectURL = errors.New("invalid redirect url")
	ErrAlreadyPremium     = errors.New("already premium")
	ErrTooManyRequests    = errors.New("too many requests")
	ErrMayarRateLimited   = errors.New("mayar rate limited")
	ErrMayarUpstream      = errors.New("mayar upstream error")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// Config stores configuration values for config.
type Config struct {
	RedirectAllowlist []string
	IdempotencyWindow time.Duration
	RateLimitWindow   time.Duration
}

// Service coordinates application use cases for the package.
type Service struct {
	identityRepository identity.Repository
	repository         billingdomain.Repository
	provider           billingdomain.Provider
	redirectHosts      map[string]struct{}
	idempotencyWindow  time.Duration
	rateLimitWindow    time.Duration
	now                func() time.Time

	rateLimitMu       sync.Mutex
	lastCheckoutByUID map[string]time.Time
}

// CreateCheckoutSessionInput contains input parameters for create checkout session.
type CreateCheckoutSessionInput struct {
	UserID         string
	PlanCode       string
	CouponCode     string
	RedirectURL    string
	IdempotencyKey string
}

// CheckoutSession represents checkout session.
type CheckoutSession struct {
	Provider          billingdomain.PaymentProvider
	PlanCode          billingdomain.PlanCode
	InvoiceID         string
	TransactionID     string
	CheckoutURL       string
	OriginalAmount    int64
	DiscountAmount    int64
	FinalAmount       int64
	CouponCode        string
	ExpiredAt         time.Time
	SubscriptionState identity.SubscriptionState
	TransactionStatus billingdomain.TransactionStatus
	Reused            bool
}

type planDefinition struct {
	Code        billingdomain.PlanCode
	Amount      int64
	Name        string
	Description string
}

var supportedPlans = map[billingdomain.PlanCode]planDefinition{
	billingdomain.PlanCodeProMonthly: {
		Code:        billingdomain.PlanCodeProMonthly,
		Amount:      49_000,
		Name:        "Bisakerja Pro Monthly",
		Description: "Bisakerja Pro subscription (monthly)",
	},
}

// NewService creates a new service instance.
func NewService(
	identityRepository identity.Repository,
	repository billingdomain.Repository,
	provider billingdomain.Provider,
	config Config,
) *Service {
	idempotencyWindow := config.IdempotencyWindow
	if idempotencyWindow <= 0 {
		idempotencyWindow = 15 * time.Minute
	}

	rateLimitWindow := config.RateLimitWindow
	if rateLimitWindow <= 0 {
		rateLimitWindow = 10 * time.Second
	}

	return &Service{
		identityRepository: identityRepository,
		repository:         repository,
		provider:           provider,
		redirectHosts:      normalizeAllowedHosts(config.RedirectAllowlist),
		idempotencyWindow:  idempotencyWindow,
		rateLimitWindow:    rateLimitWindow,
		now:                func() time.Time { return time.Now().UTC() },
		lastCheckoutByUID:  make(map[string]time.Time),
	}
}

// CreateCheckoutSession creates checkout session.
func (s *Service) CreateCheckoutSession(
	ctx context.Context,
	input CreateCheckoutSessionInput,
) (CheckoutSession, error) {
	if s.identityRepository == nil || s.repository == nil || s.provider == nil {
		return CheckoutSession{}, errors.New("billing service dependency is not fully configured")
	}

	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return CheckoutSession{}, identity.ErrUserNotFound
	}

	planCode := billingdomain.PlanCode(strings.TrimSpace(input.PlanCode))
	plan, ok := supportedPlans[planCode]
	if !ok {
		return CheckoutSession{}, ErrInvalidPlanCode
	}
	couponCode := normalizeCouponCode(input.CouponCode)
	if couponCode != "" && !isCouponCodeValid(couponCode) {
		return CheckoutSession{}, ErrInvalidCouponCode
	}

	redirectURL := strings.TrimSpace(input.RedirectURL)
	if !isRedirectURLAllowed(redirectURL, s.redirectHosts) {
		return CheckoutSession{}, ErrInvalidRedirectURL
	}

	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	now := s.now().UTC()
	if idempotencyKey != "" {
		existing, err := s.repository.FindPendingByUserAndIdempotencyKey(
			ctx,
			userID,
			idempotencyKey,
			s.idempotencyWindow,
			now,
		)
		if err == nil {
			return mapTransactionToCheckout(existing, now, true), nil
		}
		if !errors.Is(err, billingdomain.ErrTransactionNotFound) {
			return CheckoutSession{}, fmt.Errorf("find pending transaction by idempotency key: %w", err)
		}
	}

	user, err := s.identityRepository.GetUserByID(ctx, userID)
	if err != nil {
		return CheckoutSession{}, fmt.Errorf("get user profile: %w", err)
	}
	if isPremiumActive(user, now) {
		return CheckoutSession{}, ErrAlreadyPremium
	}

	if !s.allowCheckout(userID, now) {
		return CheckoutSession{}, ErrTooManyRequests
	}

	discountAmount := int64(0)
	finalAmount := plan.Amount
	appliedCouponCode := ""
	if couponCode != "" {
		couponValidator, ok := s.provider.(billingdomain.CouponValidator)
		if !ok {
			return CheckoutSession{}, ErrServiceUnavailable
		}
		coupon, couponErr := couponValidator.ValidateCoupon(ctx, billingdomain.ValidateCouponInput{
			Code:   couponCode,
			Amount: plan.Amount,
		})
		if couponErr != nil {
			switch {
			case errors.Is(couponErr, billingdomain.ErrCouponInvalid):
				return CheckoutSession{}, ErrInvalidCouponCode
			default:
				return CheckoutSession{}, mapProviderError("validate coupon", couponErr)
			}
		}
		var normalizeErr error
		discountAmount, finalAmount, normalizeErr = normalizeCouponAmounts(plan.Amount, coupon)
		if normalizeErr != nil {
			return CheckoutSession{}, mapProviderError("validate coupon", normalizeErr)
		}
		appliedCouponCode = normalizeCouponCode(coupon.Code)
		if appliedCouponCode == "" {
			appliedCouponCode = couponCode
		}
	}

	customer, err := s.provider.EnsureCustomer(ctx, billingdomain.EnsureCustomerInput{
		Name:  user.Name,
		Email: user.Email,
	})
	if err != nil {
		return CheckoutSession{}, mapProviderError("ensure customer", err)
	}

	invoice, err := s.provider.CreateInvoice(ctx, billingdomain.CreateInvoiceInput{
		CustomerID:  customer.ID,
		PlanCode:    plan.Code,
		Amount:      finalAmount,
		Description: plan.Description,
		RedirectURL: redirectURL,
		ExternalID:  buildExternalID(userID, idempotencyKey, now),
	})
	if err != nil {
		return CheckoutSession{}, mapProviderError("create invoice", err)
	}

	expiredAt := invoice.ExpiresAt
	if expiredAt == nil {
		defaultExpiry := now.Add(24 * time.Hour)
		expiredAt = &defaultExpiry
	}

	effectiveFinalAmount := invoice.Amount
	if effectiveFinalAmount <= 0 {
		effectiveFinalAmount = finalAmount
	}
	if plan.Amount > effectiveFinalAmount {
		discountAmount = plan.Amount - effectiveFinalAmount
	}

	metadata := map[string]any{
		"redirect_url":    redirectURL,
		"customer_id":     customer.ID,
		"original_amount": plan.Amount,
		"discount_amount": discountAmount,
		"final_amount":    effectiveFinalAmount,
	}
	if appliedCouponCode != "" {
		metadata["coupon_code"] = appliedCouponCode
	}

	created, err := s.repository.CreatePending(ctx, billingdomain.CreatePendingTransactionInput{
		UserID:             userID,
		Provider:           billingdomain.PaymentProviderMayar,
		PlanCode:           plan.Code,
		MayarTransactionID: invoice.TransactionID,
		InvoiceID:          invoice.ID,
		CheckoutURL:        invoice.CheckoutURL,
		Amount:             effectiveFinalAmount,
		IdempotencyKey:     idempotencyKey,
		ExpiresAt:          expiredAt,
		Metadata:           metadata,
	})
	if err != nil {
		return CheckoutSession{}, fmt.Errorf("create pending transaction: %w", err)
	}

	return mapTransactionToCheckout(created, now, false), nil
}

func normalizeAllowedHosts(hosts []string) map[string]struct{} {
	result := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		normalized := strings.ToLower(strings.TrimSpace(host))
		if normalized == "" {
			continue
		}
		result[normalized] = struct{}{}
	}

	return result
}

func isRedirectURLAllowed(rawURL string, allowlist map[string]struct{}) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Host))
	if host == "" {
		return false
	}
	if len(allowlist) == 0 {
		return false
	}
	_, allowed := allowlist[host]
	return allowed
}

func isPremiumActive(user identity.User, now time.Time) bool {
	if !user.IsPremium {
		return false
	}
	if user.PremiumExpiredAt == nil {
		return true
	}
	return user.PremiumExpiredAt.After(now)
}

func (s *Service) allowCheckout(userID string, now time.Time) bool {
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()

	lastRequestAt, exists := s.lastCheckoutByUID[userID]
	if exists && now.Before(lastRequestAt.Add(s.rateLimitWindow)) {
		return false
	}

	s.lastCheckoutByUID[userID] = now
	return true
}

func mapProviderError(operation string, err error) error {
	switch {
	case errors.Is(err, billingdomain.ErrProviderRateLimited):
		return fmt.Errorf("%s: %w", operation, ErrMayarRateLimited)
	case errors.Is(err, billingdomain.ErrProviderUnavailable):
		return fmt.Errorf("%s: %w", operation, ErrServiceUnavailable)
	case errors.Is(err, billingdomain.ErrProviderUpstream):
		return fmt.Errorf("%s: %w", operation, ErrMayarUpstream)
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}

func buildExternalID(userID, idempotencyKey string, now time.Time) string {
	if idempotencyKey != "" {
		return "checkout:" + userID + ":" + idempotencyKey
	}
	return fmt.Sprintf("checkout:%s:%d", userID, now.UnixNano())
}

func mapTransactionToCheckout(
	transaction billingdomain.Transaction,
	now time.Time,
	reused bool,
) CheckoutSession {
	expiredAt := now.Add(24 * time.Hour)
	if transaction.ExpiresAt != nil {
		expiredAt = transaction.ExpiresAt.UTC()
	}

	originalAmount := transaction.Amount
	if value, ok := metadataNumber(transaction.Metadata, "original_amount"); ok && value > 0 {
		originalAmount = value
	}
	finalAmount := transaction.Amount
	if value, ok := metadataNumber(transaction.Metadata, "final_amount"); ok && value > 0 {
		finalAmount = value
	}
	discountAmount := int64(0)
	if value, ok := metadataNumber(transaction.Metadata, "discount_amount"); ok && value >= 0 {
		discountAmount = value
	}
	if discountAmount == 0 && originalAmount > finalAmount {
		discountAmount = originalAmount - finalAmount
	}
	if discountAmount < 0 {
		discountAmount = 0
	}
	couponCode := normalizeCouponCode(metadataString(transaction.Metadata, "coupon_code"))

	return CheckoutSession{
		Provider:          transaction.Provider,
		PlanCode:          transaction.PlanCode,
		InvoiceID:         transaction.InvoiceID,
		TransactionID:     transaction.MayarTransactionID,
		CheckoutURL:       transaction.CheckoutURL,
		OriginalAmount:    originalAmount,
		DiscountAmount:    discountAmount,
		FinalAmount:       finalAmount,
		CouponCode:        couponCode,
		ExpiredAt:         expiredAt,
		SubscriptionState: identity.SubscriptionStatePendingPayment,
		TransactionStatus: transaction.Status,
		Reused:            reused,
	}
}

func normalizeCouponCode(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}

func isCouponCodeValid(code string) bool {
	if len(code) < 3 || len(code) > 64 {
		return false
	}
	for _, character := range code {
		isUpper := character >= 'A' && character <= 'Z'
		isDigit := character >= '0' && character <= '9'
		if isUpper || isDigit || character == '-' || character == '_' {
			continue
		}
		return false
	}
	return true
}

func normalizeCouponAmounts(planAmount int64, coupon billingdomain.Coupon) (int64, int64, error) {
	discountAmount := coupon.DiscountAmount
	finalAmount := coupon.FinalAmount

	if finalAmount > 0 {
		if finalAmount > planAmount {
			return 0, 0, fmt.Errorf("%w: coupon final amount exceeds plan amount", billingdomain.ErrProviderUpstream)
		}
		expectedDiscount := planAmount - finalAmount
		if discountAmount == 0 {
			discountAmount = expectedDiscount
		} else if discountAmount != expectedDiscount {
			return 0, 0, fmt.Errorf("%w: coupon amount mismatch", billingdomain.ErrProviderUpstream)
		}
	}
	if discountAmount < 0 || discountAmount >= planAmount {
		return 0, 0, fmt.Errorf("%w: coupon discount amount out of range", billingdomain.ErrProviderUpstream)
	}
	return discountAmount, planAmount - discountAmount, nil
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	raw, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(raw)
}

func metadataNumber(metadata map[string]any, key string) (int64, bool) {
	if metadata == nil {
		return 0, false
	}
	value, ok := metadata[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	case float64:
		return int64(typed), true
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}
