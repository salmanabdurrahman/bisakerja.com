package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	growthapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/growth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/growth"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

// GrowthService defines behavior for growth service.
type GrowthService interface {
	CreateSavedSearch(ctx context.Context, input growthapp.CreateSavedSearchInput) (growth.SavedSearch, error)
	ListSavedSearches(ctx context.Context, userID string) ([]growth.SavedSearch, error)
	DeleteSavedSearch(ctx context.Context, userID string, savedSearchID string) error
	AddWatchlistCompany(ctx context.Context, userID string, companySlug string) (growth.CompanyWatchlist, error)
	ListWatchlistCompanies(ctx context.Context, userID string) ([]growth.CompanyWatchlist, error)
	RemoveWatchlistCompany(ctx context.Context, userID string, companySlug string) error
}

// GrowthHandler represents growth handler.
type GrowthHandler struct {
	service GrowthService
}

type createSavedSearchRequest struct {
	Query     string `json:"query"`
	Location  string `json:"location"`
	Source    string `json:"source"`
	SalaryMin *int64 `json:"salary_min"`
	Frequency string `json:"frequency"`
	IsActive  *bool  `json:"is_active"`
}

type createWatchlistCompanyRequest struct {
	CompanySlug string `json:"company_slug"`
}

// NewGrowthHandler creates a new growth handler instance.
func NewGrowthHandler(service GrowthService) *GrowthHandler {
	return &GrowthHandler{service: service}
}

// CreateSavedSearch creates saved search.
func (h *GrowthHandler) CreateSavedSearch(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	var request createSavedSearchRequest
	if err := decodeJSONBody(r, &request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	created, err := h.service.CreateSavedSearch(r.Context(), growthapp.CreateSavedSearchInput{
		UserID:    authUser.UserID,
		Query:     request.Query,
		Location:  request.Location,
		Source:    request.Source,
		SalaryMin: request.SalaryMin,
		Frequency: request.Frequency,
		IsActive:  isActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, growthapp.ErrInvalidSavedSearchQuery):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "query",
				Code:    errcode.BadRequest,
				Message: "query must be 2..200 characters",
			}})
		case errors.Is(err, growthapp.ErrInvalidSavedSearchLocation):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "location",
				Code:    errcode.BadRequest,
				Message: "location must be <= 100 characters",
			}})
		case errors.Is(err, growthapp.ErrInvalidSavedSearchSource):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "source",
				Code:    errcode.InvalidSource,
				Message: "source must be one of glints, kalibrr, jobstreet",
			}})
		case errors.Is(err, growthapp.ErrInvalidSavedSearchSalaryMin):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "salary_min",
				Code:    errcode.InvalidSalaryMin,
				Message: "salary_min must be >= 0",
			}})
		case errors.Is(err, growthapp.ErrInvalidSavedSearchFrequency):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "frequency",
				Code:    errcode.BadRequest,
				Message: "frequency must be one of instant, daily_digest, weekly_digest",
			}})
		case errors.Is(err, growth.ErrSavedSearchAlreadyExists):
			response.WriteError(w, http.StatusConflict, "Conflict", requestID, []response.ErrorItem{{
				Code:    errcode.BadRequest,
				Message: "saved search already exists",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to create saved search",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, "Saved search created", requestID, mapSavedSearch(created))
}

// ListSavedSearches returns a list of saved searches.
func (h *GrowthHandler) ListSavedSearches(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	items, err := h.service.ListSavedSearches(r.Context(), authUser.UserID)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to list saved searches",
		}})
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, mapSavedSearch(item))
	}
	response.WriteSuccess(w, http.StatusOK, "Saved searches retrieved", requestID, result)
}

