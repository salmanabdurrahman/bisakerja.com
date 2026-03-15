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
	trackerapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/tracker"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

func setupTrackerHandler(t *testing.T) (*TrackerHandler, string) {
	t.Helper()

	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "tracker-handler@example.com",
		PasswordHash: "hash",
		Name:         "Tracker Handler User",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	service := trackerapp.NewService(identityRepository, memory.NewTrackerRepository())
	return NewTrackerHandler(service), user.ID
}

func TestTrackerHandler_BookmarkCRUD(t *testing.T) {
	handler, userID := setupTrackerHandler(t)

	createPayload := map[string]any{
		"job_id": "job-123",
	}
	createResponse := performTrackerRequest(t, handler.CreateBookmark, http.MethodPost, "/api/v1/bookmarks", userID, createPayload, "")
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create bookmark 201, got %d (%s)", createResponse.Code, createResponse.Body.String())
	}

	var createResult struct {
		Data struct {
			JobID string `json:"job_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createResult); err != nil {
		t.Fatalf("decode create bookmark response: %v", err)
	}
	if createResult.Data.JobID != "job-123" {
		t.Fatalf("expected job_id 'job-123', got %s", createResult.Data.JobID)
	}

	listResponse := performTrackerRequest(t, handler.ListBookmarks, http.MethodGet, "/api/v1/bookmarks", userID, nil, "")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list bookmarks 200, got %d (%s)", listResponse.Code, listResponse.Body.String())
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/bookmarks/job-123", nil)
	deleteRequest = deleteRequest.WithContext(observability.WithRequestID(deleteRequest.Context(), "req_bookmark_delete"))
	deleteRequest = deleteRequest.WithContext(middleware.WithAuthUser(deleteRequest.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	deleteRequest.SetPathValue("job_id", "job-123")
	deleteResponse := httptest.NewRecorder()
	handler.DeleteBookmark(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected delete bookmark 200, got %d (%s)", deleteResponse.Code, deleteResponse.Body.String())
	}
}

func TestTrackerHandler_BookmarkDuplicate(t *testing.T) {
	handler, userID := setupTrackerHandler(t)

	createPayload := map[string]any{"job_id": "job-dup"}
	performTrackerRequest(t, handler.CreateBookmark, http.MethodPost, "/api/v1/bookmarks", userID, createPayload, "")

	dupResponse := performTrackerRequest(t, handler.CreateBookmark, http.MethodPost, "/api/v1/bookmarks", userID, createPayload, "")
	if dupResponse.Code != http.StatusConflict {
		t.Fatalf("expected duplicate bookmark 409, got %d (%s)", dupResponse.Code, dupResponse.Body.String())
	}
}

func TestTrackerHandler_TrackedApplicationCRUD(t *testing.T) {
	handler, userID := setupTrackerHandler(t)

	createPayload := map[string]any{
		"job_id": "job-456",
		"notes":  "Applied via LinkedIn",
	}
	createResponse := performTrackerRequest(t, handler.CreateTrackedApplication, http.MethodPost, "/api/v1/applications", userID, createPayload, "")
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create tracked application 201, got %d (%s)", createResponse.Code, createResponse.Body.String())
	}

	var createResult struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createResult); err != nil {
		t.Fatalf("decode create tracked application response: %v", err)
	}
	if createResult.Data.ID == "" {
		t.Fatalf("expected tracked application id, got %s", createResponse.Body.String())
	}
	if createResult.Data.Status != "applied" {
		t.Fatalf("expected status 'applied', got %s", createResult.Data.Status)
	}

	patchRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/applications/"+createResult.Data.ID+"/status", nil)
	patchPayload := map[string]any{"status": "interview"}
	var patchBody bytes.Buffer
	if err := json.NewEncoder(&patchBody).Encode(patchPayload); err != nil {
		t.Fatalf("encode patch payload: %v", err)
	}
	patchRequest = httptest.NewRequest(http.MethodPatch, "/api/v1/applications/"+createResult.Data.ID+"/status", &patchBody)
	patchRequest = patchRequest.WithContext(observability.WithRequestID(patchRequest.Context(), "req_app_patch"))
	patchRequest = patchRequest.WithContext(middleware.WithAuthUser(patchRequest.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	patchRequest.Header.Set("Content-Type", "application/json")
	patchRequest.SetPathValue("id", createResult.Data.ID)
	patchResponse := httptest.NewRecorder()
	handler.UpdateApplicationStatus(patchResponse, patchRequest)
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d (%s)", patchResponse.Code, patchResponse.Body.String())
	}

	listResponse := performTrackerRequest(t, handler.ListTrackedApplications, http.MethodGet, "/api/v1/applications", userID, nil, "")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list applications 200, got %d (%s)", listResponse.Code, listResponse.Body.String())
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/applications/"+createResult.Data.ID, nil)
	deleteRequest = deleteRequest.WithContext(observability.WithRequestID(deleteRequest.Context(), "req_app_delete"))
	deleteRequest = deleteRequest.WithContext(middleware.WithAuthUser(deleteRequest.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	deleteRequest.SetPathValue("id", createResult.Data.ID)
	deleteResponse := httptest.NewRecorder()
	handler.DeleteTrackedApplication(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected delete application 200, got %d (%s)", deleteResponse.Code, deleteResponse.Body.String())
	}
}

func TestTrackerHandler_InvalidStatus(t *testing.T) {
	handler, userID := setupTrackerHandler(t)

	createPayload := map[string]any{"job_id": "job-status-test"}
	createResponse := performTrackerRequest(t, handler.CreateTrackedApplication, http.MethodPost, "/api/v1/applications", userID, createPayload, "")
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("seed application: %d %s", createResponse.Code, createResponse.Body.String())
	}

	var createResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createResult); err != nil {
		t.Fatalf("decode: %v", err)
	}

	var patchBody bytes.Buffer
	if err := json.NewEncoder(&patchBody).Encode(map[string]any{"status": "not_a_status"}); err != nil {
		t.Fatalf("encode: %v", err)
	}
	patchRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/applications/"+createResult.Data.ID+"/status", &patchBody)
	patchRequest = patchRequest.WithContext(observability.WithRequestID(patchRequest.Context(), "req_invalid_status"))
	patchRequest = patchRequest.WithContext(middleware.WithAuthUser(patchRequest.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	patchRequest.Header.Set("Content-Type", "application/json")
	patchRequest.SetPathValue("id", createResult.Data.ID)
	patchResponse := httptest.NewRecorder()
	handler.UpdateApplicationStatus(patchResponse, patchRequest)
	if patchResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid status, got %d (%s)", patchResponse.Code, patchResponse.Body.String())
	}
}

func performTrackerRequest(
	t *testing.T,
	handlerFunc http.HandlerFunc,
	method string,
	path string,
	userID string,
	payload map[string]any,
	pathValueID string,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode payload: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, &body)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_tracker_"+time.Now().UTC().Format("150405.000000000")))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: userID,
		Role:   identity.RoleUser,
	}))
	request.Header.Set("Content-Type", "application/json")
	if pathValueID != "" {
		request.SetPathValue("id", pathValueID)
	}

	response := httptest.NewRecorder()
	handlerFunc(response, request)
	return response
}
