package router

import (
	"log/slog"
	"net/http"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/handler"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

// Dependencies represents dependencies.
type Dependencies struct {
	JobsHandler         *handler.JobsHandler
	AuthHandler         *handler.AuthHandler
	PreferencesHandler  *handler.PreferencesHandler
	BillingHandler      *handler.BillingHandler
	GrowthHandler       *handler.GrowthHandler
	NotificationHandler *handler.NotificationHandler
	AuthMiddleware      *middleware.Authenticator
}

// New creates a new instance.
func New(logger *slog.Logger, dependencies ...Dependencies) http.Handler {
	var deps Dependencies
	if len(dependencies) > 0 {
		deps = dependencies[0]
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.Healthz)
	mux.HandleFunc("/readyz", handler.Readyz)
	mux.HandleFunc("/api/v1/healthz", handler.Healthz)
	mux.HandleFunc("/api/v1/readyz", handler.Readyz)

	if deps.AuthHandler != nil {
		mux.HandleFunc("POST /api/v1/auth/register", deps.AuthHandler.Register)
		mux.HandleFunc("POST /api/v1/auth/login", deps.AuthHandler.Login)
		mux.HandleFunc("POST /api/v1/auth/refresh", deps.AuthHandler.Refresh)
		if deps.AuthMiddleware != nil {
			mux.Handle("GET /api/v1/auth/me", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.AuthHandler.Me)))
		}
	}

	if deps.JobsHandler != nil {
		mux.HandleFunc("GET /api/v1/jobs", deps.JobsHandler.ListJobs)
		mux.HandleFunc("GET /api/v1/jobs/{id}", deps.JobsHandler.GetJobByID)
	}
	if deps.PreferencesHandler != nil && deps.AuthMiddleware != nil {
		mux.Handle("GET /api/v1/preferences", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PreferencesHandler.GetPreferences)))
		mux.Handle("PUT /api/v1/preferences", deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PreferencesHandler.UpdatePreferences)))
		mux.Handle(
			"PUT /api/v1/preferences/notification",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.PreferencesHandler.UpdateNotificationPreferences)),
		)
	}
	if deps.BillingHandler != nil && deps.AuthMiddleware != nil {
		mux.Handle(
			"POST /api/v1/billing/checkout-session",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.BillingHandler.CreateCheckoutSession)),
		)
		mux.Handle(
			"GET /api/v1/billing/status",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.BillingHandler.GetBillingStatus)),
		)
		mux.Handle(
			"GET /api/v1/billing/transactions",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.BillingHandler.GetBillingTransactions)),
		)
	}
	if deps.BillingHandler != nil {
		mux.HandleFunc("POST /api/v1/webhook/mayar", deps.BillingHandler.HandleMayarWebhook)
	}
	if deps.GrowthHandler != nil && deps.AuthMiddleware != nil {
		mux.Handle(
			"POST /api/v1/saved-searches",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.GrowthHandler.CreateSavedSearch)),
		)
		mux.Handle(
			"GET /api/v1/saved-searches",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.GrowthHandler.ListSavedSearches)),
		)
		mux.Handle(
			"DELETE /api/v1/saved-searches/{id}",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.GrowthHandler.DeleteSavedSearch)),
		)
		mux.Handle(
			"POST /api/v1/watchlist/companies",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.GrowthHandler.CreateWatchlistCompany)),
		)
		mux.Handle(
			"GET /api/v1/watchlist/companies",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.GrowthHandler.ListWatchlistCompanies)),
		)
		mux.Handle(
			"DELETE /api/v1/watchlist/companies/{company_slug}",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.GrowthHandler.DeleteWatchlistCompany)),
		)
	}
	if deps.NotificationHandler != nil && deps.AuthMiddleware != nil {
		mux.Handle(
			"GET /api/v1/notifications",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.NotificationHandler.ListNotifications)),
		)
		mux.Handle(
			"PATCH /api/v1/notifications/{id}/read",
			deps.AuthMiddleware.RequireAuth(http.HandlerFunc(deps.NotificationHandler.MarkNotificationRead)),
		)
	}

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