// DeleteSavedSearch deletes saved search.
func (h *GrowthHandler) DeleteSavedSearch(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	savedSearchID := strings.TrimSpace(r.PathValue("id"))
	if savedSearchID == "" {
		response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
			Field:   "id",
			Code:    errcode.BadRequest,
			Message: "saved search id is required",
		}})
		return
	}

	if err := h.service.DeleteSavedSearch(r.Context(), authUser.UserID, savedSearchID); err != nil {
		switch {
		case errors.Is(err, growth.ErrSavedSearchNotFound):
			response.WriteError(w, http.StatusNotFound, "Not found", requestID, []response.ErrorItem{{
				Code:    errcode.NotFound,
				Message: "saved search not found",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to delete saved search",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Saved search deleted", requestID, map[string]any{
		"id": savedSearchID,
	})
}

// CreateWatchlistCompany creates watchlist company.
func (h *GrowthHandler) CreateWatchlistCompany(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	var request createWatchlistCompanyRequest
	if err := decodeJSONBody(r, &request); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body", requestID, []response.ErrorItem{{
			Code:    errcode.BadRequest,
			Message: "request body must be valid JSON",
		}})
		return
	}

	created, err := h.service.AddWatchlistCompany(r.Context(), authUser.UserID, request.CompanySlug)
	if err != nil {
		switch {
		case errors.Is(err, growthapp.ErrInvalidCompanySlug):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "company_slug",
				Code:    errcode.BadRequest,
				Message: "company_slug must be slug format (a-z, 0-9, -) length 2..80",
			}})
		case errors.Is(err, growth.ErrWatchlistCompanyAlreadyExists):
			response.WriteError(w, http.StatusConflict, "Conflict", requestID, []response.ErrorItem{{
				Code:    errcode.BadRequest,
				Message: "company already in watchlist",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to add company watchlist",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusCreated, "Company added to watchlist", requestID, mapWatchlistCompany(created))
}

// ListWatchlistCompanies returns a list of watchlist companies.
func (h *GrowthHandler) ListWatchlistCompanies(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	items, err := h.service.ListWatchlistCompanies(r.Context(), authUser.UserID)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to list watchlist companies",
		}})
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, mapWatchlistCompany(item))
	}
	response.WriteSuccess(w, http.StatusOK, "Watchlist companies retrieved", requestID, result)
}

// DeleteWatchlistCompany deletes watchlist company.
func (h *GrowthHandler) DeleteWatchlistCompany(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
			Code:    errcode.Unauthorized,
			Message: "authentication context missing",
		}})
		return
	}

	companySlug := strings.TrimSpace(r.PathValue("company_slug"))
	if companySlug == "" {
		response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
			Field:   "company_slug",
			Code:    errcode.BadRequest,
			Message: "company_slug is required",
		}})
		return
	}

	if err := h.service.RemoveWatchlistCompany(r.Context(), authUser.UserID, companySlug); err != nil {
		switch {
		case errors.Is(err, growthapp.ErrInvalidCompanySlug):
			response.WriteError(w, http.StatusBadRequest, "Validation error", requestID, []response.ErrorItem{{
				Field:   "company_slug",
				Code:    errcode.BadRequest,
				Message: "company_slug must be slug format (a-z, 0-9, -) length 2..80",
			}})
		case errors.Is(err, growth.ErrWatchlistCompanyNotFound):
			response.WriteError(w, http.StatusNotFound, "Not found", requestID, []response.ErrorItem{{
				Code:    errcode.NotFound,
				Message: "watchlist company not found",
			}})
		case errors.Is(err, identity.ErrUserNotFound):
			response.WriteError(w, http.StatusUnauthorized, "Unauthorized", requestID, []response.ErrorItem{{
				Code:    errcode.Unauthorized,
				Message: "user not found",
			}})
		default:
			response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
				Code:    errcode.InternalServerError,
				Message: "failed to delete watchlist company",
			}})
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Watchlist company deleted", requestID, map[string]any{
		"company_slug": strings.ToLower(companySlug),
	})
}

func mapSavedSearch(item growth.SavedSearch) map[string]any {
	return map[string]any{
		"id":         item.ID,
		"query":      item.Query,
		"location":   item.Location,
		"source":     item.Source,
		"salary_min": item.SalaryMin,
		"frequency":  item.Frequency,
		"is_active":  item.IsActive,
		"created_at": item.CreatedAt,
		"updated_at": item.UpdatedAt,
	}
}

func mapWatchlistCompany(item growth.CompanyWatchlist) map[string]any {
	return map[string]any{
		"company_slug": item.CompanySlug,
		"created_at":   item.CreatedAt,
	}
}
