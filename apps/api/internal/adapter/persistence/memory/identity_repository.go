package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

type IdentityRepository struct {
	mu              sync.RWMutex
	usersByID       map[string]identity.User
	userIDByEmail   map[string]string
	preferencesByID map[string]identity.Preferences
}

func NewIdentityRepository() *IdentityRepository {
	return &IdentityRepository{
		usersByID:       make(map[string]identity.User),
		userIDByEmail:   make(map[string]string),
		preferencesByID: make(map[string]identity.Preferences),
	}
}

func (r *IdentityRepository) CreateUser(_ context.Context, input identity.CreateUserInput) (identity.User, error) {
	normalizedEmail := identity.NormalizeEmail(input.Email)
	if normalizedEmail == "" {
		return identity.User{}, fmt.Errorf("create user: email is required")
	}
	if strings.TrimSpace(input.PasswordHash) == "" {
		return identity.User{}, fmt.Errorf("create user: password hash is required")
	}
	if strings.TrimSpace(input.Name) == "" {
		return identity.User{}, fmt.Errorf("create user: name is required")
	}

	now := time.Now().UTC()
	role := input.Role
	if role == "" {
		role = identity.RoleUser
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.userIDByEmail[normalizedEmail]; exists {
		return identity.User{}, identity.ErrEmailAlreadyRegistered
	}

	userID := "usr_" + randomHex(12)
	user := identity.User{
		ID:               userID,
		Email:            normalizedEmail,
		PasswordHash:     input.PasswordHash,
		Name:             strings.TrimSpace(input.Name),
		Role:             role,
		IsPremium:        input.IsPremium,
		PremiumExpiredAt: input.PremiumExpiredAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	r.usersByID[userID] = user
	r.userIDByEmail[normalizedEmail] = userID
	return user, nil
}

func (r *IdentityRepository) GetUserByID(_ context.Context, userID string) (identity.User, error) {
	trimmedUserID := strings.TrimSpace(userID)
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.usersByID[trimmedUserID]
	if !ok {
		return identity.User{}, identity.ErrUserNotFound
	}
	return user, nil
}

func (r *IdentityRepository) GetUserByEmail(_ context.Context, email string) (identity.User, error) {
	normalizedEmail := identity.NormalizeEmail(email)
	r.mu.RLock()
	defer r.mu.RUnlock()

	userID, ok := r.userIDByEmail[normalizedEmail]
	if !ok {
		return identity.User{}, identity.ErrUserNotFound
	}
	user, ok := r.usersByID[userID]
	if !ok {
		return identity.User{}, identity.ErrUserNotFound
	}
	return user, nil
}

func (r *IdentityRepository) ListUsers(_ context.Context) ([]identity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]identity.User, 0, len(r.usersByID))
	for _, user := range r.usersByID {
		result = append(result, user)
	}
	return result, nil
}

func (r *IdentityRepository) GetPreferences(_ context.Context, userID string) (identity.Preferences, error) {
	trimmedUserID := strings.TrimSpace(userID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, exists := r.usersByID[trimmedUserID]; !exists {
		return identity.Preferences{}, identity.ErrUserNotFound
	}

	if preferences, ok := r.preferencesByID[trimmedUserID]; ok {
		return clonePreferences(preferences), nil
	}

	return identity.Preferences{
		UserID:    trimmedUserID,
		Keywords:  []string{},
		Locations: []string{},
		JobTypes:  []string{},
		SalaryMin: 0,
		UpdatedAt: nil,
	}, nil
}

func (r *IdentityRepository) SavePreferences(_ context.Context, preferences identity.Preferences) (identity.Preferences, error) {
	trimmedUserID := strings.TrimSpace(preferences.UserID)
	if trimmedUserID == "" {
		return identity.Preferences{}, fmt.Errorf("save preferences: user id is required")
	}
	if preferences.UpdatedAt == nil {
		return identity.Preferences{}, fmt.Errorf("save preferences: updated_at is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByID[trimmedUserID]; !exists {
		return identity.Preferences{}, identity.ErrUserNotFound
	}

	snapshot := identity.Preferences{
		UserID:    trimmedUserID,
		Keywords:  append([]string(nil), preferences.Keywords...),
		Locations: append([]string(nil), preferences.Locations...),
		JobTypes:  append([]string(nil), preferences.JobTypes...),
		SalaryMin: preferences.SalaryMin,
		UpdatedAt: preferences.UpdatedAt,
	}
	r.preferencesByID[trimmedUserID] = snapshot
	return clonePreferences(snapshot), nil
}

func clonePreferences(value identity.Preferences) identity.Preferences {
	result := identity.Preferences{
		UserID:    value.UserID,
		Keywords:  append([]string(nil), value.Keywords...),
		Locations: append([]string(nil), value.Locations...),
		JobTypes:  append([]string(nil), value.JobTypes...),
		SalaryMin: value.SalaryMin,
		UpdatedAt: value.UpdatedAt,
	}
	return result
}
