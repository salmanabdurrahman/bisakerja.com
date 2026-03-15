package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/jobs"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/errcode"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/pkg/response"
)

var allowedSort = map[string]struct{}{
	"-posted_at":  {},
	"posted_at":   {},
	"-created_at": {},
	"created_at":  {},
}

// JobsHandler represents jobs handler.
type JobsHandler struct {
	service *jobs.Service
}

type queryValidationError struct {
	code    string
	message string
}

// Error returns the error message.
func (e queryValidationError) Error() string {
	return e.message
}

// NewJobsHandler creates a new jobs handler instance.
func NewJobsHandler(service *jobs.Service) *JobsHandler {
	return &JobsHandler{service: service}
}

// ListJobs returns a list of jobs.
func (h *JobsHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	query, err := parseSearchQuery(r)
	if err != nil {
		errorCode := errcode.BadRequest
		var validationErr queryValidationError
		if errors.As(err, &validationErr) {
			errorCode = validationErr.code
		}

		response.WriteError(w, http.StatusBadRequest, "Invalid query parameters", requestID, []response.ErrorItem{{
			Code:    errorCode,
			Message: err.Error(),
		}})
		return
	}

	result, serviceErr := h.service.Search(r.Context(), query)
	if serviceErr != nil {
		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to query jobs",
		}})
		return
	}

	response.WriteSuccessWithPagination(
		w,
		http.StatusOK,
		"Jobs retrieved",
		requestID,
		mapJobsList(result.Data),
		response.Pagination{
			Page:         result.Page,
			Limit:        result.Limit,
			TotalPages:   result.TotalPages,
			TotalRecords: result.TotalRecords,
		},
	)
}

// GetJobByID returns job by id.
func (h *JobsHandler) GetJobByID(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		response.WriteError(w, http.StatusNotFound, "Job not found", requestID, []response.ErrorItem{{
			Code:    errcode.NotFound,
			Message: "job id is required",
		}})
		return
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, job.ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "Job not found", requestID, []response.ErrorItem{{
				Code:    errcode.NotFound,
				Message: "job not found",
			}})
			return
		}

		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to load job",
		}})
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Job detail retrieved", requestID, mapJobDetail(item))
}

func parseSearchQuery(r *http.Request) (job.SearchQuery, error) {
	query := r.URL.Query()
	result := job.SearchQuery{
		Q:        strings.TrimSpace(query.Get("q")),
		Location: strings.TrimSpace(query.Get("location")),
		Sort:     strings.TrimSpace(query.Get("sort")),
		Page:     1,
		Limit:    20,
	}

	if len(result.Q) > 200 {
		return job.SearchQuery{}, queryValidationError{
			code:    errcode.BadRequest,
			message: "q must be 200 characters or less",
		}
	}
	if len(result.Location) > 100 {
		return job.SearchQuery{}, queryValidationError{
			code:    errcode.BadRequest,
			message: "location must be 100 characters or less",
		}
	}

	if rawPage := strings.TrimSpace(query.Get("page")); rawPage != "" {
		page, err := strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			return job.SearchQuery{}, queryValidationError{
				code:    errcode.InvalidPage,
				message: "page must be an integer >= 1",
			}
		}
		result.Page = page
	}

	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > 100 {
			return job.SearchQuery{}, queryValidationError{
				code:    errcode.InvalidLimit,
				message: "limit must be between 1 and 100",
			}
		}
		result.Limit = limit
	}

	if rawSalaryMin := strings.TrimSpace(query.Get("salary_min")); rawSalaryMin != "" {
		salary, err := strconv.ParseInt(rawSalaryMin, 10, 64)
		if err != nil || salary < 0 {
			return job.SearchQuery{}, queryValidationError{
				code:    errcode.InvalidSalaryMin,
				message: "salary_min must be an integer >= 0",
			}
		}
		result.SalaryMin = salary
		result.HasSalaryMin = true
	}

	if result.Sort == "" {
		result.Sort = "-posted_at"
	}
	if _, ok := allowedSort[result.Sort]; !ok {
		return job.SearchQuery{}, queryValidationError{
			code:    errcode.InvalidSort,
			message: "sort must be one of -posted_at, posted_at, -created_at, created_at",
		}
	}

	rawSource := strings.TrimSpace(query.Get("source"))
	if rawSource != "" {
		source, ok := job.ParseSource(rawSource)
		if !ok {
			return job.SearchQuery{}, queryValidationError{
				code:    errcode.InvalidSource,
				message: "source must be one of glints, kalibrr, jobstreet",
			}
		}
		result.Source = source
	}

	return result, nil
}

func mapJobsList(items []job.Job) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		row := map[string]any{
			"id":           item.ID,
			"title":        item.Title,
			"company":      item.Company,
			"location":     item.Location,
			"salary_range": item.SalaryRange,
			"source":       item.Source,
			"posted_at":    item.PostedAt,
		}
		if row["salary_range"] == "" {
			row["salary_range"] = salaryRangeFallback(item)
		}
		result = append(result, row)
	}
	return result
}

func mapJobDetail(item job.Job) map[string]any {
	salaryRange := item.SalaryRange
	if salaryRange == "" {
		salaryRange = salaryRangeFallback(item)
	}

	return map[string]any{
		"id":           item.ID,
		"title":        item.Title,
		"company":      item.Company,
		"location":     item.Location,
		"description":  item.Description,
		"salary_range": salaryRange,
		"source":       item.Source,
		"url":          item.URL,
		"posted_at":    item.PostedAt,
		"created_at":   item.CreatedAt,
	}
}

func salaryRangeFallback(item job.Job) string {
	if item.SalaryMin == nil && item.SalaryMax == nil {
		return ""
	}
	if item.SalaryMin != nil && item.SalaryMax != nil {
		if *item.SalaryMin == *item.SalaryMax {
			return strconv.FormatInt(*item.SalaryMin, 10)
		}
		return strconv.FormatInt(*item.SalaryMin, 10) + " - " + strconv.FormatInt(*item.SalaryMax, 10)
	}
	if item.SalaryMin != nil {
		return ">= " + strconv.FormatInt(*item.SalaryMin, 10)
	}
	return "<= " + strconv.FormatInt(*item.SalaryMax, 10)
}

// SearchTitles returns distinct job titles matching a prefix.
func (h *JobsHandler) SearchTitles(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestIDFromContext(r.Context())
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	// Limit q to 100 characters
	if len(q) > 100 {
		q = q[:100]
	}

	query := job.TitleSearchQuery{
		Q:     q,
		Limit: 10,
	}

	titles, err := h.service.SearchTitles(r.Context(), query)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Internal server error", requestID, []response.ErrorItem{{
			Code:    errcode.InternalServerError,
			Message: "failed to search job titles",
		}})
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Job titles retrieved", requestID, map[string]any{
		"titles": titles,
	})
}
