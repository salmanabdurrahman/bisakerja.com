package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	aidomain "github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/ai"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-4.1-mini"
)

const systemPrompt = `You are an AI assistant for job search optimization.
Return ONLY a valid JSON object with this shape:
{
  "query": "string",
  "locations": ["string"],
  "job_types": ["fulltime|parttime|contract|internship"],
  "salary_min": 0,
  "summary": "string"
}
Rules:
- Keep "query" concise and practical for job boards.
- Include only relevant locations/job_types; use empty arrays when unknown.
- Use salary_min as integer without separators, or omit/null when unknown.
- Never return markdown fences, explanations, or additional keys.`

// ClientConfig stores configuration values for OpenAI-compatible client.
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	HTTPClient *http.Client
}

// Client represents OpenAI-compatible provider adapter.
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewClient creates a new OpenAI-compatible client instance.
func NewClient(config ClientConfig) *Client {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	} else if httpClient.Timeout <= 0 {
		httpClient.Timeout = timeout
	}

	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	model := strings.TrimSpace(config.Model)
	if model == "" {
		model = defaultModel
	}

	return &Client{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(config.APIKey),
		model:      model,
		httpClient: httpClient,
	}
}

// GenerateSearchAssistant generates search assistant suggestion from AI provider.
func (c *Client) GenerateSearchAssistant(
	ctx context.Context,
	input aidomain.SearchAssistantInput,
) (aidomain.SearchAssistantResult, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return aidomain.SearchAssistantResult{}, fmt.Errorf("%w: ai provider api key is empty", aidomain.ErrProviderUnavailable)
	}

	requestPayload := map[string]any{
		"model":       c.model,
		"temperature": 0.2,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": buildUserPrompt(input),
			},
		},
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return aidomain.SearchAssistantResult{}, fmt.Errorf("%w: encode request payload", aidomain.ErrProviderUpstream)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return aidomain.SearchAssistantResult{}, fmt.Errorf("%w: build request", aidomain.ErrProviderUpstream)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return aidomain.SearchAssistantResult{}, fmt.Errorf("%w: execute request: %v", aidomain.ErrProviderUnavailable, err)
	}

	responseBody, err := readBody(response.Body)
	if err != nil {
		return aidomain.SearchAssistantResult{}, fmt.Errorf("%w: read response body: %v", aidomain.ErrProviderUnavailable, err)
	}

	if response.StatusCode == http.StatusTooManyRequests {
		return aidomain.SearchAssistantResult{}, fmt.Errorf(
			"%w: provider returned status %d (%s)",
			aidomain.ErrProviderRateLimited,
			response.StatusCode,
			summarizeBody(responseBody),
		)
	}
	if response.StatusCode >= http.StatusInternalServerError {
		return aidomain.SearchAssistantResult{}, fmt.Errorf(
			"%w: provider returned status %d (%s)",
			aidomain.ErrProviderUnavailable,
			response.StatusCode,
			summarizeBody(responseBody),
		)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return aidomain.SearchAssistantResult{}, fmt.Errorf(
			"%w: provider returned status %d (%s)",
			aidomain.ErrProviderUpstream,
			response.StatusCode,
			summarizeBody(responseBody),
		)
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return aidomain.SearchAssistantResult{}, fmt.Errorf("%w: invalid response JSON", aidomain.ErrProviderUpstream)
	}
	if len(completion.Choices) == 0 {
		return aidomain.SearchAssistantResult{}, fmt.Errorf("%w: empty choices", aidomain.ErrProviderUpstream)
	}

	content := extractMessageContent(completion.Choices[0].Message.Content)
	if content == "" {
		return aidomain.SearchAssistantResult{}, fmt.Errorf("%w: empty completion content", aidomain.ErrProviderUpstream)
	}

	parsed, ok := parseAssistantPayload(content)
	if !ok {
		parsed = aidomain.SearchAssistantResult{
			SuggestedQuery: truncate(cleanSingleLine(content), 200),
			Summary:        truncate(content, 600),
		}
	}
	if strings.TrimSpace(parsed.SuggestedQuery) == "" {
		parsed.SuggestedQuery = strings.TrimSpace(input.Prompt)
	}
	if strings.TrimSpace(parsed.Summary) == "" {
		parsed.Summary = "Search suggestion generated."
	}

	parsed.Provider = "openai_compatible"
	parsed.Model = strings.TrimSpace(completion.Model)
	if parsed.Model == "" {
		parsed.Model = c.model
	}
	parsed.TokensIn = max(completion.Usage.PromptTokens, 0)
	parsed.TokensOut = max(completion.Usage.CompletionTokens, 0)
	return parsed, nil
}

type chatCompletionResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content any `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func buildUserPrompt(input aidomain.SearchAssistantInput) string {
	builder := strings.Builder{}
	builder.WriteString("User intent:\n")
	builder.WriteString(strings.TrimSpace(input.Prompt))

	if location := strings.TrimSpace(input.Context.Location); location != "" {
		builder.WriteString("\n\nPreferred location: ")
		builder.WriteString(location)
	}
	if len(input.Context.JobTypes) > 0 {
		builder.WriteString("\nPreferred job types: ")
		builder.WriteString(strings.Join(input.Context.JobTypes, ", "))
	}
	if input.Context.SalaryMin != nil && *input.Context.SalaryMin >= 0 {
		builder.WriteString("\nMinimum salary: ")
		builder.WriteString(strconv.FormatInt(*input.Context.SalaryMin, 10))
	}

	return builder.String()
}

func parseAssistantPayload(raw string) (aidomain.SearchAssistantResult, bool) {
	jsonCandidate := extractJSONCandidate(raw)
	if jsonCandidate == "" {
		return aidomain.SearchAssistantResult{}, false
	}

	payload := map[string]any{}
	if err := json.Unmarshal([]byte(jsonCandidate), &payload); err != nil {
		return aidomain.SearchAssistantResult{}, false
	}

	query := pickString(payload, "query", "suggested_query")
	summary := pickString(payload, "summary", "reasoning", "rationale")
	if query == "" && summary == "" {
		return aidomain.SearchAssistantResult{}, false
	}

	return aidomain.SearchAssistantResult{
		SuggestedQuery:     query,
		SuggestedLocations: pickStringList(payload, "locations", "suggested_locations"),
		SuggestedJobTypes:  pickStringList(payload, "job_types", "suggested_job_types"),
		SuggestedSalaryMin: pickInt64Pointer(payload, "salary_min", "suggested_salary_min"),
		Summary:            summary,
	}, true
}

func extractJSONCandidate(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```JSON")
		trimmed = strings.TrimPrefix(trimmed, "```")
		if endFence := strings.LastIndex(trimmed, "```"); endFence >= 0 {
			trimmed = trimmed[:endFence]
		}
		trimmed = strings.TrimSpace(trimmed)
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start < 0 || end < 0 || end <= start {
		return ""
	}
	return strings.TrimSpace(trimmed[start : end+1])
}

func extractMessageContent(content any) string {
	switch typed := content.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			entry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			text := strings.TrimSpace(toString(entry["text"]))
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	default:
		return ""
	}
}

func pickString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		normalized := strings.TrimSpace(toString(value))
		if normalized != "" {
			return normalized
		}
	}
	return ""
}

func pickStringList(payload map[string]any, keys ...string) []string {
	result := []string{}
	seen := map[string]struct{}{}
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}

		switch typed := value.(type) {
		case []any:
			for _, item := range typed {
				normalized := strings.TrimSpace(toString(item))
				if normalized == "" {
					continue
				}
				lower := strings.ToLower(normalized)
				if _, exists := seen[lower]; exists {
					continue
				}
				seen[lower] = struct{}{}
				result = append(result, normalized)
			}
		case []string:
			for _, item := range typed {
				normalized := strings.TrimSpace(item)
				if normalized == "" {
					continue
				}
				lower := strings.ToLower(normalized)
				if _, exists := seen[lower]; exists {
					continue
				}
				seen[lower] = struct{}{}
				result = append(result, normalized)
			}
		case string:
			for _, part := range strings.Split(typed, ",") {
				normalized := strings.TrimSpace(part)
				if normalized == "" {
					continue
				}
				lower := strings.ToLower(normalized)
				if _, exists := seen[lower]; exists {
					continue
				}
				seen[lower] = struct{}{}
				result = append(result, normalized)
			}
		}
	}
	return result
}

func pickInt64Pointer(payload map[string]any, keys ...string) *int64 {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		parsed, parsedOK := toInt64(value)
		if !parsedOK || parsed < 0 {
			continue
		}
		cloned := parsed
		return &cloned
	}
	return nil
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	default:
		return ""
	}
}

func toInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	case float64:
		return int64(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func summarizeBody(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "empty body"
	}
	return truncate(cleanSingleLine(trimmed), 220)
}

func cleanSingleLine(value string) string {
	compact := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	return strings.TrimSpace(compact)
}

func truncate(value string, maxChars int) string {
	if maxChars <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= maxChars {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:maxChars]))
}

func readBody(body io.ReadCloser) ([]byte, error) {
	defer func() {
		_ = body.Close()
	}()
	return io.ReadAll(body)
}
