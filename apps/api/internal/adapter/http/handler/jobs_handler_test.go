package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/jobs"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

func TestJobsHandler_ListJobs_ReturnsPaginatedResult(t *testing.T) {
	repository := memory.NewJobsRepository()
	service := jobs.NewService(repository)
	handler := NewJobsHandler(service)

	postedAt := time.Now().UTC()
	salaryMin := int64(11_000_000)
	_, err := repository.UpsertMany(context.Background(), job.SourceGlints, []job.UpsertInput{
		{
			OriginalJobID: "g-100",
			Title:         "Backend Engineer",
			Company:       "Acme",
			Location:      "Jakarta",
			Description:   "Golang",
			URL:           "https://example.com/jobs/g-100",
			SalaryMin:     &salaryMin,
			PostedAt:      &postedAt,
		},
	})
	if err != nil {
		t.Fatalf("seed jobs: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs?q=backend&source=glints&page=1&limit=20", nil)
	req = req.WithContext(observability.WithRequestID(req.Context(), "req_list_jobs"))
	recorder := httptest.NewRecorder()

	handler.ListJobs(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Meta struct {
			Status     string `json:"status"`
			Pagination struct {
				TotalRecords int `json:"total_records"`
			} `json:"pagination"`
		} `json:"meta"`
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload.Meta.Status != "success" {
		t.Fatalf("expected success status, got %s", payload.Meta.Status)
	}
	if payload.Meta.Pagination.TotalRecords != 1 {
		t.Fatalf("expected total_records 1, got %d", payload.Meta.Pagination.TotalRecords)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(payload.Data))
	}
}

func TestJobsHandler_ListJobs_InvalidLimit_ReturnsBadRequest(t *testing.T) {
	handler := NewJobsHandler(jobs.NewService(memory.NewJobsRepository()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs?limit=101", nil)
	req = req.WithContext(observability.WithRequestID(req.Context(), "req_invalid_limit"))
	recorder := httptest.NewRecorder()

	handler.ListJobs(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}
