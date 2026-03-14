package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	aidomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/ai"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

type aiProviderStub struct {
	searchAssistantResult aidomain.SearchAssistantResult
	searchAssistantErr    error
	searchAssistantCalls  int
	jobFitSummaryResult   aidomain.JobFitSummaryResult
	jobFitSummaryErr      error
	jobFitSummaryCalls    int
	coverLetterResult     aidomain.CoverLetterDraftResult
	coverLetterErr        error
	coverLetterCalls      int
}

func (s *aiProviderStub) GenerateSearchAssistant(
	_ context.Context,
	_ aidomain.SearchAssistantInput,
) (aidomain.SearchAssistantResult, error) {
	s.searchAssistantCalls++
	if s.searchAssistantErr != nil {
		return aidomain.SearchAssistantResult{}, s.searchAssistantErr
	}
	return s.searchAssistantResult, nil
}

func (s *aiProviderStub) GenerateJobFitSummary(
	_ context.Context,
	_ aidomain.JobFitSummaryInput,
) (aidomain.JobFitSummaryResult, error) {
	s.jobFitSummaryCalls++
	if s.jobFitSummaryErr != nil {
		return aidomain.JobFitSummaryResult{}, s.jobFitSummaryErr
	}
	return s.jobFitSummaryResult, nil
}

func (s *aiProviderStub) GenerateCoverLetterDraft(
	_ context.Context,
	_ aidomain.CoverLetterDraftInput,
) (aidomain.CoverLetterDraftResult, error) {
	s.coverLetterCalls++
	if s.coverLetterErr != nil {
		return aidomain.CoverLetterDraftResult{}, s.coverLetterErr
	}
	return s.coverLetterResult, nil
}

func TestService_GenerateSearchAssistant_Success(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "ai-service@example.com",
		PasswordHash: "hash",
		Name:         "AI Service User",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	usageRepository := memory.NewAIRepository()
	provider := &aiProviderStub{
		searchAssistantResult: aidomain.SearchAssistantResult{
			SuggestedQuery:     "golang backend remote",
			SuggestedLocations: []string{"Jakarta", "Remote"},
			SuggestedJobTypes:  []string{"fulltime"},
			Summary:            "Prioritize backend opportunities with remote option.",
			Provider:           "openai_compatible",
			Model:              "gpt-test-model",
			TokensIn:           25,
			TokensOut:          18,
		},
	}
	service := NewService(identityRepository, memory.NewJobsRepository(), usageRepository, provider, Config{
		DailyQuotaFree:    2,
		DailyQuotaPremium: 10,
	})

	result, err := service.GenerateSearchAssistant(context.Background(), GenerateSearchAssistantInput{
		UserID: user.ID,
		Prompt: "find golang backend jobs",
		Context: SearchAssistantContext{
			Location: "Jakarta",
		},
	})
	if err != nil {
		t.Fatalf("generate search assistant: %v", err)
	}
	if result.Feature != aidomain.FeatureSearchAssistant {
		t.Fatalf("expected feature search_assistant, got %q", result.Feature)
	}
	if result.Tier != "free" {
		t.Fatalf("expected free tier, got %q", result.Tier)
	}
	if result.Quota.DailyQuota != 2 || result.Quota.Used != 1 || result.Quota.Remaining != 1 {
		t.Fatalf("unexpected quota payload: %+v", result.Quota)
	}
	if provider.searchAssistantCalls != 1 {
		t.Fatalf("expected provider to be called once, got %d", provider.searchAssistantCalls)
	}

	usage, err := service.GetUsage(context.Background(), GetUsageInput{
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("get usage: %v", err)
	}
	if usage.Quota.Used != 1 || usage.Quota.Remaining != 1 {
		t.Fatalf("unexpected usage snapshot: %+v", usage.Quota)
	}
}

func TestService_GenerateSearchAssistant_QuotaExceeded(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "ai-quota@example.com",
		PasswordHash: "hash",
		Name:         "AI Quota User",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	usageRepository := memory.NewAIRepository()
	provider := &aiProviderStub{
		searchAssistantResult: aidomain.SearchAssistantResult{
			SuggestedQuery: "golang backend remote",
			Summary:        "Use focused search terms.",
		},
	}
	service := NewService(identityRepository, memory.NewJobsRepository(), usageRepository, provider, Config{
		DailyQuotaFree:    1,
		DailyQuotaPremium: 5,
	})

	_, err = service.GenerateSearchAssistant(context.Background(), GenerateSearchAssistantInput{
		UserID: user.ID,
		Prompt: "golang backend jobs",
	})
	if err != nil {
		t.Fatalf("first generate should succeed: %v", err)
	}

	_, err = service.GenerateSearchAssistant(context.Background(), GenerateSearchAssistantInput{
		UserID: user.ID,
		Prompt: "golang backend jobs second try",
	})
	if !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
}

func TestService_GenerateSearchAssistant_PremiumTierQuota(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	expiry := time.Now().UTC().Add(24 * time.Hour)
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:            "ai-premium@example.com",
		PasswordHash:     "hash",
		Name:             "AI Premium User",
		Role:             identity.RoleUser,
		IsPremium:        true,
		PremiumExpiredAt: &expiry,
	})
	if err != nil {
		t.Fatalf("seed premium user: %v", err)
	}

	usageRepository := memory.NewAIRepository()
	provider := &aiProviderStub{
		searchAssistantResult: aidomain.SearchAssistantResult{
			SuggestedQuery: "senior golang backend",
			Summary:        "Focus on senior backend roles.",
		},
	}
	service := NewService(identityRepository, memory.NewJobsRepository(), usageRepository, provider, Config{
		DailyQuotaFree:    1,
		DailyQuotaPremium: 3,
	})

	result, err := service.GenerateSearchAssistant(context.Background(), GenerateSearchAssistantInput{
		UserID: user.ID,
		Prompt: "senior golang roles",
	})
	if err != nil {
		t.Fatalf("generate premium search assistant: %v", err)
	}
	if result.Tier != "premium" {
		t.Fatalf("expected premium tier, got %q", result.Tier)
	}
	if result.Quota.DailyQuota != 3 || result.Quota.Remaining != 2 {
		t.Fatalf("unexpected premium quota payload: %+v", result.Quota)
	}
}

