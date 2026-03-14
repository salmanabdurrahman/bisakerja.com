package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

func TestManager_IssueAndParseAccessToken(t *testing.T) {
	manager, err := NewManager("unit-test-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	token, err := manager.IssueAccessToken("usr_123", identity.RoleUser)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	claims, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}

	if claims.Subject != "usr_123" {
		t.Fatalf("expected subject usr_123, got %s", claims.Subject)
	}
	if claims.Role != identity.RoleUser {
		t.Fatalf("expected role user, got %s", claims.Role)
	}
}

func TestManager_ParseAccessToken_RejectsRefreshToken(t *testing.T) {
	manager, err := NewManager("unit-test-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	refreshToken, err := manager.IssueRefreshToken("usr_123", identity.RoleUser)
	if err != nil {
		t.Fatalf("issue refresh token: %v", err)
	}

	_, err = manager.ParseAccessToken(refreshToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
