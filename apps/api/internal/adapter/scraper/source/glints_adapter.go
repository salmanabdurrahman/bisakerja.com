package source

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	Cookie      string
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
	countryCode := firstNonEmpty(a.CountryCode, "ID")

	payload := map[string]any{
		"operationName": "searchJobsV3",
		"variables": map[string]any{
			"data": map[string]any{
				"SearchTerm":          request.Keyword,
				"CountryCode":         countryCode,
				"includeExternalJobs": true,
				"pageSize":            request.Limit,
				"page":                request.Page,
			},
		},
		"query": glintsSearchQuery,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return scraper.FetchResult{}, scraper.WrapSourceError("marshal_request_payload", 0, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return scraper.FetchResult{}, scraper.WrapSourceError("build_request", 0, err)
	}
	req.Header.Set("Accept", defaultJSONAcceptHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://glints.com")
	req.Header.Set("Referer", "https://glints.com/id/opportunities/jobs/explore")
	req.Header.Set("User-Agent", defaultBrowserUserAgent)
	req.Header.Set("x-glints-country-code", countryCode)
	if cookie := strings.TrimSpace(a.Cookie); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		return scraper.FetchResult{}, scraper.WrapSourceError("execute_request", 0, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		snippet := readBodySnippet(resp.Body, 1_024)
		if snippet != "" {
			return scraper.FetchResult{}, scraper.WrapSourceError(
				"execute_request",
				resp.StatusCode,
				fmt.Errorf("unexpected upstream response: %s", snippet),
			)
		}

		return scraper.FetchResult{}, scraper.WrapSourceError("execute_request", resp.StatusCode, errors.New("unexpected upstream response"))
	}

	var parsed struct {
		Data struct {
			SearchJobsV3 struct {
				JobsInPage []struct {
					ID      string `json:"id"`
					Title   string `json:"title"`
					Company struct {
						Name string `json:"name"`
					} `json:"company"`
					City struct {
						Name string `json:"name"`
					} `json:"city"`
					Country struct {
						Name string `json:"name"`
					} `json:"country"`
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
		return scraper.FetchResult{}, scraper.WrapSourceError("decode_response", resp.StatusCode, err)
	}

	items := make([]job.UpsertInput, 0, len(parsed.Data.SearchJobsV3.JobsInPage))
	for _, row := range parsed.Data.SearchJobsV3.JobsInPage {
		postedAt := parseOptionalTime(row.CreatedAt)
		items = append(items, job.UpsertInput{
			OriginalJobID: row.ID,
			Title:         row.Title,
			Company:       row.Company.Name,
			Location:      firstNonEmpty(row.City.Name, row.Country.Name),
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
      workArrangementOption
      status
      createdAt
      updatedAt
      isActivelyHiring
      isHot
      isApplied
      shouldShowSalary
      educationLevel
      type
      fraudReportFlag
      salaryEstimate {
        minAmount
        maxAmount
        CurrencyCode
        __typename
      }
      company {
        id
        name
        logo
        status
        isVIP
        IndustryId
      }
      city {
        id
        name
      }
      country {
        code
        name
      }
      minYearsOfExperience
      maxYearsOfExperience
      source
      jobSource
      traceInfo
    }
    expInfo
    hasMore
  }
}`
