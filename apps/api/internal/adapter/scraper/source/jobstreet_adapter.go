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

// JobstreetAdapter represents jobstreet adapter.
type JobstreetAdapter struct {
	Endpoint       string
	Client         *http.Client
	SiteKey        string
	Locale         string
	Timezone       string
	Channel        string
	RequestSource  string
	Cookie         string
	AcceptLanguage string
	CustomFeatures string
	SeekSessionID  string
	SeekVisitorID  string
}

// NewJobstreetAdapter creates a new jobstreet adapter instance.
func NewJobstreetAdapter(client *http.Client) *JobstreetAdapter {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &JobstreetAdapter{
		Endpoint:       defaultJobstreetEndpoint,
		Client:         client,
		SiteKey:        "ID",
		Locale:         "id-ID",
		Timezone:       "Asia/Jakarta",
		Channel:        "mobileWeb",
		RequestSource:  "FE_HOME",
		AcceptLanguage: "en-US,en;q=0.9",
		CustomFeatures: "application/features.seek.all+json",
	}
}

// Source handles source.
func (a *JobstreetAdapter) Source() job.Source {
	return job.SourceJobstreet
}

// RequiresAuth handles requires auth.
func (a *JobstreetAdapter) RequiresAuth() bool {
	return true
}

// Fetch handles fetch.
func (a *JobstreetAdapter) Fetch(ctx context.Context, request scraper.FetchRequest) (scraper.FetchResult, error) {
	if strings.TrimSpace(request.Token) == "" {
		return scraper.FetchResult{}, fmt.Errorf("%w: missing bearer token", scraper.ErrTokenUnavailable)
	}

	endpoint := strings.TrimSpace(a.Endpoint)
	if endpoint == "" {
		endpoint = defaultJobstreetEndpoint
	}
	siteKey := firstNonEmpty(a.SiteKey, "ID")
	locale := firstNonEmpty(a.Locale, "id-ID")
	timezone := firstNonEmpty(a.Timezone, "Asia/Jakarta")
	sessionID := strings.TrimSpace(a.SeekSessionID)
	visitorID := strings.TrimSpace(a.SeekVisitorID)
	solID := extractCookieValue(a.Cookie, "sol_id")

	params := map[string]any{
		"channel":              firstNonEmpty(a.Channel, "mobileWeb"),
		"include":              []string{"seoData", "gptTargeting", "relatedSearches"},
		"keywords":             request.Keyword,
		"locale":               locale,
		"newSince":             time.Now().UTC().Unix(),
		"page":                 request.Page,
		"pageSize":             request.Limit,
		"queryHints":           []string{"spellingCorrection"},
		"relatedSearchesCount": 12,
		"siteKey":              siteKey,
		"source":               firstNonEmpty(a.RequestSource, "FE_HOME"),
	}
	if sessionID != "" {
		params["eventCaptureSessionId"] = sessionID
		params["userSessionId"] = sessionID
	}
	if visitorID != "" {
		params["eventCaptureUserId"] = visitorID
	}
	if solID != "" {
		params["solId"] = solID
	}

	payload := map[string]any{
		"operationName": "JobSearchV6",
		"variables": map[string]any{
			"params":   params,
			"locale":   locale,
			"timezone": timezone,
		},
		"query": jobstreetSearchQuery,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return scraper.FetchResult{}, scraper.WrapSourceError("marshal_request_payload", 0, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return scraper.FetchResult{}, scraper.WrapSourceError("build_request", 0, err)
	}
	req.Header.Set("Accept", defaultWildcardAccept)
	req.Header.Set("Accept-Language", firstNonEmpty(a.AcceptLanguage, "en-US,en;q=0.9"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(request.Token))
	req.Header.Set("Origin", "https://id.jobstreet.com")
	req.Header.Set("Referer", buildJobstreetReferer(request.Keyword, request.Page))
	req.Header.Set("User-Agent", defaultBrowserUserAgent)
	req.Header.Set("seek-request-brand", "jobstreet")
	req.Header.Set("seek-request-country", siteKey)
	req.Header.Set("x-custom-features", firstNonEmpty(a.CustomFeatures, "application/features.seek.all+json"))
	req.Header.Set("x-seek-site", "chalice")
	if cookie := strings.TrimSpace(a.Cookie); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if sessionID != "" {
		req.Header.Set("x-seek-ec-sessionid", sessionID)
	}
	if visitorID != "" {
		req.Header.Set("x-seek-ec-visitorid", visitorID)
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		return scraper.FetchResult{}, scraper.WrapSourceError("execute_request", 0, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return scraper.FetchResult{}, scraper.WrapSourceError("execute_request", resp.StatusCode, scraper.ErrSourceUnauthorized)
	}
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
		return scraper.FetchResult{}, scraper.WrapSourceError("decode_response", resp.StatusCode, err)
	}
	if len(parsed.Errors) > 0 {
		return scraper.FetchResult{}, scraper.WrapSourceError("graphql_response", resp.StatusCode, errors.New("upstream graphql returned errors"))
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

func buildJobstreetReferer(keyword string, page int) string {
	if page <= 0 {
		page = 1
	}

	return fmt.Sprintf(
		"https://id.jobstreet.com/id/%s-jobs?page=%d",
		slugifyPathSegment(keyword),
		page,
	)
}

const jobstreetSearchQuery = `query JobSearchV6($params: JobSearchV6QueryInput!, $locale: Locale!, $timezone: Timezone!) {
  jobSearchV6(params: $params) {
    totalCount
    canonicalCompany {
      description
    }
    data {
      advertiser {
        id
        description
      }
      branding {
        serpLogoUrl
      }
      bulletPoints
      classifications {
        classification {
          id
          description
        }
        subclassification {
          id
          description
        }
      }
      id
      title
      companyName
      teaser
      salaryLabel
      currencyLabel
      displayType
      employer {
        companyUrl
      }
      externalReferences {
        id
        sourceSystem
        type
      }
      isFeatured
      workTypes
      workArrangements { displayText }
      locations {
        countryCode
        label
        seoHierarchy {
          contextualName
        }
      }
      listingDate {
        dateTimeUtc
        label(context: JOB_POSTED, length: SHORT, timezone: $timezone, locale: $locale)
      }
      roleId
      solMetadata
      tags {
        label
        type
      }
      tracking
    }
    userQueryId
  }
}`
