package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
)

func TestAuthAndPreferencesFlow(t *testing.T) {
	tokenManager, err := auth.NewManager("integration-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)

	appHandler := router.New(
		logger.New("test"),
		router.Dependencies{
			AuthHandler:        handler.NewAuthHandler(identityService),
			PreferencesHandler: handler.NewPreferencesHandler(identityService, logger.New("test")),
			AuthMiddleware:     middleware.NewAuthenticator(tokenManager),
		},
	)

	registerPayload := map[string]any{
		"email":    "user@example.com",
		"password": "StrongPass1",
		"name":     "Budi",
	}
	registerResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/register", registerPayload, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d (%s)", registerResponse.Code, registerResponse.Body.String())
	}

	loginPayload := map[string]any{
		"email":    "user@example.com",
		"password": "StrongPass1",
	}
	loginResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/login", loginPayload, "")
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d (%s)", loginResponse.Code, loginResponse.Body.String())
	}
	var loginResult struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &loginResult); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResult.Data.AccessToken == "" || loginResult.Data.RefreshToken == "" {
		t.Fatalf("expected access_token and refresh_token, got %+v", loginResult.Data)
	}

	meResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/auth/me", nil, loginResult.Data.AccessToken)
	if meResponse.Code != http.StatusOK {
		t.Fatalf("expected me status 200, got %d (%s)", meResponse.Code, meResponse.Body.String())
	}

	getPreferencesResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/preferences", nil, loginResult.Data.AccessToken)
	if getPreferencesResponse.Code != http.StatusOK {
		t.Fatalf("expected get preferences status 200, got %d (%s)", getPreferencesResponse.Code, getPreferencesResponse.Body.String())
	}

	updatePreferencesPayload := map[string]any{
		"keywords":   []string{"Golang", "Backend", "Golang"},
		"locations":  []string{"Jakarta", "Remote"},
		"job_types":  []string{"fulltime", "contract"},
		"salary_min": 12000000,
	}
	updatePreferencesResponse := performJSONRequest(t, appHandler, http.MethodPut, "/api/v1/preferences", updatePreferencesPayload, loginResult.Data.AccessToken)
	if updatePreferencesResponse.Code != http.StatusOK {
		t.Fatalf("expected update preferences status 200, got %d (%s)", updatePreferencesResponse.Code, updatePreferencesResponse.Body.String())
	}

	refreshPayload := map[string]any{
		"refresh_token": loginResult.Data.RefreshToken,
	}
	refreshResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/refresh", refreshPayload, "")
	if refreshResponse.Code != http.StatusOK {
		t.Fatalf("expected refresh status 200, got %d (%s)", refreshResponse.Code, refreshResponse.Body.String())
	}
}

func TestProtectedRoutes_RequireAuthentication(t *testing.T) {
	tokenManager, err := auth.NewManager("integration-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)

	appHandler := router.New(
		logger.New("test"),
		router.Dependencies{
			AuthHandler:        handler.NewAuthHandler(identityService),
			PreferencesHandler: handler.NewPreferencesHandler(identityService, logger.New("test")),
			AuthMiddleware:     middleware.NewAuthenticator(tokenManager),
		},
	)

	meResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/auth/me", nil, "")
	if meResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected /auth/me status 401, got %d", meResponse.Code)
	}

	preferencesResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/preferences", nil, "")
	if preferencesResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected /preferences status 401, got %d", preferencesResponse.Code)
	}
}

func performJSONRequest(t *testing.T, appHandler http.Handler, method, path string, payload any, accessToken string) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode payload: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, &body)
	request.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}

	response := httptest.NewRecorder()
	appHandler.ServeHTTP(response, request)
	return response
}
