package source

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

const defaultKalibrrEndpoint = "https://www.kalibrr.id/kjs/job_board/search"

type KalibrrAdapter struct {
	Endpoint string
	Client   *http.Client
}

func NewKalibrrAdapter(client *http.Client) *KalibrrAdapter {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &KalibrrAdapter{
		Endpoint: defaultKalibrrEndpoint,
		Client:   client,
	}
}

func (a *KalibrrAdapter) Source() job.Source {
	return job.SourceKalibrr
}

func (a *KalibrrAdapter) RequiresAuth() bool {
	return false
}

func (a *KalibrrAdapter) Fetch(ctx context.Context, request scraper.FetchRequest) (scraper.FetchResult, error) {
	endpoint := strings.TrimSpace(a.Endpoint)
	if endpoint == "" {
		endpoint = defaultKalibrrEndpoint
	}

	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", request.Limit))
	params.Set("offset", fmt.Sprintf("%d", (request.Page-1)*request.Limit))
	params.Set("text", request.Keyword)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return scraper.FetchResult{}, fmt.Errorf("build kalibrr request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.kalibrr.id")

	resp, err := a.Client.Do(req)
	if err != nil {
		return scraper.FetchResult{}, fmt.Errorf("execute kalibrr request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return scraper.FetchResult{}, fmt.Errorf("kalibrr returned status %d", resp.StatusCode)
	}

	var parsed struct {
		Count int `json:"count"`
		Jobs  []struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			CompanyName string `json:"company_name"`
			Description string `json:"description"`
			Slug        string `json:"slug"`
			CreatedAt   string `json:"created_at"`
			Company     struct {
				Code string `json:"code"`
			} `json:"company"`
			GoogleLocation struct {
				AddressComponents struct {
					City   string `json:"city"`
					Region string `json:"region"`
				} `json:"address_components"`
			} `json:"google_location"`
			MinimumSalary *int64 `json:"minimum_salary"`
			MaximumSalary *int64 `json:"maximum_salary"`
		} `json:"jobs"`
		Results []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			CompanyName string `json:"company_name"`
			Location    string `json:"location"`
			URL         string `json:"url"`
			PostedAt    string `json:"posted_at"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return scraper.FetchResult{}, fmt.Errorf("decode kalibrr response: %w", err)
	}

	items := make([]job.UpsertInput, 0, len(parsed.Jobs)+len(parsed.Results))
	for _, row := range parsed.Jobs {
		location := firstNonEmpty(row.GoogleLocation.AddressComponents.City, row.GoogleLocation.AddressComponents.Region)
		postedAt := parseOptionalTime(row.CreatedAt)
		items = append(items, job.UpsertInput{
			OriginalJobID: fmt.Sprintf("%d", row.ID),
			Title:         row.Name,
			Company:       row.CompanyName,
			Location:      location,
			Description:   row.Description,
			URL:           buildKalibrrURL(row.Company.Code, row.Slug, row.ID),
			SalaryMin:     row.MinimumSalary,
			SalaryMax:     row.MaximumSalary,
			PostedAt:      postedAt,
			RawData:       map[string]any{},
		})
	}

	for _, row := range parsed.Results {
		postedAt := parseOptionalTime(row.PostedAt)
		items = append(items, job.UpsertInput{
			OriginalJobID: row.ID,
			Title:         row.Name,
			Company:       row.CompanyName,
			Location:      row.Location,
			URL:           row.URL,
			PostedAt:      postedAt,
			RawData:       map[string]any{},
		})
	}

	hasMore := false
	if parsed.Count > 0 && request.Page*request.Limit < parsed.Count {
		hasMore = true
	}

	return scraper.FetchResult{
		Jobs:    items,
		HasMore: hasMore,
	}, nil
}

func buildKalibrrURL(companyCode, slug string, id int64) string {
	if strings.TrimSpace(companyCode) != "" && strings.TrimSpace(slug) != "" {
		return fmt.Sprintf("https://www.kalibrr.id/c/%s/jobs/%s", companyCode, slug)
	}
	return fmt.Sprintf("https://www.kalibrr.id/job/%d", id)
}
