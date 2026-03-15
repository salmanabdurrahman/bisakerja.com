package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	growthapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/growth"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	notificationapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/notification"
	notificationdomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
)

func TestGrowthFeaturesFlow(t *testing.T) {
	tokenManager, err := auth.NewManager("integration-secret", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)
	authHandler := handler.NewAuthHandler(identityService)
	preferencesHandler := handler.NewPreferencesHandler(identityService, logger.New("test"))
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	growthRepository := memory.NewGrowthRepository()
	growthService := growthapp.NewService(identityRepository, growthRepository)
	growthHandler := handler.NewGrowthHandler(growthService)

	notificationRepository := memory.NewNotificationRepository()
	notificationCenterService := notificationapp.NewCenterService(identityRepository, notificationRepository)
	notificationHandler := handler.NewNotificationHandler(notificationCenterService)

	appHandler := router.New(
		logger.New("test"),
		router.Dependencies{
			AuthHandler:         authHandler,
			PreferencesHandler:  preferencesHandler,
			GrowthHandler:       growthHandler,
			NotificationHandler: notificationHandler,
			AuthMiddleware:      authMiddleware,
		},
	)

	registerPayload := map[string]any{
		"email":    "growth-flow@example.com",
		"password": "StrongPass1",
		"name":     "Growth Flow User",
	}
	registerResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/register", registerPayload, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d (%s)", registerResponse.Code, registerResponse.Body.String())
	}
	var registerResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(registerResponse.Body.Bytes(), &registerResult); err != nil {
		t.Fatalf("decode register response: %v", err)
	}

	loginPayload := map[string]any{
		"email":    "growth-flow@example.com",
		"password": "StrongPass1",
	}
	loginResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/login", loginPayload, "")
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d (%s)", loginResponse.Code, loginResponse.Body.String())
	}
	var loginResult struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &loginResult); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	savedSearchPayload := map[string]any{
		"query":      "golang backend",
		"location":   "jakarta",
		"source":     "glints",
		"salary_min": 12000000,
		"frequency":  "daily_digest",
	}
	createSavedSearchResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/saved-searches", savedSearchPayload, loginResult.Data.AccessToken)
	if createSavedSearchResponse.Code != http.StatusCreated {
		t.Fatalf("expected create saved search status 201, got %d (%s)", createSavedSearchResponse.Code, createSavedSearchResponse.Body.String())
	}
	listSavedSearchesResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/saved-searches", nil, loginResult.Data.AccessToken)
	if listSavedSearchesResponse.Code != http.StatusOK {
		t.Fatalf("expected list saved searches status 200, got %d (%s)", listSavedSearchesResponse.Code, listSavedSearchesResponse.Body.String())
	}

	watchlistPayload := map[string]any{
		"company_slug": "acme-group",
	}
	createWatchlistResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/watchlist/companies", watchlistPayload, loginResult.Data.AccessToken)
	if createWatchlistResponse.Code != http.StatusCreated {
		t.Fatalf("expected create watchlist status 201, got %d (%s)", createWatchlistResponse.Code, createWatchlistResponse.Body.String())
	}
	listWatchlistResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/watchlist/companies", nil, loginResult.Data.AccessToken)
	if listWatchlistResponse.Code != http.StatusOK {
		t.Fatalf("expected list watchlist status 200, got %d (%s)", listWatchlistResponse.Code, listWatchlistResponse.Body.String())
	}

	updateNotificationPreferencesPayload := map[string]any{
		"alert_mode":  "daily_digest",
		"digest_hour": 7,
	}
	updateNotificationPreferencesResponse := performJSONRequest(t, appHandler, http.MethodPut, "/api/v1/preferences/notification", updateNotificationPreferencesPayload, loginResult.Data.AccessToken)
	if updateNotificationPreferencesResponse.Code != http.StatusOK {
		t.Fatalf("expected update notification preferences status 200, got %d (%s)", updateNotificationPreferencesResponse.Code, updateNotificationPreferencesResponse.Body.String())
	}

	createdNotification, err := notificationRepository.CreatePending(context.Background(), notificationdomain.CreateInput{
		UserID:  registerResult.Data.ID,
		JobID:   "job_growth_1",
		Channel: notificationdomain.ChannelEmail,
	})
	if err != nil {
		t.Fatalf("seed notification: %v", err)
	}

	listNotificationsResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/notifications?page=1&limit=10&unread_only=true", nil, loginResult.Data.AccessToken)
	if listNotificationsResponse.Code != http.StatusOK {
		t.Fatalf("expected list notifications status 200, got %d (%s)", listNotificationsResponse.Code, listNotificationsResponse.Body.String())
	}

	markReadResponse := performJSONRequest(t, appHandler, http.MethodPatch, "/api/v1/notifications/"+createdNotification.ID+"/read", nil, loginResult.Data.AccessToken)
	if markReadResponse.Code != http.StatusOK {
		t.Fatalf("expected mark notification read status 200, got %d (%s)", markReadResponse.Code, markReadResponse.Body.String())
	}
}
