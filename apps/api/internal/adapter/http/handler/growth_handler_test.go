package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	growthapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/growth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

func setupGrowthHandler(t *testing.T) (*GrowthHandler, string) {
	t.Helper()

	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "growth-handler@example.com",
		PasswordHash: "hash",
		Name:         "Growth Handler User",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	service := growthapp.NewService(identityRepository, memory.NewGrowthRepository())
	return NewGrowthHandler(service), user.ID
}

func TestGrowthHandler_SavedSearchCRUD(t *testing.T) {
	handler, userID := setupGrowthHandler(t)

	createPayload := map[string]any{
		"query":      "golang backend",
		"location":   "jakarta",
		"source":     "glints",
		"salary_min": 12000000,
		"frequency":  "daily_digest",
	}
	createResponse := performGrowthRequest(t, handler.CreateSavedSearch, http.MethodPost, "/api/v1/saved-searches", userID, createPayload)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create saved search 201, got %d (%s)", createResponse.Code, createResponse.Body.String())
	}

	var createResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createResult); err != nil {
		t.Fatalf("decode create saved search response: %v", err)
	}
	if createResult.Data.ID == "" {
		t.Fatalf("expected saved search id, got %s", createResponse.Body.String())
	}

	listResponse := performGrowthRequest(t, handler.ListSavedSearches, http.MethodGet, "/api/v1/saved-searches", userID, nil)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list saved searches 200, got %d (%s)", listResponse.Code, listResponse.Body.String())
	}

	deleteResponse := performGrowthRequest(t, handler.DeleteSavedSearch, http.MethodDelete, "/api/v1/saved-searches/"+createResult.Data.ID, userID, nil)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected delete saved search 200, got %d (%s)", deleteResponse.Code, deleteResponse.Body.String())
	}
}

func TestGrowthHandler_WatchlistCRUD(t *testing.T) {
	handler, userID := setupGrowthHandler(t)

	createPayload := map[string]any{
		"company_slug": "acme-group",
	}
	createResponse := performGrowthRequest(t, handler.CreateWatchlistCompany, http.MethodPost, "/api/v1/watchlist/companies", userID, createPayload)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create watchlist 201, got %d (%s)", createResponse.Code, createResponse.Body.String())
	}

	listResponse := performGrowthRequest(t, handler.ListWatchlistCompanies, http.MethodGet, "/api/v1/watchlist/companies", userID, nil)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list watchlist 200, got %d (%s)", listResponse.Code, listResponse.Body.String())
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/watchlist/companies/acme-group", nil)
	deleteRequest = deleteRequest.WithContext(observability.WithRequestID(deleteRequest.Context(), "req_watchlist_delete"))
	deleteRequest = deleteRequest.WithContext(middleware.WithAuthUser(deleteRequest.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	deleteRequest.SetPathValue("company_slug", "acme-group")
	deleteResponse := httptest.NewRecorder()
	handler.DeleteWatchlistCompany(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected delete watchlist 200, got %d (%s)", deleteResponse.Code, deleteResponse.Body.String())
	}
}

func performGrowthRequest(
	t *testing.T,
	handlerFunc http.HandlerFunc,
	method string,
	path string,
	userID string,
	payload map[string]any,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode payload: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_growth_"+time.Now().UTC().Format("150405.000000000")))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	request.Header.Set("Content-Type", "application/json")
	if strings.HasPrefix(path, "/api/v1/saved-searches/") {
		request.SetPathValue("id", strings.TrimPrefix(path, "/api/v1/saved-searches/"))
	}

	response := httptest.NewRecorder()
	handlerFunc(response, request)
	return response
}
