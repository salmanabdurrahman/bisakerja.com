package identity

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	domain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	platformauth "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
)

var (
	ErrInvalidEmail         = errors.New("email is invalid")
	ErrInvalidPassword      = errors.New("password is invalid")
	ErrInvalidName          = errors.New("name is invalid")
	ErrInvalidToken         = errors.New("token is invalid")
	ErrKeywordsRequired     = errors.New("keywords is required")
	ErrInvalidKeyword       = errors.New("keywords contains invalid value")
	ErrLocationsRequired    = errors.New("locations is required")
	ErrInvalidLocation      = errors.New("locations contains invalid value")
	ErrJobTypesRequired     = errors.New("job_types is required")
	ErrInvalidJobType       = errors.New("job_types contains invalid value")
	ErrInvalidSalaryMin     = errors.New("salary_min is invalid")
	ErrInvalidAlertMode     = errors.New("alert_mode is invalid")
	ErrInvalidDigestHour    = errors.New("digest_hour is invalid")
	ErrPreferencesUserEmpty = errors.New("user id is required")
)

var allowedJobTypes = map[string]struct{}{
	"fulltime":   {},
	"parttime":   {},
	"contract":   {},
	"internship": {},
}

// TokenManager defines behavior for token manager.
type TokenManager interface {
	IssueAccessToken(userID string, role domain.Role) (string, error)
	IssueRefreshToken(userID string, role domain.Role) (string, error)
	ParseRefreshToken(rawToken string) (platformauth.Claims, error)
	AccessTokenTTLSeconds() int
}

// Service coordinates application use cases for the package.
type Service struct {
	repository   domain.Repository
	tokenManager TokenManager
	now          func() time.Time
}

// RegisterInput contains input parameters for register.
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// LoginInput contains input parameters for login.
type LoginInput struct {
	Email    string
	Password string
}

// Profile represents profile.
type Profile struct {
	ID                string
	Email             string
	Name              string
	Role              domain.Role
	IsPremium         bool
	PremiumExpiredAt  *time.Time
	SubscriptionState domain.SubscriptionState
}

// AuthTokens represents auth tokens.
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
}

// AccessToken represents access token.
type AccessToken struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
}

// UpdatePreferencesInput contains input parameters for update preferences.
type UpdatePreferencesInput struct {
	Keywords     []string
	KeywordsSet  bool
	Locations    []string
	LocationsSet bool
	JobTypes     []string
	JobTypesSet  bool
	SalaryMin    int64
	SalaryMinSet bool
}

// UpdateNotificationPreferencesInput contains input parameters for update notification preferences.
type UpdateNotificationPreferencesInput struct {
	AlertMode     string
	AlertModeSet  bool
	DigestHour    int
	DigestHourSet bool
}

