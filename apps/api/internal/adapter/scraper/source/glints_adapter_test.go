package source

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
)

func TestGlintsAdapter_Fetch_SetsReferenceHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		cookie         string
		expectCookie   bool
		expectedCookie string
	}{
		{
			name:         "without optional cookie",
			cookie:       "",
			expectCookie: false,
		},
		{
			name:           "with optional cookie",
			cookie:         "session=abc123",
			expectCookie:   true,
			expectedCookie: "session=abc123",
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedHeader http.Header
			client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				capturedHeader = request.Header.Clone()
				return jsonHTTPResponse(request, `{"data":{"searchJobsV3":{"jobsInPage":[],"hasMore":false}}}`), nil
			})}

			adapter := NewGlintsAdapter(client)
			adapter.Endpoint = "https://glints.example.test/api/v2-alc/graphql?op=searchJobsV3"
			adapter.Cookie = testCase.cookie

			_, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
				Keyword: "backend",
				Page:    1,
				Limit:   30,
			})
			if err != nil {
				t.Fatalf("fetch: %v", err)
			}

			if got := capturedHeader.Get("Accept"); got != defaultJSONAcceptHeader {
				t.Fatalf("expected Accept %q, got %q", defaultJSONAcceptHeader, got)
			}
			if got := capturedHeader.Get("User-Agent"); got != defaultBrowserUserAgent {
				t.Fatalf("expected User-Agent %q, got %q", defaultBrowserUserAgent, got)
			}
			if got := capturedHeader.Get("Origin"); got != "https://glints.com" {
				t.Fatalf("expected Origin %q, got %q", "https://glints.com", got)
			}
			if got := capturedHeader.Get("Referer"); got != "https://glints.com/id/opportunities/jobs/explore" {
				t.Fatalf("expected Referer %q, got %q", "https://glints.com/id/opportunities/jobs/explore", got)
			}
			if got := capturedHeader.Get("x-glints-country-code"); got != "ID" {
				t.Fatalf("expected x-glints-country-code %q, got %q", "ID", got)
			}

			cookie := capturedHeader.Get("Cookie")
			if testCase.expectCookie && cookie != testCase.expectedCookie {
				t.Fatalf("expected Cookie %q, got %q", testCase.expectedCookie, cookie)
			}
			if !testCase.expectCookie && cookie != "" {
				t.Fatalf("expected Cookie header to be omitted, got %q", cookie)
			}
		})
	}
}

func TestGlintsAdapter_Fetch_UsesReferenceQueryShape(t *testing.T) {
	t.Parallel()

	var requestBody string
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		payload, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		requestBody = string(payload)

		return jsonHTTPResponse(request, `{
			"data": {
				"searchJobsV3": {
					"jobsInPage": [
						{
							"id": "job-1",
							"title": "Backend Engineer",
							"createdAt": "2026-03-15T00:00:00Z",
							"company": {"name": "Acme"},
							"city": {"name": "Jakarta"},
							"country": {"name": "Indonesia"},
							"salaryEstimate": {"minAmount": 10000000, "maxAmount": 20000000},
							"traceInfo": "trace-1"
						}
					],
					"hasMore": false
				}
			}
		}`), nil
	})}

	adapter := NewGlintsAdapter(client)
	adapter.Endpoint = "https://glints.example.test/api/v2-alc/graphql?op=searchJobsV3"

	result, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
		Keyword: "backend",
		Page:    1,
		Limit:   30,
	})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(result.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(result.Jobs))
	}
	if result.Jobs[0].Title != "Backend Engineer" {
		t.Fatalf("expected title %q, got %q", "Backend Engineer", result.Jobs[0].Title)
	}
	if result.Jobs[0].Location != "Jakarta" {
		t.Fatalf("expected location %q, got %q", "Jakarta", result.Jobs[0].Location)
	}
	if strings.Contains(requestBody, "description") {
		t.Fatalf("expected reference query without description field, got payload %s", requestBody)
	}
	if !strings.Contains(requestBody, "traceInfo") {
		t.Fatalf("expected reference query to include traceInfo, got payload %s", requestBody)
	}
}

func TestGlintsAdapter_Fetch_IncludesUpstreamBodySnippetOnBadRequest(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		response := jsonHTTPResponse(request, `{"error":"forbidden by upstream gateway"}`)
		response.StatusCode = http.StatusForbidden
		return response, nil
	})}

	adapter := NewGlintsAdapter(client)

	_, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
		Keyword: "backend",
		Page:    1,
		Limit:   30,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "forbidden by upstream gateway") {
		t.Fatalf("expected upstream response snippet in error, got %v", err)
	}
}
