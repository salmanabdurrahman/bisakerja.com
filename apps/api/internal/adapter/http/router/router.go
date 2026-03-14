package router

import (
	"log/slog"
	"net/http"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

func New(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.Healthz)
	mux.HandleFunc("/readyz", handler.Readyz)
	mux.HandleFunc("/api/v1/healthz", handler.Healthz)
	mux.HandleFunc("/api/v1/readyz", handler.Readyz)

	return observability.RequestID(withRecovery(logger, mux))
}

func withRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID := observability.RequestIDFromContext(r.Context())
				logger.Error("panic recovered", "path", r.URL.Path, "request_id", requestID, "panic", recovered)
				response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
					Code:    errcode.InternalServerError,
					Message: "unexpected server error",
				}})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