// NewService creates a new service instance.
func NewService(repository domain.Repository, tokenManager TokenManager) *Service {
	return &Service{
		repository:   repository,
		tokenManager: tokenManager,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

// Register handles register.
func (s *Service) Register(ctx context.Context, input RegisterInput) (domain.User, error) {
	email := domain.NormalizeEmail(input.Email)
	if !isValidEmail(email) {
		return domain.User{}, ErrInvalidEmail
	}

	if !isValidPassword(input.Password) {
		return domain.User{}, ErrInvalidPassword
	}

	name := strings.TrimSpace(input.Name)
	if len(name) < 2 || len(name) > 100 {
		return domain.User{}, ErrInvalidName
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repository.CreateUser(ctx, domain.CreateUserInput{
		Email:        email,
		PasswordHash: string(passwordHash),
		Name:         name,
		Role:         domain.RoleUser,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// Login handles login.
func (s *Service) Login(ctx context.Context, input LoginInput) (AuthTokens, error) {
	email := domain.NormalizeEmail(input.Email)
	if !isValidEmail(email) {
		return AuthTokens{}, ErrInvalidEmail
	}
	if strings.TrimSpace(input.Password) == "" {
		return AuthTokens{}, ErrInvalidPassword
	}

	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return AuthTokens{}, domain.ErrInvalidCredentials
		}
		return AuthTokens{}, fmt.Errorf("get user by email: %w", err)
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); compareErr != nil {
		return AuthTokens{}, domain.ErrInvalidCredentials
	}

	accessToken, err := s.tokenManager.IssueAccessToken(user.ID, user.Role)
	if err != nil {
		return AuthTokens{}, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err := s.tokenManager.IssueRefreshToken(user.ID, user.Role)
	if err != nil {
		return AuthTokens{}, fmt.Errorf("issue refresh token: %w", err)
	}

	return AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.tokenManager.AccessTokenTTLSeconds(),
	}, nil
}

// Refresh handles refresh.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (AccessToken, error) {
	claims, err := s.tokenManager.ParseRefreshToken(strings.TrimSpace(refreshToken))
	if err != nil {
		if errors.Is(err, platformauth.ErrInvalidToken) || errors.Is(err, platformauth.ErrTokenExpired) {
			return AccessToken{}, ErrInvalidToken
		}
		return AccessToken{}, fmt.Errorf("parse refresh token: %w", err)
	}

	user, err := s.repository.GetUserByID(ctx, claims.Subject)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return AccessToken{}, ErrInvalidToken
		}
		return AccessToken{}, fmt.Errorf("get user by id: %w", err)
	}

	accessToken, err := s.tokenManager.IssueAccessToken(user.ID, user.Role)
	if err != nil {
		return AccessToken{}, fmt.Errorf("issue access token: %w", err)
	}

	return AccessToken{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   s.tokenManager.AccessTokenTTLSeconds(),
	}, nil
}

// GetProfile returns profile.
func (s *Service) GetProfile(ctx context.Context, userID string) (Profile, error) {
	user, err := s.repository.GetUserByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return Profile{}, fmt.Errorf("get user profile: %w", err)
	}

	return Profile{
		ID:                user.ID,
		Email:             user.Email,
		Name:              user.Name,
		Role:              user.Role,
		IsPremium:         user.IsPremium,
		PremiumExpiredAt:  user.PremiumExpiredAt,
		SubscriptionState: subscriptionStateForUser(user, s.now()),
	}, nil
}

// GetPreferences returns preferences.
func (s *Service) GetPreferences(ctx context.Context, userID string) (domain.Preferences, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return domain.Preferences{}, ErrPreferencesUserEmpty
	}

	preferences, err := s.repository.GetPreferences(ctx, trimmedUserID)
	if err != nil {
		return domain.Preferences{}, fmt.Errorf("get preferences: %w", err)
	}
	return preferences, nil
}

// UpdatePreferences updates preferences.
func (s *Service) UpdatePreferences(ctx context.Context, userID string, input UpdatePreferencesInput) (domain.Preferences, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return domain.Preferences{}, ErrPreferencesUserEmpty
	}
	if !input.KeywordsSet {
		return domain.Preferences{}, ErrKeywordsRequired
	}
	if !input.LocationsSet {
		return domain.Preferences{}, ErrLocationsRequired
	}
	if !input.JobTypesSet {
		return domain.Preferences{}, ErrJobTypesRequired
	}

	current, err := s.repository.GetPreferences(ctx, trimmedUserID)
	if err != nil {
		return domain.Preferences{}, fmt.Errorf("get existing preferences: %w", err)
	}

	keywords, err := normalizeStringList(input.Keywords, 1, 10, 2, 50)
	if err != nil {
		return domain.Preferences{}, ErrInvalidKeyword
	}

	locations, locationErr := normalizeStringList(input.Locations, 1, 5, 2, 100)
	if locationErr != nil {
		return domain.Preferences{}, ErrInvalidLocation
	}

	jobTypes, jobTypeErr := normalizeJobTypes(input.JobTypes)
	if jobTypeErr != nil {
		return domain.Preferences{}, ErrInvalidJobType
	}

	salaryMin := current.SalaryMin
	if input.SalaryMinSet {
		if input.SalaryMin < 0 || input.SalaryMin > 999_000_000 {
			return domain.Preferences{}, ErrInvalidSalaryMin
		}
		salaryMin = input.SalaryMin
	}

	updatedAt := s.now()
	updated, err := s.repository.SavePreferences(ctx, domain.Preferences{
		UserID:     trimmedUserID,
		Keywords:   keywords,
		Locations:  locations,
		JobTypes:   jobTypes,
		SalaryMin:  salaryMin,
		AlertMode:  normalizeAlertModeOrDefault(string(current.AlertMode)),
		DigestHour: cloneInt(current.DigestHour),
		UpdatedAt:  &updatedAt,
	})
	if err != nil {
		return domain.Preferences{}, fmt.Errorf("save preferences: %w", err)
	}

	return updated, nil
}

