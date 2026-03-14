package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/http/middleware"
	aiapp "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/ai"
	aidomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/ai"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/platform/observability"
)

type aiHandlerServiceStub struct {
	generateResult aiapp.SearchAssistantResult
	generateErr    error
	jobFitResult   aiapp.JobFitSummaryResult
	jobFitErr      error
	getUsageResult aiapp.UsageSnapshot
	getUsageErr    error
}

func (s *aiHandlerServiceStub) GenerateSearchAssistant(
	_ context.Context,
	_ aiapp.GenerateSearchAssistantInput,
) (aiapp.SearchAssistantResult, error) {
	if s.generateErr != nil {
		return aiapp.SearchAssistantResult{}, s.generateErr
	}
	return s.generateResult, nil
}

func (s *aiHandlerServiceStub) GenerateJobFitSummary(
	_ context.Context,
	_ aiapp.GenerateJobFitSummaryInput,
) (aiapp.JobFitSummaryResult, error) {
	if s.jobFitErr != nil {
		return aiapp.JobFitSummaryResult{}, s.jobFitErr
	}
	return s.jobFitResult, nil
}

func (s *aiHandlerServiceStub) GetUsage(_ context.Context, _ aiapp.GetUsageInput) (aiapp.UsageSnapshot, error) {
	if s.getUsageErr != nil {
		return aiapp.UsageSnapshot{}, s.getUsageErr
	}
	return s.getUsageResult, nil
}

func TestAIHandler_GenerateSearchAssistant_Success(t *testing.T) {
	resetAt := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	handler := NewAIHandler(&aiHandlerServiceStub{
		generateResult: aiapp.SearchAssistantResult{
			Feature:            aidomain.FeatureSearchAssistant,
			Prompt:             "golang backend jobs",
			SuggestedQuery:     "golang backend remote",
			SuggestedLocations: []string{"Jakarta", "Remote"},
			SuggestedJobTypes:  []string{"fulltime"},
			Summary:            "Prioritize backend roles with remote flexibility.",
			Tier:               "free",
			Provider:           "openai_compatible",
			Model:              "gpt-test-model",
			Quota: aiapp.UsageQuota{
				DailyQuota: 5,
				Used:       1,
				Remaining:  4,
				ResetAt:    resetAt,
			},
		},
	})

	requestBody := map[string]any{
		"prompt": "golang backend jobs",
		"context": map[string]any{
			"location": "Jakarta",
		},
	}
	body := bytes.Buffer{}
	if err := json.NewEncoder(&body).Encode(requestBody); err != nil {
		t.Fatalf("encode request body: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/ai/search-assistant", &body)
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_ai_generate_success"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: "usr_1",
		Role:   identity.RoleUser,
	}))

	responseRecorder := httptest.NewRecorder()
	handler.GenerateSearchAssistant(responseRecorder, request)
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (%s)", responseRecorder.Code, responseRecorder.Body.String())
	}

	var responsePayload struct {
		Data struct {
			Feature        string `json:"feature"`
			SuggestedQuery string `json:"suggested_query"`
			QuotaRemaining int    `json:"quota_remaining"`
			DailyQuota     int    `json:"daily_quota"`
		} `json:"data"`
	}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &responsePayload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if responsePayload.Data.Feature != "search_assistant" {
		t.Fatalf("expected feature search_assistant, got %q", responsePayload.Data.Feature)
	}
	if responsePayload.Data.SuggestedQuery != "golang backend remote" {
		t.Fatalf("expected suggested query, got %q", responsePayload.Data.SuggestedQuery)
	}
	if responsePayload.Data.DailyQuota != 5 || responsePayload.Data.QuotaRemaining != 4 {
		t.Fatalf("unexpected quota payload: %+v", responsePayload.Data)
	}
}