func TestService_GenerateSearchAssistant_ProviderErrorMapping(t *testing.T) {
	testCases := []struct {
		name        string
		providerErr error
		expectedErr error
	}{
		{
			name:        "provider rate limited",
			providerErr: aidomain.ErrProviderRateLimited,
			expectedErr: ErrProviderRateLimited,
		},
		{
			name:        "provider upstream",
			providerErr: aidomain.ErrProviderUpstream,
			expectedErr: ErrProviderUpstream,
		},
		{
			name:        "provider unavailable",
			providerErr: aidomain.ErrProviderUnavailable,
			expectedErr: ErrServiceUnavailable,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			identityRepository := memory.NewIdentityRepository()
			user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
				Email:        "ai-provider-" + testCase.name + "@example.com",
				PasswordHash: "hash",
				Name:         "AI Provider User",
				Role:         identity.RoleUser,
			})
			if err != nil {
				t.Fatalf("seed user: %v", err)
			}

			service := NewService(identityRepository, memory.NewJobsRepository(), memory.NewAIRepository(), &aiProviderStub{
				searchAssistantErr: testCase.providerErr,
			}, Config{
				DailyQuotaFree:    5,
				DailyQuotaPremium: 10,
			})

			_, err = service.GenerateSearchAssistant(context.Background(), GenerateSearchAssistantInput{
				UserID: user.ID,
				Prompt: "golang backend jobs",
			})
			if !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("expected error %v, got %v", testCase.expectedErr, err)
			}
		})
	}
}

func TestService_GetUsage_InvalidFeature(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "ai-invalid-feature@example.com",
		PasswordHash: "hash",
		Name:         "AI Invalid Feature User",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	service := NewService(identityRepository, memory.NewJobsRepository(), memory.NewAIRepository(), &aiProviderStub{
		searchAssistantResult: aidomain.SearchAssistantResult{
			SuggestedQuery: "golang",
			Summary:        "golang",
		},
	}, Config{})

	_, err = service.GetUsage(context.Background(), GetUsageInput{
		UserID:  user.ID,
		Feature: "unsupported_feature",
	})
	if !errors.Is(err, ErrInvalidFeature) {
		t.Fatalf("expected ErrInvalidFeature, got %v", err)
	}
}

