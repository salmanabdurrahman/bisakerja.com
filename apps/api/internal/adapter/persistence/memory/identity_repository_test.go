package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

func TestIdentityRepository_CreateAndFindUser(t *testing.T) {
	repository := NewIdentityRepository()

	user, err := repository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "user@example.com",
		PasswordHash: "hash",
		Name:         "Budi",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	foundByEmail, err := repository.GetUserByEmail(context.Background(), "USER@example.com")
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}
	if foundByEmail.ID != user.ID {
		t.Fatalf("expected user id %s, got %s", user.ID, foundByEmail.ID)
	}

	_, err = repository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "user@example.com",
		PasswordHash: "hash",
		Name:         "Budi",
		Role:         identity.RoleUser,
	})
	if !errors.Is(err, identity.ErrEmailAlreadyRegistered) {
		t.Fatalf("expected email already registered, got %v", err)
	}
}

func TestIdentityRepository_DefaultAndSavePreferences(t *testing.T) {
	repository := NewIdentityRepository()
	user, err := repository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "prefs@example.com",
		PasswordHash: "hash",
		Name:         "Nina",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	defaultPreferences, err := repository.GetPreferences(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get default preferences: %v", err)
	}
	if len(defaultPreferences.Keywords) != 0 || defaultPreferences.SalaryMin != 0 {
		t.Fatalf("expected empty default preferences, got %+v", defaultPreferences)
	}

	now := time.Now().UTC()
	saved, err := repository.SavePreferences(context.Background(), identity.Preferences{
		UserID:    user.ID,
		Keywords:  []string{"golang"},
		Locations: []string{"jakarta"},
		JobTypes:  []string{"fulltime"},
		SalaryMin: 10000000,
		UpdatedAt: &now,
	})
	if err != nil {
		t.Fatalf("save preferences: %v", err)
	}
	if len(saved.Keywords) != 1 || saved.Keywords[0] != "golang" {
		t.Fatalf("unexpected saved preferences: %+v", saved)
	}
}

func TestIdentityRepository_UpdatePremiumStatus(t *testing.T) {
	repository := NewIdentityRepository()
	user, err := repository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "premium@example.com",
		PasswordHash: "hash",
		Name:         "Santi",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	expiredAt := time.Now().UTC().Add(30 * 24 * time.Hour)
	updated, err := repository.UpdatePremiumStatus(context.Background(), user.ID, true, &expiredAt)
	if err != nil {
		t.Fatalf("update premium status: %v", err)
	}
	if !updated.IsPremium {
		t.Fatal("expected is_premium true after update")
	}
	if updated.PremiumExpiredAt == nil || !updated.PremiumExpiredAt.Equal(expiredAt) {
		t.Fatalf("expected premium_expired_at %v, got %v", expiredAt, updated.PremiumExpiredAt)
	}
}
