package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	billingdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/billing"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

type fakeProvider struct {
	ensureCustomerFn func(context.Context, billingdomain.EnsureCustomerInput) (billingdomain.Customer, error)
	createInvoiceFn  func(context.Context, billingdomain.CreateInvoiceInput) (billingdomain.Invoice, error)
	validateCouponFn func(context.Context, billingdomain.ValidateCouponInput) (billingdomain.Coupon, error)
	ensureCalls      int
	invoiceCalls     int
	couponCalls      int
	lastInvoiceInput billingdomain.CreateInvoiceInput
}

func (f *fakeProvider) EnsureCustomer(
	ctx context.Context,
	input billingdomain.EnsureCustomerInput,
) (billingdomain.Customer, error) {
	f.ensureCalls++
	if f.ensureCustomerFn != nil {
		return f.ensureCustomerFn(ctx, input)
	}
	return billingdomain.Customer{ID: "cust_1", Email: input.Email, Name: input.Name}, nil
}

func (f *fakeProvider) CreateInvoice(
	ctx context.Context,
	input billingdomain.CreateInvoiceInput,
) (billingdomain.Invoice, error) {
	f.invoiceCalls++
	f.lastInvoiceInput = input
	if f.createInvoiceFn != nil {
		return f.createInvoiceFn(ctx, input)
	}
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	return billingdomain.Invoice{
		ID:            "inv_1",
		TransactionID: "trx_1",
		CheckoutURL:   "https://pay.example.com/checkout",
		Amount:        input.Amount,
		ExpiresAt:     &expiresAt,
	}, nil
}

func (f *fakeProvider) ValidateCoupon(
	ctx context.Context,
	input billingdomain.ValidateCouponInput,
) (billingdomain.Coupon, error) {
	f.couponCalls++
	if f.validateCouponFn != nil {
		return f.validateCouponFn(ctx, input)
	}
	return billingdomain.Coupon{
		Code:           input.Code,
		DiscountAmount: 0,
		FinalAmount:    input.Amount,
	}, nil
}

func TestService_CreateCheckoutSession_Success(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user := seedUser(t, identityRepository, false, nil)
	transactionRepository := memory.NewBillingRepository()
	provider := &fakeProvider{}

	service := NewService(identityRepository, transactionRepository, provider, Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
		IdempotencyWindow: 15 * time.Minute,
		RateLimitWindow:   10 * time.Second,
	})

	checkout, err := service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:         user.ID,
		PlanCode:       "pro_monthly",
		RedirectURL:    "https://app.bisakerja.com/billing/success",
		IdempotencyKey: "idem-1",
	})
	if err != nil {
		t.Fatalf("create checkout session: %v", err)
	}

	if checkout.Provider != billingdomain.PaymentProviderMayar {
		t.Fatalf("expected provider mayar, got %s", checkout.Provider)
	}
	if checkout.PlanCode != billingdomain.PlanCodeProMonthly {
		t.Fatalf("expected plan code pro_monthly, got %s", checkout.PlanCode)
	}
	if checkout.InvoiceID == "" || checkout.TransactionID == "" || checkout.CheckoutURL == "" {
		t.Fatalf("expected checkout ids and url to be set, got %+v", checkout)
	}
	if checkout.SubscriptionState != identity.SubscriptionStatePendingPayment {
		t.Fatalf("expected pending_payment state, got %s", checkout.SubscriptionState)
	}
	if checkout.TransactionStatus != billingdomain.TransactionStatusPending {
		t.Fatalf("expected transaction status pending, got %s", checkout.TransactionStatus)
	}
	if checkout.OriginalAmount != 49_000 || checkout.DiscountAmount != 0 || checkout.FinalAmount != 49_000 {
		t.Fatalf("unexpected checkout amount details: %+v", checkout)
	}
	if checkout.CouponCode != "" {
		t.Fatalf("expected empty coupon code, got %q", checkout.CouponCode)
	}
	if checkout.Reused {
		t.Fatal("expected first checkout to not be reused")
	}
}