func TestService_GenerateJobFitSummary_Success(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	expiry := time.Now().UTC().Add(24 * time.Hour)
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:            "ai-job-fit@example.com",
		PasswordHash:     "hash",
		Name:             "AI Job Fit User",
		Role:             identity.RoleUser,
		IsPremium:        true,
		PremiumExpiredAt: &expiry,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	_, err = identityRepository.SavePreferences(context.Background(), identity.Preferences{
		UserID:    user.ID,
		Keywords:  []string{"golang", "backend", "microservices"},
		Locations: []string{"Jakarta", "Remote"},
		JobTypes:  []string{"fulltime"},
		SalaryMin: 15_000_000,
		AlertMode: identity.NotificationAlertModeInstant,
		UpdatedAt: ptrTime(time.Now().UTC()),
	})
	if err != nil {
		t.Fatalf("seed preferences: %v", err)
	}

	jobsRepository := memory.NewJobsRepository()
	upsertResult, err := jobsRepository.UpsertMany(context.Background(), job.SourceGlints, []job.UpsertInput{
		{
			OriginalJobID: "job-fit-1",
			Title:         "Senior Backend Engineer",
			Company:       "Acme Tech",
			Location:      "Jakarta",
			Description:   "Build scalable Go services and API integrations.",
			URL:           "https://example.com/jobs/job-fit-1",
		},
	})
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}
	if len(upsertResult.Inserted) != 1 {
		t.Fatalf("expected one inserted job, got %d", len(upsertResult.Inserted))
	}

	provider := &aiProviderStub{
		jobFitSummaryResult: aidomain.JobFitSummaryResult{
			FitScore:    82,
			Verdict:     "strong_match",
			Strengths:   []string{"Strong Go backend alignment"},
			Gaps:        []string{"Needs stronger event-driven architecture exposure"},
			NextActions: []string{"Highlight API scale achievements in CV"},
			Summary:     "Profile is strongly aligned with backend responsibilities.",
			Provider:    "openai_compatible",
			Model:       "gpt-test-model",
			TokensIn:    72,
			TokensOut:   55,
		},
	}
	service := NewService(identityRepository, jobsRepository, memory.NewAIRepository(), provider, Config{
		DailyQuotaFree:    1,
		DailyQuotaPremium: 3,
	})

	result, err := service.GenerateJobFitSummary(context.Background(), GenerateJobFitSummaryInput{
		UserID: user.ID,
		JobID:  upsertResult.Inserted[0].ID,
		Focus:  "prioritize backend architecture and api ownership",
	})
	if err != nil {
		t.Fatalf("generate job fit summary: %v", err)
	}
	if result.Feature != aidomain.FeatureJobFitSummary {
		t.Fatalf("expected feature job_fit_summary, got %q", result.Feature)
	}
	if result.Tier != "premium" {
		t.Fatalf("expected premium tier, got %q", result.Tier)
	}
	if result.FitScore != 82 || result.Verdict != "strong_match" {
		t.Fatalf("unexpected job fit result: %+v", result)
	}
	if result.Quota.DailyQuota != 3 || result.Quota.Used != 1 || result.Quota.Remaining != 2 {
		t.Fatalf("unexpected quota payload: %+v", result.Quota)
	}
	if provider.jobFitSummaryCalls != 1 {
		t.Fatalf("expected job-fit provider to be called once, got %d", provider.jobFitSummaryCalls)
	}

	usage, err := service.GetUsage(context.Background(), GetUsageInput{
		UserID:  user.ID,
		Feature: "job_fit_summary",
	})
	if err != nil {
		t.Fatalf("get usage: %v", err)
	}
	if usage.Quota.Used != 1 || usage.Feature != aidomain.FeatureJobFitSummary {
		t.Fatalf("unexpected usage snapshot: %+v", usage)
	}
}

func TestService_GenerateJobFitSummary_PremiumRequired(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:        "ai-job-fit-free@example.com",
		PasswordHash: "hash",
		Name:         "AI Free User",
		Role:         identity.RoleUser,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	jobsRepository := memory.NewJobsRepository()
	upsertResult, err := jobsRepository.UpsertMany(context.Background(), job.SourceGlints, []job.UpsertInput{
		{
			OriginalJobID: "job-fit-free-1",
			Title:         "Backend Engineer",
			Company:       "Acme",
			URL:           "https://example.com/jobs/job-fit-free-1",
		},
	})
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}

	service := NewService(identityRepository, jobsRepository, memory.NewAIRepository(), &aiProviderStub{
		jobFitSummaryResult: aidomain.JobFitSummaryResult{
			FitScore:  75,
			Verdict:   "moderate_match",
			Summary:   "Moderate match.",
			Provider:  "openai_compatible",
			Model:     "gpt-test-model",
			TokensIn:  20,
			TokensOut: 10,
		},
	}, Config{
		DailyQuotaFree:    1,
		DailyQuotaPremium: 2,
	})

	_, err = service.GenerateJobFitSummary(context.Background(), GenerateJobFitSummaryInput{
		UserID: user.ID,
		JobID:  upsertResult.Inserted[0].ID,
	})
	if !errors.Is(err, ErrPremiumRequired) {
		t.Fatalf("expected ErrPremiumRequired, got %v", err)
	}
}