func TestAIHandler_GenerateSearchAssistant_ErrorMatrix(t *testing.T) {
	testCases := []struct {
		name         string
		serviceError error
		expectedCode int
		errorCode    string
	}{
		{
			name:         "invalid prompt",
			serviceError: aiapp.ErrPromptTooShort,
			expectedCode: http.StatusBadRequest,
			errorCode:    "INVALID_AI_PROMPT",
		},
		{
			name:         "invalid feature",
			serviceError: aiapp.ErrInvalidFeature,
			expectedCode: http.StatusBadRequest,
			errorCode:    "INVALID_AI_FEATURE",
		},
		{
			name:         "quota exceeded",
			serviceError: aiapp.ErrQuotaExceeded,
			expectedCode: http.StatusTooManyRequests,
			errorCode:    "AI_QUOTA_EXCEEDED",
		},
		{
			name:         "provider rate limited",
			serviceError: aiapp.ErrProviderRateLimited,
			expectedCode: http.StatusServiceUnavailable,
			errorCode:    "AI_PROVIDER_RATE_LIMITED",
		},
		{
			name:         "provider upstream",
			serviceError: aiapp.ErrProviderUpstream,
			expectedCode: http.StatusBadGateway,
			errorCode:    "AI_PROVIDER_UPSTREAM_ERROR",
		},
		{
			name:         "provider unavailable",
			serviceError: aiapp.ErrServiceUnavailable,
			expectedCode: http.StatusServiceUnavailable,
			errorCode:    "AI_PROVIDER_UNAVAILABLE",
		},
		{
			name:         "user not found",
			serviceError: identity.ErrUserNotFound,
			expectedCode: http.StatusUnauthorized,
			errorCode:    "UNAUTHORIZED",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewAIHandler(&aiHandlerServiceStub{
				generateErr: testCase.serviceError,
			})

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/ai/search-assistant",
				strings.NewReader(`{"prompt":"golang backend jobs"}`),
			)
			request.Header.Set("Content-Type", "application/json")
			request = request.WithContext(observability.WithRequestID(request.Context(), "req_ai_generate_error"))
			request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
				UserID: "usr_1",
				Role:   identity.RoleUser,
			}))

			responseRecorder := httptest.NewRecorder()
			handler.GenerateSearchAssistant(responseRecorder, request)
			if responseRecorder.Code != testCase.expectedCode {
				t.Fatalf("expected status %d, got %d (%s)", testCase.expectedCode, responseRecorder.Code, responseRecorder.Body.String())
			}
			if !strings.Contains(responseRecorder.Body.String(), testCase.errorCode) {
				t.Fatalf("expected response body to contain %q, got %s", testCase.errorCode, responseRecorder.Body.String())
			}
		})
	}
}

func TestAIHandler_GenerateJobFitSummary_Success(t *testing.T) {
	resetAt := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	handler := NewAIHandler(&aiHandlerServiceStub{
		jobFitResult: aiapp.JobFitSummaryResult{
			Feature:     aidomain.FeatureJobFitSummary,
			JobID:       "job_1",
			FitScore:    84,
			Verdict:     "strong_match",
			Strengths:   []string{"Strong backend API ownership"},
			Gaps:        []string{"Needs deeper distributed tracing examples"},
			NextActions: []string{"Add impact metrics from previous backend projects"},
			Summary:     "Profile strongly matches the job requirements.",
			Tier:        "premium",
			Provider:    "openai_compatible",
			Model:       "gpt-test-model",
			Quota: aiapp.UsageQuota{
				DailyQuota: 30,
				Used:       2,
				Remaining:  28,
				ResetAt:    resetAt,
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/ai/job-fit-summary", strings.NewReader(`{"job_id":"job_1","focus":"system design"}`))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_ai_job_fit_success"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: "usr_1",
		Role:   identity.RoleUser,
	}))

	responseRecorder := httptest.NewRecorder()
	handler.GenerateJobFitSummary(responseRecorder, request)
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (%s)", responseRecorder.Code, responseRecorder.Body.String())
	}

	var responsePayload struct {
		Data struct {
			Feature        string `json:"feature"`
			JobID          string `json:"job_id"`
			FitScore       int    `json:"fit_score"`
			Verdict        string `json:"verdict"`
			QuotaRemaining int    `json:"quota_remaining"`
		} `json:"data"`
	}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &responsePayload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if responsePayload.Data.Feature != "job_fit_summary" {
		t.Fatalf("expected feature job_fit_summary, got %q", responsePayload.Data.Feature)
	}
	if responsePayload.Data.JobID != "job_1" || responsePayload.Data.FitScore != 84 || responsePayload.Data.Verdict != "strong_match" {
		t.Fatalf("unexpected job fit response payload: %+v", responsePayload.Data)
	}
	if responsePayload.Data.QuotaRemaining != 28 {
		t.Fatalf("expected quota remaining 28, got %d", responsePayload.Data.QuotaRemaining)
	}
}

