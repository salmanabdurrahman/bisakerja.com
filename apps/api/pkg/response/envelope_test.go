package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteSuccess(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteSuccess(recorder, http.StatusCreated, "created", "req_abc", map[string]string{"id": "1"})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected content type application/json, got %q", contentType)
	}

	var payload struct {
		Meta Meta              `json:"meta"`
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload.Meta.Status != "success" {
		t.Fatalf("expected meta status success, got %q", payload.Meta.Status)
	}

	if payload.Meta.RequestID != "req_abc" {
		t.Fatalf("expected request id req_abc, got %q", payload.Meta.RequestID)
	}

	if payload.Data["id"] != "1" {
		t.Fatalf("expected data.id=1, got %q", payload.Data["id"])
	}
}

func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteError(recorder, http.StatusBadRequest, "validation error", "req_xyz", []ErrorItem{{
		Field:   "email",
		Code:    "INVALID_EMAIL",
		Message: "email is required",
	}})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var payload struct {
		Meta   Meta        `json:"meta"`
		Errors []ErrorItem `json:"errors"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload.Meta.Status != "error" {
		t.Fatalf("expected meta status error, got %q", payload.Meta.Status)
	}

	if len(payload.Errors) != 1 || payload.Errors[0].Code != "INVALID_EMAIL" {
		t.Fatalf("expected INVALID_EMAIL error item, got %+v", payload.Errors)
	}
}
