package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/adapter/persistence/memory"
	aidomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/ai"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/identity"
)

type aiProviderStub struct {
	result aidomain.SearchAssistantResult
	err    error
	calls  int
}

func (s *aiProviderStub) GenerateSearchAssistant(
	_ context.Context,
	_ aidomain.SearchAssistantInput,
) (aidomain.SearchAssistantResult, error) {
	s.calls++
	if s.err != nil {
		return aidomain.SearchAssistantResult{}, s.err
	}
	return s.result, nil
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
		result: aidomain.SearchAssistantResult{
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
	service := NewService(identityRepository, usageRepository, provider, Config{
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
	if provider.calls != 1 {
		t.Fatalf("expected provider to be called once, got %d", provider.calls)
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
		result: aidomain.SearchAssistantResult{
			SuggestedQuery: "golang backend remote",
			Summary:        "Use focused search terms.",
		},
	}
	service := NewService(identityRepository, usageRepository, provider, Config{
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
		result: aidomain.SearchAssistantResult{
			SuggestedQuery: "senior golang backend",
			Summary:        "Focus on senior backend roles.",
		},
	}
	service := NewService(identityRepository, usageRepository, provider, Config{
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

			service := NewService(identityRepository, memory.NewAIRepository(), &aiProviderStub{
				err: testCase.providerErr,
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

	service := NewService(identityRepository, memory.NewAIRepository(), &aiProviderStub{
		result: aidomain.SearchAssistantResult{
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