func TestAIHandler_GenerateJobFitSummary_ErrorMatrix(t *testing.T) {
	testCases := []struct {
		name         string
		serviceError error
		expectedCode int
		errorCode    string
	}{
		{
			name:         "missing job id",
			serviceError: aiapp.ErrJobIDRequired,
			expectedCode: http.StatusBadRequest,
			errorCode:    "BAD_REQUEST",
		},
		{
			name:         "premium required",
			serviceError: aiapp.ErrPremiumRequired,
			expectedCode: http.StatusForbidden,
			errorCode:    "FORBIDDEN",
		},
		{
			name:         "job not found",
			serviceError: fmt.Errorf("get job detail: %w", job.ErrNotFound),
			expectedCode: http.StatusNotFound,
			errorCode:    "NOT_FOUND",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewAIHandler(&aiHandlerServiceStub{
				jobFitErr: testCase.serviceError,
			})

			request := httptest.NewRequest(http.MethodPost, "/api/v1/ai/job-fit-summary", strings.NewReader(`{"job_id":"job_1"}`))
			request.Header.Set("Content-Type", "application/json")
			request = request.WithContext(observability.WithRequestID(request.Context(), "req_ai_job_fit_error"))
			request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
				UserID: "usr_1",
				Role:   identity.RoleUser,
			}))

			responseRecorder := httptest.NewRecorder()
			handler.GenerateJobFitSummary(responseRecorder, request)
			if responseRecorder.Code != testCase.expectedCode {
				t.Fatalf("expected status %d, got %d (%s)", testCase.expectedCode, responseRecorder.Code, responseRecorder.Body.String())
			}
			if !strings.Contains(responseRecorder.Body.String(), testCase.errorCode) {
				t.Fatalf("expected response body to contain %q, got %s", testCase.errorCode, responseRecorder.Body.String())
			}
		})
	}
}

func TestAIHandler_GetUsage_Success(t *testing.T) {
	handler := NewAIHandler(&aiHandlerServiceStub{
		getUsageResult: aiapp.UsageSnapshot{
			Feature: aidomain.FeatureSearchAssistant,
			Tier:    "premium",
			Quota: aiapp.UsageQuota{
				DailyQuota: 30,
				Used:       8,
				Remaining:  22,
				ResetAt:    time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
			},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ai/usage?feature=search_assistant", nil)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_ai_usage"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: "usr_1",
		Role:   identity.RoleUser,
	}))

	responseRecorder := httptest.NewRecorder()
	handler.GetUsage(responseRecorder, request)
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (%s)", responseRecorder.Code, responseRecorder.Body.String())
	}

	var responsePayload struct {
		Data struct {
			Feature    string `json:"feature"`
			Tier       string `json:"tier"`
			DailyQuota int    `json:"daily_quota"`
			Used       int    `json:"used"`
			Remaining  int    `json:"remaining"`
		} `json:"data"`
	}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &responsePayload); err != nil {
		t.Fatalf("decode usage response: %v", err)
	}
	if responsePayload.Data.Feature != "search_assistant" || responsePayload.Data.Tier != "premium" {
		t.Fatalf("unexpected usage response payload: %+v", responsePayload.Data)
	}
	if responsePayload.Data.DailyQuota != 30 || responsePayload.Data.Used != 8 || responsePayload.Data.Remaining != 22 {
		t.Fatalf("unexpected quota values: %+v", responsePayload.Data)
	}
}

func TestAIHandler_GetUsage_InvalidFeature(t *testing.T) {
	handler := NewAIHandler(&aiHandlerServiceStub{
		getUsageErr: aiapp.ErrInvalidFeature,
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ai/usage?feature=unsupported", nil)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_ai_usage_invalid_feature"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: "usr_1",
		Role:   identity.RoleUser,
	}))

	responseRecorder := httptest.NewRecorder()
	handler.GetUsage(responseRecorder, request)
	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d (%s)", responseRecorder.Code, responseRecorder.Body.String())
	}
	if !strings.Contains(responseRecorder.Body.String(), "INVALID_AI_FEATURE") {
		t.Fatalf("expected INVALID_AI_FEATURE in response, got %s", responseRecorder.Body.String())
	}
}

func TestAIHandler_GenerateSearchAssistant_DecodeRequestError(t *testing.T) {
	handler := NewAIHandler(&aiHandlerServiceStub{})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/ai/search-assistant", strings.NewReader(`{"prompt":`))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_ai_decode_error"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: "usr_1",
		Role:   identity.RoleUser,
	}))

	responseRecorder := httptest.NewRecorder()
	handler.GenerateSearchAssistant(responseRecorder, request)
	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d (%s)", responseRecorder.Code, responseRecorder.Body.String())
	}
}

func TestAIHandler_GetUsage_InternalError(t *testing.T) {
	handler := NewAIHandler(&aiHandlerServiceStub{
		getUsageErr: errors.New("internal"),
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ai/usage", nil)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req_ai_usage_internal_error"))
	request = request.WithContext(middleware.WithAuthUser(request.Context(), middleware.AuthUser{
		UserID: "usr_1",
		Role:   identity.RoleUser,
	}))

	responseRecorder := httptest.NewRecorder()
	handler.GetUsage(responseRecorder, request)
	if responseRecorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d (%s)", responseRecorder.Code, responseRecorder.Body.String())
	}
}
