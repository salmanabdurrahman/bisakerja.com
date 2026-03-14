package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	domain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	platformauth "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
)

func newIdentityServiceForTest(t *testing.T) *Service {
	t.Helper()
	tokenManager, err := platformauth.NewManager("service-test-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	return NewService(memory.NewIdentityRepository(), tokenManager)
}

func TestService_RegisterLoginAndRefresh(t *testing.T) {
	service := newIdentityServiceForTest(t)

	user, err := service.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "StrongPass1",
		Name:     "Budi",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Email != "user@example.com" {
		t.Fatalf("expected normalized email user@example.com, got %s", user.Email)
	}

	tokens, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "StrongPass1",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens, got %+v", tokens)
	}

	newToken, err := service.Refresh(context.Background(), tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if newToken.AccessToken == "" {
		t.Fatal("expected non-empty refreshed access token")
	}
}

func TestService_LoginInvalidPassword_ReturnsInvalidCredentials(t *testing.T) {
	service := newIdentityServiceForTest(t)
	_, err := service.Register(context.Background(), RegisterInput{
		Email:    "user2@example.com",
		Password: "StrongPass1",
		Name:     "Siti",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = service.Login(context.Background(), LoginInput{
		Email:    "user2@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestService_UpdatePreferences_NormalizesAndValidates(t *testing.T) {
	service := newIdentityServiceForTest(t)
	user, err := service.Register(context.Background(), RegisterInput{
		Email:    "prefs@example.com",
		Password: "StrongPass1",
		Name:     "Nina",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	updated, err := service.UpdatePreferences(context.Background(), user.ID, UpdatePreferencesInput{
		Keywords:     []string{" Golang ", "Backend", "golang"},
		KeywordsSet:  true,
		Locations:    []string{" Jakarta ", "Remote"},
		LocationsSet: true,
		JobTypes:     []string{"FullTime", "contract"},
		JobTypesSet:  true,
		SalaryMin:    10_000_000,
		SalaryMinSet: true,
	})
	if err != nil {
		t.Fatalf("update preferences: %v", err)
	}

	if len(updated.Keywords) != 2 || updated.Keywords[0] != "golang" {
		t.Fatalf("expected normalized deduplicated keywords, got %+v", updated.Keywords)
	}
	if len(updated.JobTypes) != 2 || updated.JobTypes[0] != "fulltime" {
		t.Fatalf("expected normalized job types, got %+v", updated.JobTypes)
	}

	_, err = service.UpdatePreferences(context.Background(), user.ID, UpdatePreferencesInput{
		Keywords:    []string{"go"},
		KeywordsSet: true,
		JobTypes:    []string{"invalid-type"},
		JobTypesSet: true,
	})
	if !errors.Is(err, ErrInvalidJobType) {
		t.Fatalf("expected invalid job type error, got %v", err)
	}
}

func TestService_UpdateNotificationPreferences_ValidatesAndPersists(t *testing.T) {
	service := newIdentityServiceForTest(t)
	user, err := service.Register(context.Background(), RegisterInput{
		Email:    "notify-prefs@example.com",
		Password: "StrongPass1",
		Name:     "Rina",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	updated, err := service.UpdateNotificationPreferences(context.Background(), user.ID, UpdateNotificationPreferencesInput{
		AlertMode:     "daily_digest",
		AlertModeSet:  true,
		DigestHour:    7,
		DigestHourSet: true,
	})
	if err != nil {
		t.Fatalf("update notification preferences: %v", err)
	}
	if updated.AlertMode != domain.NotificationAlertModeDailyDigest {
		t.Fatalf("expected daily_digest alert mode, got %s", updated.AlertMode)
	}
	if updated.DigestHour == nil || *updated.DigestHour != 7 {
		t.Fatalf("expected digest_hour=7, got %+v", updated.DigestHour)
	}

	_, err = service.UpdateNotificationPreferences(context.Background(), user.ID, UpdateNotificationPreferencesInput{
		AlertMode:     "instant",
		AlertModeSet:  true,
		DigestHour:    9,
		DigestHourSet: true,
	})
	if !errors.Is(err, ErrInvalidDigestHour) {
		t.Fatalf("expected invalid digest hour for instant mode, got %v", err)
	}
}
