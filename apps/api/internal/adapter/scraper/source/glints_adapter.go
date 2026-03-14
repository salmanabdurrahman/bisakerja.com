package source

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

const defaultGlintsEndpoint = "https://glints.com/api/v2-alc/graphql?op=searchJobsV3"

// GlintsAdapter represents glints adapter.
type GlintsAdapter struct {
	Endpoint    string
	CountryCode string
	Client      *http.Client
}

// NewGlintsAdapter creates a new glints adapter instance.
func NewGlintsAdapter(client *http.Client) *GlintsAdapter {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &GlintsAdapter{
		Endpoint:    defaultGlintsEndpoint,
		CountryCode: "ID",
		Client:      client,
	}
}

// Source handles source.
func (a *GlintsAdapter) Source() job.Source {
	return job.SourceGlints
}

// RequiresAuth handles requires auth.
func (a *GlintsAdapter) RequiresAuth() bool {
	return false
}

// Fetch handles fetch.
func (a *GlintsAdapter) Fetch(ctx context.Context, request scraper.FetchRequest) (scraper.FetchResult, error) {
	endpoint := strings.TrimSpace(a.Endpoint)
	if endpoint == "" {
		endpoint = defaultGlintsEndpoint
	}

	payload := map[string]any{
		"operationName": "searchJobsV3",
		"variables": map[string]any{
			"data": map[string]any{
				"SearchTerm":          request.Keyword,
				"CountryCode":         a.CountryCode,
				"includeExternalJobs": true,
				"pageSize":            request.Limit,
				"page":                request.Page,
			},
		},
		"query": glintsSearchQuery,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return scraper.FetchResult{}, fmt.Errorf("marshal glints payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return scraper.FetchResult{}, fmt.Errorf("build glints request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://glints.com")
	req.Header.Set("Referer", "https://glints.com/id/opportunities/jobs/explore")
	req.Header.Set("x-glints-country-code", a.CountryCode)

	resp, err := a.Client.Do(req)
	if err != nil {
		return scraper.FetchResult{}, fmt.Errorf("execute glints request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return scraper.FetchResult{}, fmt.Errorf("glints returned status %d", resp.StatusCode)
	}

	var parsed struct {
		Data struct {
			SearchJobsV3 struct {
				JobsInPage []struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Description string `json:"description"`
					Company     struct {
						Name string `json:"name"`
					} `json:"company"`
					City struct {
						Name string `json:"name"`
					} `json:"city"`
					CreatedAt      string `json:"createdAt"`
					SalaryEstimate struct {
						MinAmount *int64 `json:"minAmount"`
						MaxAmount *int64 `json:"maxAmount"`
					} `json:"salaryEstimate"`
				} `json:"jobsInPage"`
				HasMore bool `json:"hasMore"`
			} `json:"searchJobsV3"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return scraper.FetchResult{}, fmt.Errorf("decode glints response: %w", err)
	}

	items := make([]job.UpsertInput, 0, len(parsed.Data.SearchJobsV3.JobsInPage))
	for _, row := range parsed.Data.SearchJobsV3.JobsInPage {
		postedAt := parseOptionalTime(row.CreatedAt)
		items = append(items, job.UpsertInput{
			OriginalJobID: row.ID,
			Title:         row.Title,
			Company:       row.Company.Name,
			Location:      row.City.Name,
			Description:   row.Description,
			URL:           buildGlintsURL(row.ID),
			SalaryMin:     row.SalaryEstimate.MinAmount,
			SalaryMax:     row.SalaryEstimate.MaxAmount,
			PostedAt:      postedAt,
			RawData:       map[string]any{},
		})
	}

	return scraper.FetchResult{
		Jobs:    items,
		HasMore: parsed.Data.SearchJobsV3.HasMore,
	}, nil
}

func buildGlintsURL(originalID string) string {
	return "https://glints.com/id/opportunities/jobs/" + strings.TrimSpace(originalID)
}

const glintsSearchQuery = `query searchJobsV3($data: JobSearchConditionInput!) {
  searchJobsV3(data: $data) {
    jobsInPage {
      id
      title
      description
      createdAt
      company { name }
      city { name }
      salaryEstimate { minAmount maxAmount }
    }
    hasMore
  }
}`
