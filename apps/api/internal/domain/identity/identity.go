package identity

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Role represents role.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// SubscriptionState represents subscription state.
type SubscriptionState string

const (
	SubscriptionStateFree           SubscriptionState = "free"
	SubscriptionStatePendingPayment SubscriptionState = "pending_payment"
	SubscriptionStatePremiumActive  SubscriptionState = "premium_active"
	SubscriptionStatePremiumExpired SubscriptionState = "premium_expired"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrForbidden              = errors.New("forbidden")
)

// User represents user.
type User struct {
	ID               string
	Email            string
	PasswordHash     string
	Name             string
	Role             Role
	IsPremium        bool
	PremiumExpiredAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Preferences represents preferences.
type Preferences struct {
	UserID     string
	Keywords   []string
	Locations  []string
	JobTypes   []string
	SalaryMin  int64
	AlertMode  NotificationAlertMode
	DigestHour *int
	UpdatedAt  *time.Time
}

// NotificationAlertMode represents notification alert mode.
type NotificationAlertMode string

const (
	NotificationAlertModeInstant      NotificationAlertMode = "instant"
	NotificationAlertModeDailyDigest  NotificationAlertMode = "daily_digest"
	NotificationAlertModeWeeklyDigest NotificationAlertMode = "weekly_digest"
)

// CreateUserInput contains input parameters for create user.
type CreateUserInput struct {
	Email            string
	PasswordHash     string
	Name             string
	Role             Role
	IsPremium        bool
	PremiumExpiredAt *time.Time
}

// Repository defines behavior for repository.
type Repository interface {
	CreateUser(ctx context.Context, input CreateUserInput) (User, error)
	GetUserByID(ctx context.Context, userID string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	UpdatePremiumStatus(ctx context.Context, userID string, isPremium bool, premiumExpiredAt *time.Time) (User, error)
	ListUsers(ctx context.Context) ([]User, error)
	GetPreferences(ctx context.Context, userID string) (Preferences, error)
	SavePreferences(ctx context.Context, preferences Preferences) (Preferences, error)
}

// NormalizeEmail handles normalize email.
func NormalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}