func TestService_CreateCheckoutSession_WithCoupon(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user := seedUser(t, identityRepository, false, nil)
	transactionRepository := memory.NewBillingRepository()
	provider := &fakeProvider{
		validateCouponFn: func(context.Context, billingdomain.ValidateCouponInput) (billingdomain.Coupon, error) {
			return billingdomain.Coupon{
				Code:           "SAVE10",
				DiscountAmount: 10_000,
				FinalAmount:    39_000,
			}, nil
		},
	}

	service := NewService(identityRepository, transactionRepository, provider, Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
		IdempotencyWindow: 15 * time.Minute,
		RateLimitWindow:   10 * time.Second,
	})

	checkout, err := service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:         user.ID,
		PlanCode:       "pro_monthly",
		CouponCode:     "save10",
		RedirectURL:    "https://app.bisakerja.com/billing/success",
		IdempotencyKey: "idem-coupon",
	})
	if err != nil {
		t.Fatalf("create checkout session with coupon: %v", err)
	}
	if provider.lastInvoiceInput.Amount != 39_000 {
		t.Fatalf("expected invoice amount 39000, got %d", provider.lastInvoiceInput.Amount)
	}
	if checkout.OriginalAmount != 49_000 || checkout.DiscountAmount != 10_000 || checkout.FinalAmount != 39_000 {
		t.Fatalf("unexpected checkout amount details: %+v", checkout)
	}
	if checkout.CouponCode != "SAVE10" {
		t.Fatalf("expected coupon code SAVE10, got %q", checkout.CouponCode)
	}
}

func TestService_CreateCheckoutSession_InvalidCoupon(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user := seedUser(t, identityRepository, false, nil)
	transactionRepository := memory.NewBillingRepository()
	provider := &fakeProvider{
		validateCouponFn: func(context.Context, billingdomain.ValidateCouponInput) (billingdomain.Coupon, error) {
			return billingdomain.Coupon{}, billingdomain.ErrCouponInvalid
		},
	}

	service := NewService(identityRepository, transactionRepository, provider, Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
		IdempotencyWindow: 15 * time.Minute,
		RateLimitWindow:   10 * time.Second,
	})

	_, err := service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:      user.ID,
		PlanCode:    "pro_monthly",
		CouponCode:  "bad-code",
		RedirectURL: "https://app.bisakerja.com/billing/success",
	})
	if !errors.Is(err, ErrInvalidCouponCode) {
		t.Fatalf("expected ErrInvalidCouponCode, got %v", err)
	}
}

func TestService_CreateCheckoutSession_IdempotencyReuse(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user := seedUser(t, identityRepository, false, nil)
	transactionRepository := memory.NewBillingRepository()
	provider := &fakeProvider{}

	service := NewService(identityRepository, transactionRepository, provider, Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
		IdempotencyWindow: 15 * time.Minute,
		RateLimitWindow:   10 * time.Second,
	})

	first, err := service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:         user.ID,
		PlanCode:       "pro_monthly",
		RedirectURL:    "https://app.bisakerja.com/billing/success",
		IdempotencyKey: "idem-reuse",
	})
	if err != nil {
		t.Fatalf("first create checkout: %v", err)
	}
	second, err := service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:         user.ID,
		PlanCode:       "pro_monthly",
		RedirectURL:    "https://app.bisakerja.com/billing/success",
		IdempotencyKey: "idem-reuse",
	})
	if err != nil {
		t.Fatalf("second create checkout: %v", err)
	}

	if !second.Reused {
		t.Fatal("expected second checkout request to be reused")
	}
	if first.TransactionID != second.TransactionID {
		t.Fatalf("expected reused transaction_id %q, got %q", first.TransactionID, second.TransactionID)
	}
	if provider.ensureCalls != 1 || provider.invoiceCalls != 1 {
		t.Fatalf("expected provider called once for idempotent replay, got ensure=%d invoice=%d", provider.ensureCalls, provider.invoiceCalls)
	}
}

