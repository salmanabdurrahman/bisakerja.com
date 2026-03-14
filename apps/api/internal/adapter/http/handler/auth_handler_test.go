package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

func setupAuthHandler(t *testing.T) (*AuthHandler, *identityapp.Service) {
	t.Helper()
	tokenManager, err := auth.NewManager("auth-handler-test-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	service := identityapp.NewService(memory.NewIdentityRepository(), tokenManager)
	return NewAuthHandler(service), service
}

func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	authHandler, _ := setupAuthHandler(t)

	requestBody := map[string]any{
		"email":    "not-an-email",
		"password": "StrongPass1",
		"name":     "Budi",
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(requestBody); err != nil {
		t.Fatalf("encode request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_register_invalid_email"))
	response := httptest.NewRecorder()

	authHandler.Register(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	authHandler, service := setupAuthHandler(t)
	_, err := service.Register(context.Background(), identityapp.RegisterInput{
		Email:    "user@example.com",
		Password: "StrongPass1",
		Name:     "Budi",
	})
	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	requestBody := map[string]any{
		"email":    "user@example.com",
		"password": "WrongPass1",
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(requestBody); err != nil {
		t.Fatalf("encode request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_login_invalid_credentials"))
	response := httptest.NewRecorder()

	authHandler.Login(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
}

func TestAuthHandler_Me_UsesAuthenticatedContext(t *testing.T) {
	authHandler, service := setupAuthHandler(t)
	user, err := service.Register(context.Background(), identityapp.RegisterInput{
		Email:    "me@example.com",
		Password: "StrongPass1",
		Name:     "Siti",
	})
	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_me"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: user.ID,
		Role:   user.Role,
	}))
	response := httptest.NewRecorder()

	authHandler.Me(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}
