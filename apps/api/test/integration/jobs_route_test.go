package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/router"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/jobs"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/logger"
)

func TestJobsRoutes_ListAndDetail(t *testing.T) {
	repository := memory.NewJobsRepository()
	result, err := repository.UpsertMany(context.Background(), job.SourceKalibrr, []job.UpsertInput{
		{
			OriginalJobID: "k-100",
			Title:         "Backend Golang",
			Company:       "Bisakerja",
			Location:      "Remote",
			URL:           "https://example.com/jobs/k-100",
		},
	})
	if err != nil {
		t.Fatalf("seed jobs: %v", err)
	}

	jobsHandler := handler.NewJobsHandler(jobs.NewService(repository))
	appHandler := router.New(
		logger.New("test"),
		router.Dependencies{JobsHandler: jobsHandler},
	)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/jobs?source=kalibrr", nil)
	listResp := httptest.NewRecorder()
	appHandler.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected jobs list 200, got %d", listResp.Code)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/"+result.Inserted[0].ID, nil)
	detailResp := httptest.NewRecorder()
	appHandler.ServeHTTP(detailResp, detailReq)
	if detailResp.Code != http.StatusOK {
		t.Fatalf("expected jobs detail 200, got %d", detailResp.Code)
	}

	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/unknown-id", nil)
	notFoundResp := httptest.NewRecorder()
	appHandler.ServeHTTP(notFoundResp, notFoundReq)
	if notFoundResp.Code != http.StatusNotFound {
		t.Fatalf("expected jobs detail 404 for unknown id, got %d", notFoundResp.Code)
	}
}