func TestService_CreateCheckoutSession_RateLimited(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user := seedUser(t, identityRepository, false, nil)
	transactionRepository := memory.NewBillingRepository()
	provider := &fakeProvider{}

	service := NewService(identityRepository, transactionRepository, provider, Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
		RateLimitWindow:   10 * time.Second,
	})

	_, err := service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:      user.ID,
		PlanCode:    "pro_monthly",
		RedirectURL: "https://app.bisakerja.com/billing/success",
	})
	if err != nil {
		t.Fatalf("first create checkout: %v", err)
	}

	_, err = service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:      user.ID,
		PlanCode:    "pro_monthly",
		RedirectURL: "https://app.bisakerja.com/billing/success",
	})
	if !errors.Is(err, ErrTooManyRequests) {
		t.Fatalf("expected ErrTooManyRequests, got %v", err)
	}
}

func TestService_CreateCheckoutSession_ValidationAndState(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	activePremiumUntil := time.Now().UTC().Add(48 * time.Hour)
	premiumUser := seedUser(t, identityRepository, true, &activePremiumUntil)
	normalUser := seedUser(t, identityRepository, false, nil)
	transactionRepository := memory.NewBillingRepository()
	provider := &fakeProvider{}

	service := NewService(identityRepository, transactionRepository, provider, Config{
		RedirectAllowlist: []string{"app.bisakerja.com"},
	})

	_, err := service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:      normalUser.ID,
		PlanCode:    "invalid_plan",
		RedirectURL: "https://app.bisakerja.com/billing/success",
	})
	if !errors.Is(err, ErrInvalidPlanCode) {
		t.Fatalf("expected ErrInvalidPlanCode, got %v", err)
	}

	_, err = service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:      normalUser.ID,
		PlanCode:    "pro_monthly",
		RedirectURL: "http://app.bisakerja.com/billing/success",
	})
	if !errors.Is(err, ErrInvalidRedirectURL) {
		t.Fatalf("expected ErrInvalidRedirectURL, got %v", err)
	}

	_, err = service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
		UserID:      premiumUser.ID,
		PlanCode:    "pro_monthly",
		RedirectURL: "https://app.bisakerja.com/billing/success",
	})
	if !errors.Is(err, ErrAlreadyPremium) {
		t.Fatalf("expected ErrAlreadyPremium, got %v", err)
	}
}

func TestService_CreateCheckoutSession_ProviderErrorMapping(t *testing.T) {
	tests := []struct {
		name        string
		providerErr error
		expectedErr error
	}{
		{
			name:        "rate limited",
			providerErr: billingdomain.ErrProviderRateLimited,
			expectedErr: ErrMayarRateLimited,
		},
		{
			name:        "upstream invalid",
			providerErr: billingdomain.ErrProviderUpstream,
			expectedErr: ErrMayarUpstream,
		},
		{
			name:        "unavailable",
			providerErr: billingdomain.ErrProviderUnavailable,
			expectedErr: ErrServiceUnavailable,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			identityRepository := memory.NewIdentityRepository()
			user := seedUser(t, identityRepository, false, nil)
			transactionRepository := memory.NewBillingRepository()
			provider := &fakeProvider{
				ensureCustomerFn: func(context.Context, billingdomain.EnsureCustomerInput) (billingdomain.Customer, error) {
					return billingdomain.Customer{}, testCase.providerErr
				},
			}

			service := NewService(identityRepository, transactionRepository, provider, Config{
				RedirectAllowlist: []string{"app.bisakerja.com"},
			})

			_, err := service.CreateCheckoutSession(context.Background(), CreateCheckoutSessionInput{
				UserID:      user.ID,
				PlanCode:    "pro_monthly",
				RedirectURL: "https://app.bisakerja.com/billing/success",
			})
			if !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("expected %v, got %v", testCase.expectedErr, err)
			}
		})
	}
}

func seedUser(
	t *testing.T,
	identityRepository *memory.IdentityRepository,
	isPremium bool,
	premiumExpiredAt *time.Time,
) identity.User {
	t.Helper()

	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:            "user+" + time.Now().UTC().Format("20060102150405.000000000") + "@example.com",
		PasswordHash:     "hashed-password",
		Name:             "Billing User",
		Role:             identity.RoleUser,
		IsPremium:        isPremium,
		PremiumExpiredAt: premiumExpiredAt,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user
}
