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

func setupPreferencesHandler(t *testing.T) (*PreferencesHandler, *identityapp.Service, string) {
	t.Helper()
	tokenManager, err := auth.NewManager("preferences-handler-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	service := identityapp.NewService(memory.NewIdentityRepository(), tokenManager)
	user, err := service.Register(context.Background(), identityapp.RegisterInput{
		Email:    "prefs-handler@example.com",
		Password: "StrongPass1",
		Name:     "Dina",
	})
	if err != nil {
		t.Fatalf("register user: %v", err)
	}
	return NewPreferencesHandler(service), service, user.ID
}

func TestPreferencesHandler_GetPreferences_DefaultPayload(t *testing.T) {
	handler, _, userID := setupPreferencesHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/preferences", nil)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_get_preferences"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   "user",
	}))
	response := httptest.NewRecorder()

	handler.GetPreferences(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestPreferencesHandler_UpdatePreferences_InvalidJobType(t *testing.T) {
	handler, _, userID := setupPreferencesHandler(t)

	requestBody := map[string]any{
		"keywords":  []string{"golang"},
		"job_types": []string{"unknown"},
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(requestBody); err != nil {
		t.Fatalf("encode request body: %v", err)
	}

	request := httptest.NewRequest(http.MethodPut, "/api/v1/preferences", &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_update_preferences_invalid_type"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   "user",
	}))
	response := httptest.NewRecorder()

	handler.UpdatePreferences(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestPreferencesHandler_UpdatePreferences_Success(t *testing.T) {
	handler, _, userID := setupPreferencesHandler(t)

	requestBody := map[string]any{
		"keywords":   []string{"golang", "backend"},
		"locations":  []string{"jakarta"},
		"job_types":  []string{"fulltime"},
		"salary_min": 10000000,
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(requestBody); err != nil {
		t.Fatalf("encode request body: %v", err)
	}

	request := httptest.NewRequest(http.MethodPut, "/api/v1/preferences", &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_update_preferences_success"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   "user",
	}))
	response := httptest.NewRecorder()

	handler.UpdatePreferences(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}
