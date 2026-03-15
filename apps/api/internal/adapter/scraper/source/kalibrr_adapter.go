package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/job"
)

const defaultKalibrrEndpoint = "https://www.kalibrr.id/kjs/job_board/search"

// KalibrrAdapter represents kalibrr adapter.
type KalibrrAdapter struct {
	Endpoint string
	Client   *http.Client
}

// NewKalibrrAdapter creates a new kalibrr adapter instance.
func NewKalibrrAdapter(client *http.Client) *KalibrrAdapter {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &KalibrrAdapter{
		Endpoint: defaultKalibrrEndpoint,
		Client:   client,
	}
}

// Source handles source.
func (a *KalibrrAdapter) Source() job.Source {
	return job.SourceKalibrr
}

// RequiresAuth handles requires auth.
func (a *KalibrrAdapter) RequiresAuth() bool {
	return false
}

// Fetch handles fetch.
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
		return scraper.FetchResult{}, scraper.WrapSourceError("build_request", 0, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.kalibrr.id")

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
			MinimumSalary nullableSalaryAmount `json:"minimum_salary"`
			MaximumSalary nullableSalaryAmount `json:"maximum_salary"`
			Salary        nullableSalaryAmount `json:"salary"`
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
		return scraper.FetchResult{}, scraper.WrapSourceError("decode_response", resp.StatusCode, err)
	}

	items := make([]job.UpsertInput, 0, len(parsed.Jobs)+len(parsed.Results))
	for _, row := range parsed.Jobs {
		location := firstNonEmpty(row.GoogleLocation.AddressComponents.City, row.GoogleLocation.AddressComponents.Region)
		postedAt := parseOptionalTime(row.CreatedAt)
		minimumSalary := row.MinimumSalary.Ptr()
		maximumSalary := row.MaximumSalary.Ptr()
		exactSalary := row.Salary.Ptr()

		if minimumSalary == nil && exactSalary != nil {
			minimumSalary = exactSalary
		}
		if maximumSalary == nil && exactSalary != nil {
			maximumSalary = exactSalary
		}

		salaryMin, salaryMax, salaryRange := normalizeSalaryFields(minimumSalary, maximumSalary, "")
		items = append(items, job.UpsertInput{
			OriginalJobID: fmt.Sprintf("%d", row.ID),
			Title:         row.Name,
			Company:       row.CompanyName,
			Location:      location,
			Description:   normalizeDescription(row.Description),
			URL:           buildKalibrrURL(row.Company.Code, row.Slug, row.ID),
			SalaryMin:     salaryMin,
			SalaryMax:     salaryMax,
			SalaryRange:   salaryRange,
			PostedAt:      postedAt,
			RawData: map[string]any{
				"search_item": row,
			},
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
			RawData: map[string]any{
				"search_item": row,
			},
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
