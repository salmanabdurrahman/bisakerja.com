package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

// Claims represents claims.
type Claims struct {
	Role      identity.Role `json:"role"`
	TokenType string        `json:"token_type"`
	jwt.RegisteredClaims
}

// Manager represents manager.
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

// NewManager creates a new manager instance.
func NewManager(secret string, accessTTL, refreshTTL time.Duration) (*Manager, error) {
	trimmedSecret := strings.TrimSpace(secret)
	if trimmedSecret == "" {
		return nil, errors.New("jwt secret is required")
	}
	if accessTTL <= 0 {
		return nil, errors.New("access ttl must be > 0")
	}
	if refreshTTL <= 0 {
		return nil, errors.New("refresh ttl must be > 0")
	}

	return &Manager{
		secret:     []byte(trimmedSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		now:        func() time.Time { return time.Now().UTC() },
	}, nil
}

// AccessTokenTTLSeconds handles access token ttl seconds.
func (m *Manager) AccessTokenTTLSeconds() int {
	return int(m.accessTTL.Seconds())
}

// IssueAccessToken issues access token.
func (m *Manager) IssueAccessToken(userID string, role identity.Role) (string, error) {
	return m.issue(userID, role, tokenTypeAccess, m.accessTTL)
}

// IssueRefreshToken issues refresh token.
func (m *Manager) IssueRefreshToken(userID string, role identity.Role) (string, error) {
	return m.issue(userID, role, tokenTypeRefresh, m.refreshTTL)
}

// ParseAccessToken parses access token.
func (m *Manager) ParseAccessToken(rawToken string) (Claims, error) {
	return m.parse(rawToken, tokenTypeAccess)
}

// ParseRefreshToken parses refresh token.
func (m *Manager) ParseRefreshToken(rawToken string) (Claims, error) {
	return m.parse(rawToken, tokenTypeRefresh)
}

func (m *Manager) issue(userID string, role identity.Role, tokenType string, ttl time.Duration) (string, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return "", errors.New("user id is required")
	}
	if strings.TrimSpace(string(role)) == "" {
		return "", errors.New("role is required")
	}

	now := m.now()
	claims := Claims{
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   trimmedUserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (m *Manager) parse(rawToken, expectedType string) (Claims, error) {
	trimmedToken := strings.TrimSpace(rawToken)
	if trimmedToken == "" {
		return Claims{}, ErrInvalidToken
	}

	claims := Claims{}
	parsed, err := jwt.ParseWithClaims(trimmedToken, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, ErrTokenExpired
		}
		return Claims{}, ErrInvalidToken
	}
	if parsed == nil || !parsed.Valid {
		return Claims{}, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return Claims{}, ErrInvalidToken
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return Claims{}, ErrInvalidToken
	}
	if claims.Role != identity.RoleUser && claims.Role != identity.RoleAdmin {
		return Claims{}, ErrInvalidToken
	}

	return claims, nil
}