// UpdateNotificationPreferences updates notification preferences.
func (s *Service) UpdateNotificationPreferences(
	ctx context.Context,
	userID string,
	input UpdateNotificationPreferencesInput,
) (domain.Preferences, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return domain.Preferences{}, ErrPreferencesUserEmpty
	}

	current, err := s.repository.GetPreferences(ctx, trimmedUserID)
	if err != nil {
		return domain.Preferences{}, fmt.Errorf("get existing preferences: %w", err)
	}

	targetMode := normalizeAlertModeOrDefault(string(current.AlertMode))
	if input.AlertModeSet {
		parsedMode, ok := parseAlertMode(input.AlertMode)
		if !ok {
			return domain.Preferences{}, ErrInvalidAlertMode
		}
		targetMode = parsedMode
	}

	var targetDigestHour *int
	switch targetMode {
	case domain.NotificationAlertModeInstant:
		if input.DigestHourSet {
			return domain.Preferences{}, ErrInvalidDigestHour
		}
		targetDigestHour = nil
	case domain.NotificationAlertModeDailyDigest, domain.NotificationAlertModeWeeklyDigest:
		if input.DigestHourSet {
			if input.DigestHour < 0 || input.DigestHour > 23 {
				return domain.Preferences{}, ErrInvalidDigestHour
			}
			digestHour := input.DigestHour
			targetDigestHour = &digestHour
		} else if current.DigestHour != nil {
			targetDigestHour = cloneInt(current.DigestHour)
		} else {
			defaultDigestHour := 9
			targetDigestHour = &defaultDigestHour
		}
	default:
		return domain.Preferences{}, ErrInvalidAlertMode
	}

	updatedAt := s.now()
	updated, saveErr := s.repository.SavePreferences(ctx, domain.Preferences{
		UserID:     trimmedUserID,
		Keywords:   append([]string(nil), current.Keywords...),
		Locations:  append([]string(nil), current.Locations...),
		JobTypes:   append([]string(nil), current.JobTypes...),
		SalaryMin:  current.SalaryMin,
		AlertMode:  targetMode,
		DigestHour: targetDigestHour,
		UpdatedAt:  &updatedAt,
	})
	if saveErr != nil {
		return domain.Preferences{}, fmt.Errorf("save notification preferences: %w", saveErr)
	}

	return updated, nil
}

func subscriptionStateForUser(user domain.User, now time.Time) domain.SubscriptionState {
	if !user.IsPremium {
		return domain.SubscriptionStateFree
	}
	if user.PremiumExpiredAt != nil && user.PremiumExpiredAt.Before(now) {
		return domain.SubscriptionStatePremiumExpired
	}
	return domain.SubscriptionStatePremiumActive
}

func isValidEmail(email string) bool {
	if strings.TrimSpace(email) == "" {
		return false
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(parsed.Address), email)
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasDigit := false
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	return hasUpper && hasDigit
}

func normalizeStringList(values []string, minItems, maxItems, minLen, maxLen int) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{})

	for _, raw := range values {
		trimmed := strings.ToLower(strings.TrimSpace(raw))
		if trimmed == "" {
			continue
		}
		if len(trimmed) < minLen || len(trimmed) > maxLen {
			return nil, errors.New("value length out of range")
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	if len(normalized) < minItems {
		return nil, errors.New("not enough items")
	}
	if len(normalized) > maxItems {
		return nil, errors.New("too many items")
	}

	return normalized, nil
}

func normalizeJobTypes(values []string) ([]string, error) {
	result, err := normalizeStringList(values, 1, 4, 4, 10)
	if err != nil {
		return nil, err
	}

	for _, value := range result {
		if _, ok := allowedJobTypes[value]; !ok {
			return nil, errors.New("job type is not allowed")
		}
	}

	return result, nil
}

func parseAlertMode(raw string) (domain.NotificationAlertMode, bool) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case string(domain.NotificationAlertModeInstant):
		return domain.NotificationAlertModeInstant, true
	case string(domain.NotificationAlertModeDailyDigest):
		return domain.NotificationAlertModeDailyDigest, true
	case string(domain.NotificationAlertModeWeeklyDigest):
		return domain.NotificationAlertModeWeeklyDigest, true
	default:
		return "", false
	}
}

func normalizeAlertModeOrDefault(raw string) domain.NotificationAlertMode {
	if parsed, ok := parseAlertMode(raw); ok {
		return parsed
	}
	return domain.NotificationAlertModeInstant
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