func TestService_GenerateJobFitSummary_JobIDRequired(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:            "ai-job-fit-missing-job@example.com",
		PasswordHash:     "hash",
		Name:             "AI Missing Job ID",
		Role:             identity.RoleUser,
		IsPremium:        true,
		PremiumExpiredAt: ptrTime(time.Now().UTC().Add(24 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	service := NewService(identityRepository, memory.NewJobsRepository(), memory.NewAIRepository(), &aiProviderStub{}, Config{
		DailyQuotaFree:    1,
		DailyQuotaPremium: 2,
	})

	_, err = service.GenerateJobFitSummary(context.Background(), GenerateJobFitSummaryInput{
		UserID: user.ID,
		JobID:  " ",
	})
	if !errors.Is(err, ErrJobIDRequired) {
		t.Fatalf("expected ErrJobIDRequired, got %v", err)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func TestService_GenerateCoverLetterDraft_Success(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	expiry := time.Now().UTC().Add(24 * time.Hour)
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:            "ai-cover-letter@example.com",
		PasswordHash:     "hash",
		Name:             "Cover Letter User",
		Role:             identity.RoleUser,
		IsPremium:        true,
		PremiumExpiredAt: &expiry,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	jobsRepository := memory.NewJobsRepository()
	upsertResult, err := jobsRepository.UpsertMany(context.Background(), job.SourceGlints, []job.UpsertInput{
		{
			OriginalJobID: "cover-letter-1",
			Title:         "Backend Engineer",
			Company:       "Acme",
			Location:      "Jakarta",
			Description:   "Build and maintain Go APIs.",
			URL:           "https://example.com/jobs/cover-letter-1",
		},
	})
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}

	provider := &aiProviderStub{
		coverLetterResult: aidomain.CoverLetterDraftResult{
			Tone:      "professional",
			Draft:     "Dear Hiring Team, I am excited to apply for Backend Engineer role at Acme...",
			KeyPoints: []string{"5+ years Go backend", "API scalability", "production reliability"},
			Summary:   "Professional draft focused on backend impact.",
			Provider:  "openai_compatible",
			Model:     "gpt-test-model",
			TokensIn:  95,
			TokensOut: 140,
		},
	}
	service := NewService(identityRepository, jobsRepository, memory.NewAIRepository(), provider, Config{
		DailyQuotaFree:    1,
		DailyQuotaPremium: 3,
	})

	result, err := service.GenerateCoverLetterDraft(context.Background(), GenerateCoverLetterDraftInput{
		UserID:     user.ID,
		JobID:      upsertResult.Inserted[0].ID,
		Tone:       "professional",
		Highlights: []string{"Golang API architecture", "Observability leadership"},
	})
	if err != nil {
		t.Fatalf("generate cover letter draft: %v", err)
	}
	if result.Feature != aidomain.FeatureCoverLetterDraft {
		t.Fatalf("expected cover_letter_draft feature, got %q", result.Feature)
	}
	if result.Tier != "premium" {
		t.Fatalf("expected premium tier, got %q", result.Tier)
	}
	if result.Tone != "professional" || result.Draft == "" {
		t.Fatalf("unexpected cover letter payload: %+v", result)
	}
	if result.Quota.DailyQuota != 3 || result.Quota.Used != 1 || result.Quota.Remaining != 2 {
		t.Fatalf("unexpected quota payload: %+v", result.Quota)
	}
	if provider.coverLetterCalls != 1 {
		t.Fatalf("expected cover-letter provider called once, got %d", provider.coverLetterCalls)
	}
}

func TestService_GenerateCoverLetterDraft_InvalidTone(t *testing.T) {
	identityRepository := memory.NewIdentityRepository()
	user, err := identityRepository.CreateUser(context.Background(), identity.CreateUserInput{
		Email:            "ai-cover-invalid-tone@example.com",
		PasswordHash:     "hash",
		Name:             "Cover Invalid Tone",
		Role:             identity.RoleUser,
		IsPremium:        true,
		PremiumExpiredAt: ptrTime(time.Now().UTC().Add(24 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	service := NewService(identityRepository, memory.NewJobsRepository(), memory.NewAIRepository(), &aiProviderStub{}, Config{})
	_, err = service.GenerateCoverLetterDraft(context.Background(), GenerateCoverLetterDraftInput{
		UserID: user.ID,
		JobID:  "job_1",
		Tone:   "aggressive",
	})
	if !errors.Is(err, ErrInvalidTone) {
		t.Fatalf("expected ErrInvalidTone, got %v", err)
	}
}
