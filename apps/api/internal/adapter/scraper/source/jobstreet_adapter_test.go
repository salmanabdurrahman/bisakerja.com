package source

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
)

func TestJobstreetAdapter_Fetch_RequiresToken(t *testing.T) {
	adapter := NewJobstreetAdapter(nil)

	_, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
		Keyword: "backend",
		Page:    1,
		Limit:   10,
		Token:   "",
	})
	if !errors.Is(err, scraper.ErrTokenUnavailable) {
		t.Fatalf("expected ErrTokenUnavailable, got %v", err)
	}
}

func TestJobstreetAdapter_Fetch_SetsBrowserLikeHeaders(t *testing.T) {
	t.Parallel()

	var capturedHeader http.Header
	var requestBody map[string]any
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		capturedHeader = request.Header.Clone()
		payload, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if err := json.Unmarshal(payload, &requestBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		return jsonHTTPResponse(request, `{"data":{"jobSearchV6":{"totalCount":0,"data":[]}}}`), nil
	})}

	adapter := NewJobstreetAdapter(client)
	adapter.Endpoint = "https://id.jobstreet.com/graphql"
	adapter.Cookie = "sol_id=sol-123; JobseekerSessionId=test"
	adapter.SeekSessionID = "session-id"
	adapter.SeekVisitorID = "visitor-id"

	_, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
		Keyword: "backend engineer",
		Page:    2,
		Limit:   32,
		Token:   "token-123",
	})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	if got := capturedHeader.Get("Accept"); got != defaultWildcardAccept {
		t.Fatalf("expected Accept %q, got %q", defaultWildcardAccept, got)
	}
	if got := capturedHeader.Get("Accept-Language"); got != "en-US,en;q=0.9" {
		t.Fatalf("expected Accept-Language %q, got %q", "en-US,en;q=0.9", got)
	}
	if got := capturedHeader.Get("User-Agent"); got != defaultBrowserUserAgent {
		t.Fatalf("expected User-Agent %q, got %q", defaultBrowserUserAgent, got)
	}
	if got := capturedHeader.Get("Referer"); got != "https://id.jobstreet.com/id/backend-engineer-jobs?page=2" {
		t.Fatalf("expected Referer %q, got %q", "https://id.jobstreet.com/id/backend-engineer-jobs?page=2", got)
	}
	if got := capturedHeader.Get("x-custom-features"); got != "application/features.seek.all+json" {
		t.Fatalf("expected x-custom-features %q, got %q", "application/features.seek.all+json", got)
	}
	if got := capturedHeader.Get("x-seek-ec-sessionid"); got != "session-id" {
		t.Fatalf("expected x-seek-ec-sessionid %q, got %q", "session-id", got)
	}
	if got := capturedHeader.Get("x-seek-ec-visitorid"); got != "visitor-id" {
		t.Fatalf("expected x-seek-ec-visitorid %q, got %q", "visitor-id", got)
	}
	if got := capturedHeader.Get("Cookie"); got != "sol_id=sol-123; JobseekerSessionId=test" {
		t.Fatalf("expected Cookie %q, got %q", "sol_id=sol-123; JobseekerSessionId=test", got)
	}

	query, _ := requestBody["query"].(string)
	if !strings.Contains(query, "JobSearchV6QueryInput") {
		t.Fatalf("expected updated query type in request body, got %s", query)
	}
	if strings.Contains(query, "jobSearchV6(params: $params, locale: $locale, timezone: $timezone)") {
		t.Fatalf("expected jobSearchV6 without locale/timezone arguments, got %s", query)
	}

	variables, ok := requestBody["variables"].(map[string]any)
	if !ok {
		t.Fatalf("expected variables object, got %#v", requestBody["variables"])
	}
	params, ok := variables["params"].(map[string]any)
	if !ok {
		t.Fatalf("expected params object, got %#v", variables["params"])
	}
	if got, _ := params["channel"].(string); got != "mobileWeb" {
		t.Fatalf("expected channel %q, got %q", "mobileWeb", got)
	}
	if got, _ := params["source"].(string); got != "FE_HOME" {
		t.Fatalf("expected source %q, got %q", "FE_HOME", got)
	}
	if got, _ := params["solId"].(string); got != "sol-123" {
		t.Fatalf("expected solId %q, got %q", "sol-123", got)
	}
	if got, _ := params["eventCaptureSessionId"].(string); got != "session-id" {
		t.Fatalf("expected eventCaptureSessionId %q, got %q", "session-id", got)
	}
	if got, _ := params["eventCaptureUserId"].(string); got != "visitor-id" {
		t.Fatalf("expected eventCaptureUserId %q, got %q", "visitor-id", got)
	}
	if got, _ := params["userSessionId"].(string); got != "session-id" {
		t.Fatalf("expected userSessionId %q, got %q", "session-id", got)
	}
}

func TestJobstreetAdapter_Fetch_IncludesUpstreamBodySnippetOnBadRequest(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		response := jsonHTTPResponse(request, `{"errors":[{"message":"PersistedQueryNotFound"}]}`)
		response.StatusCode = http.StatusBadRequest
		return response, nil
	})}

	adapter := NewJobstreetAdapter(client)

	_, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
		Keyword: "backend",
		Page:    1,
		Limit:   32,
		Token:   "token-123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "PersistedQueryNotFound") {
		t.Fatalf("expected upstream body snippet in error, got %v", err)
	}
}
