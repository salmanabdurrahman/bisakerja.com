package handler

import (
	"net/http"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

type healthData struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	response.WriteSuccess(w, http.StatusOK, "Service healthy", requestID, healthData{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func Readyz(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	response.WriteSuccess(w, http.StatusOK, "Service ready", requestID, healthData{
		Status:    "ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
