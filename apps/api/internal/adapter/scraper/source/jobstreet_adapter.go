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

const defaultJobstreetEndpoint = "https://id.jobstreet.com/graphql"

type JobstreetAdapter struct {
	Endpoint string
	Client   *http.Client
	SiteKey  string
	Locale   string
	Timezone string
}

func NewJobstreetAdapter(client *http.Client) *JobstreetAdapter {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &JobstreetAdapter{
		Endpoint: defaultJobstreetEndpoint,
		Client:   client,
		SiteKey:  "ID",
		Locale:   "id-ID",
		Timezone: "Asia/Jakarta",
	}
}

func (a *JobstreetAdapter) Source() job.Source {
	return job.SourceJobstreet
}

func (a *JobstreetAdapter) RequiresAuth() bool {
	return true
}

func (a *JobstreetAdapter) Fetch(ctx context.Context, request scraper.FetchRequest) (scraper.FetchResult, error) {
	if strings.TrimSpace(request.Token) == "" {
		return scraper.FetchResult{}, fmt.Errorf("%w: missing bearer token", scraper.ErrTokenUnavailable)
	}

	endpoint := strings.TrimSpace(a.Endpoint)
	if endpoint == "" {
		endpoint = defaultJobstreetEndpoint
	}

	payload := map[string]any{
		"operationName": "JobSearchV6",
		"variables": map[string]any{
			"params": map[string]any{
				"keywords": request.Keyword,
				"page":     request.Page,
				"pageSize": request.Limit,
				"siteKey":  a.SiteKey,
				"locale":   a.Locale,
			},
			"locale":   a.Locale,
			"timezone": a.Timezone,
		},
		"query": jobstreetSearchQuery,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return scraper.FetchResult{}, fmt.Errorf("marshal jobstreet payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return scraper.FetchResult{}, fmt.Errorf("build jobstreet request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(request.Token))
	req.Header.Set("Origin", "https://id.jobstreet.com")
	req.Header.Set("Referer", "https://id.jobstreet.com")
	req.Header.Set("seek-request-brand", "jobstreet")
	req.Header.Set("seek-request-country", a.SiteKey)
	req.Header.Set("x-seek-site", "chalice")

	resp, err := a.Client.Do(req)
	if err != nil {
		return scraper.FetchResult{}, fmt.Errorf("execute jobstreet request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return scraper.FetchResult{}, fmt.Errorf("%w: jobstreet returned %d", scraper.ErrSourceUnauthorized, resp.StatusCode)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return scraper.FetchResult{}, fmt.Errorf("jobstreet returned status %d", resp.StatusCode)
	}

	var parsed struct {
		Data struct {
			JobSearchV6 struct {
				TotalCount int `json:"totalCount"`
				Data       []struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					CompanyName string `json:"companyName"`
					Teaser      string `json:"teaser"`
					SalaryLabel string `json:"salaryLabel"`
					Locations   []struct {
						Label string `json:"label"`
					} `json:"locations"`
					ListingDate struct {
						DateTimeUTC string `json:"dateTimeUtc"`
					} `json:"listingDate"`
				} `json:"data"`
			} `json:"jobSearchV6"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return scraper.FetchResult{}, fmt.Errorf("decode jobstreet response: %w", err)
	}
	if len(parsed.Errors) > 0 {
		return scraper.FetchResult{}, errors.New("jobstreet graphql returned errors")
	}

	items := make([]job.UpsertInput, 0, len(parsed.Data.JobSearchV6.Data))
	for _, row := range parsed.Data.JobSearchV6.Data {
		postedAt := parseOptionalTime(row.ListingDate.DateTimeUTC)
		location := ""
		for _, locationRow := range row.Locations {
			if strings.TrimSpace(locationRow.Label) != "" {
				location = strings.TrimSpace(locationRow.Label)
				break
			}
		}

		items = append(items, job.UpsertInput{
			OriginalJobID: row.ID,
			Title:         row.Title,
			Company:       row.CompanyName,
			Location:      location,
			Description:   row.Teaser,
			URL:           buildJobstreetURL(row.ID),
			SalaryRange:   row.SalaryLabel,
			PostedAt:      postedAt,
			RawData:       map[string]any{},
		})
	}

	hasMore := request.Page*request.Limit < parsed.Data.JobSearchV6.TotalCount

	return scraper.FetchResult{
		Jobs:    items,
		HasMore: hasMore,
	}, nil
}

func buildJobstreetURL(originalID string) string {
	return "https://id.jobstreet.com/id/job/" + strings.TrimSpace(originalID)
}

const jobstreetSearchQuery = `query JobSearchV6($params: JobSearchParamsInput!, $locale: String!, $timezone: String!) {
  jobSearchV6(params: $params, locale: $locale, timezone: $timezone) {
    totalCount
    data {
      id
      title
      companyName
      teaser
      salaryLabel
      locations { label }
      listingDate { dateTimeUtc }
    }
  }
}`
