package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	identityapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/identity"
	trackerapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/tracker"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/auth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
)

func TestTrackerFlow(t *testing.T) {
	tokenManager, err := auth.NewManager("integration-secret-tracker", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	identityRepository := memory.NewIdentityRepository()
	identityService := identityapp.NewService(identityRepository, tokenManager)
	authHandler := handler.NewAuthHandler(identityService)
	preferencesHandler := handler.NewPreferencesHandler(identityService)
	authMiddleware := middleware.NewAuthenticator(tokenManager)

	trackerRepository := memory.NewTrackerRepository()
	trackerService := trackerapp.NewService(identityRepository, trackerRepository)
	trackerHandler := handler.NewTrackerHandler(trackerService)

	appHandler := router.New(
		logger.New("test"),
		router.Dependencies{
			AuthHandler:        authHandler,
			PreferencesHandler: preferencesHandler,
			TrackerHandler:     trackerHandler,
			AuthMiddleware:     authMiddleware,
		},
	)

	// Register + Login
	registerPayload := map[string]any{
		"email":    "tracker-flow@example.com",
		"password": "StrongPass1",
		"name":     "Tracker Flow User",
	}
	registerResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/auth/register", registerPayload, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d (%s)", registerResponse.Code, registerResponse.Body.String())
	}

	loginPayload := map[string]any{
		"email":    "tracker-flow@example.com",
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
	token := loginResult.Data.AccessToken

	// --- Bookmark flow ---

	// Create bookmark
	bookmarkPayload := map[string]any{"job_id": "job-tracker-1"}
	createBookmarkResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/bookmarks", bookmarkPayload, token)
	if createBookmarkResponse.Code != http.StatusCreated {
		t.Fatalf("expected create bookmark status 201, got %d (%s)", createBookmarkResponse.Code, createBookmarkResponse.Body.String())
	}

	// Duplicate bookmark should 409
	dupBookmarkResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/bookmarks", bookmarkPayload, token)
	if dupBookmarkResponse.Code != http.StatusConflict {
		t.Fatalf("expected duplicate bookmark status 409, got %d (%s)", dupBookmarkResponse.Code, dupBookmarkResponse.Body.String())
	}

	// List bookmarks
	listBookmarksResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/bookmarks", nil, token)
	if listBookmarksResponse.Code != http.StatusOK {
		t.Fatalf("expected list bookmarks status 200, got %d (%s)", listBookmarksResponse.Code, listBookmarksResponse.Body.String())
	}

	// Delete bookmark
	deleteBookmarkResponse := performJSONRequest(t, appHandler, http.MethodDelete, "/api/v1/bookmarks/job-tracker-1", nil, token)
	if deleteBookmarkResponse.Code != http.StatusOK {
		t.Fatalf("expected delete bookmark status 200, got %d (%s)", deleteBookmarkResponse.Code, deleteBookmarkResponse.Body.String())
	}

	// Delete non-existent bookmark should 404
	notFoundBookmarkResponse := performJSONRequest(t, appHandler, http.MethodDelete, "/api/v1/bookmarks/job-tracker-1", nil, token)
	if notFoundBookmarkResponse.Code != http.StatusNotFound {
		t.Fatalf("expected not-found bookmark status 404, got %d (%s)", notFoundBookmarkResponse.Code, notFoundBookmarkResponse.Body.String())
	}

	// --- Tracked application flow ---

	// Create tracked application
	appPayload := map[string]any{
		"job_id": "job-tracker-2",
		"notes":  "Applied on company website",
	}
	createAppResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/applications", appPayload, token)
	if createAppResponse.Code != http.StatusCreated {
		t.Fatalf("expected create application status 201, got %d (%s)", createAppResponse.Code, createAppResponse.Body.String())
	}
	var createAppResult struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createAppResponse.Body.Bytes(), &createAppResult); err != nil {
		t.Fatalf("decode create application response: %v", err)
	}
	if createAppResult.Data.Status != "applied" {
		t.Fatalf("expected initial status 'applied', got %s", createAppResult.Data.Status)
	}
	applicationID := createAppResult.Data.ID

	// Duplicate application should 409
	dupAppResponse := performJSONRequest(t, appHandler, http.MethodPost, "/api/v1/applications", appPayload, token)
	if dupAppResponse.Code != http.StatusConflict {
		t.Fatalf("expected duplicate application status 409, got %d (%s)", dupAppResponse.Code, dupAppResponse.Body.String())
	}

	// Update application status
	statusPayload := map[string]any{"status": "interview"}
	updateStatusResponse := performJSONRequest(t, appHandler, http.MethodPatch, "/api/v1/applications/"+applicationID+"/status", statusPayload, token)
	if updateStatusResponse.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d (%s)", updateStatusResponse.Code, updateStatusResponse.Body.String())
	}

	// Invalid status update should 400
	invalidStatusResponse := performJSONRequest(t, appHandler, http.MethodPatch, "/api/v1/applications/"+applicationID+"/status", map[string]any{"status": "bad_status"}, token)
	if invalidStatusResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid status 400, got %d (%s)", invalidStatusResponse.Code, invalidStatusResponse.Body.String())
	}

	// List tracked applications
	listAppsResponse := performJSONRequest(t, appHandler, http.MethodGet, "/api/v1/applications", nil, token)
	if listAppsResponse.Code != http.StatusOK {
		t.Fatalf("expected list applications status 200, got %d (%s)", listAppsResponse.Code, listAppsResponse.Body.String())
	}

	// Delete tracked application
	deleteAppResponse := performJSONRequest(t, appHandler, http.MethodDelete, "/api/v1/applications/"+applicationID, nil, token)
	if deleteAppResponse.Code != http.StatusOK {
		t.Fatalf("expected delete application status 200, got %d (%s)", deleteAppResponse.Code, deleteAppResponse.Body.String())
	}

	// Delete non-existent application should 404
	notFoundAppResponse := performJSONRequest(t, appHandler, http.MethodDelete, "/api/v1/applications/"+applicationID, nil, token)
	if notFoundAppResponse.Code != http.StatusNotFound {
		t.Fatalf("expected not-found application status 404, got %d (%s)", notFoundAppResponse.Code, notFoundAppResponse.Body.String())
	}
}
