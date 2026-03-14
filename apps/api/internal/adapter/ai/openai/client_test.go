package openai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	aidomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/ai"
)

func TestClient_GenerateSearchAssistant_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if authHeader := r.Header.Get("Authorization"); authHeader != "Bearer test-key" {
			t.Fatalf("expected bearer token, got %q", authHeader)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "gpt-test-model",
			"usage": map[string]any{
				"prompt_tokens":     50,
				"completion_tokens": 20,
			},
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "```json\n{\"query\":\"golang backend remote\",\"locations\":[\"Jakarta\",\"Remote\"],\"job_types\":[\"fulltime\"],\"salary_min\":15000000,\"summary\":\"Focus on remote-ready backend roles.\"}\n```",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL + "/v1",
		APIKey:  "test-key",
		Model:   "gpt-test-model",
	})

	result, err := client.GenerateSearchAssistant(context.Background(), aidomain.SearchAssistantInput{
		Prompt: "help me find golang backend jobs",
		Context: aidomain.SearchAssistantContext{
			Location: "Jakarta",
		},
	})
	if err != nil {
		t.Fatalf("generate search assistant: %v", err)
	}
	if result.Provider != "openai_compatible" {
		t.Fatalf("expected provider openai_compatible, got %q", result.Provider)
	}
	if result.Model != "gpt-test-model" {
		t.Fatalf("expected model gpt-test-model, got %q", result.Model)
	}
	if result.SuggestedQuery != "golang backend remote" {
		t.Fatalf("expected suggested query, got %q", result.SuggestedQuery)
	}
	if len(result.SuggestedLocations) != 2 {
		t.Fatalf("expected two suggested locations, got %#v", result.SuggestedLocations)
	}
	if len(result.SuggestedJobTypes) != 1 || result.SuggestedJobTypes[0] != "fulltime" {
		t.Fatalf("unexpected job types: %#v", result.SuggestedJobTypes)
	}
	if result.SuggestedSalaryMin == nil || *result.SuggestedSalaryMin != 15_000_000 {
		t.Fatalf("unexpected suggested salary min: %#v", result.SuggestedSalaryMin)
	}
	if result.Summary == "" {
		t.Fatal("expected summary to be populated")
	}
	if result.TokensIn != 50 || result.TokensOut != 20 {
		t.Fatalf("unexpected token usage: in=%d out=%d", result.TokensIn, result.TokensOut)
	}
}

func TestClient_GenerateSearchAssistant_ProviderErrors(t *testing.T) {
	testCases := []struct {
		name        string
		statusCode  int
		expectedErr error
	}{
		{
			name:        "rate limited",
			statusCode:  http.StatusTooManyRequests,
			expectedErr: aidomain.ErrProviderRateLimited,
		},
		{
			name:        "upstream invalid request",
			statusCode:  http.StatusBadRequest,
			expectedErr: aidomain.ErrProviderUpstream,
		},
		{
			name:        "provider unavailable",
			statusCode:  http.StatusInternalServerError,
			expectedErr: aidomain.ErrProviderUnavailable,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(testCase.statusCode)
				_, _ = w.Write([]byte(`{"error":{"message":"provider failed"}}`))
			}))
			defer server.Close()

			client := NewClient(ClientConfig{
				BaseURL: server.URL + "/v1",
				APIKey:  "test-key",
			})

			_, err := client.GenerateSearchAssistant(context.Background(), aidomain.SearchAssistantInput{
				Prompt: "golang jobs",
			})
			if !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("expected error %v, got %v", testCase.expectedErr, err)
			}
		})
	}
}

func TestClient_GenerateSearchAssistant_FallbackContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "Try searching for: backend golang remote",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL + "/v1",
		APIKey:  "test-key",
	})

	result, err := client.GenerateSearchAssistant(context.Background(), aidomain.SearchAssistantInput{
		Prompt: "golang backend remote",
	})
	if err != nil {
		t.Fatalf("generate search assistant with fallback content: %v", err)
	}
	if strings.TrimSpace(result.SuggestedQuery) == "" {
		t.Fatal("expected fallback suggested_query to be populated")
	}
	if strings.TrimSpace(result.Summary) == "" {
		t.Fatal("expected fallback summary to be populated")
	}
}
