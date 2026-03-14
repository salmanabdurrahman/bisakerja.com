package source

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/app/scraper"
)

func TestKalibrrAdapter_Fetch_NormalizesSalaryAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		minimumSalaryJSON string
		maximumSalaryJSON string
		expectedSalaryMin int64
		expectedSalaryMax int64
	}{
		{
			name:              "integer salary values",
			minimumSalaryJSON: "15000000",
			maximumSalaryJSON: "25000000",
			expectedSalaryMin: 15000000,
			expectedSalaryMax: 25000000,
		},
		{
			name:              "decimal salary values",
			minimumSalaryJSON: "15000000.99",
			maximumSalaryJSON: "61754809.31151658",
			expectedSalaryMin: 15000000,
			expectedSalaryMax: 61754809,
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return jsonHTTPResponse(request, `{
					"count": 1,
					"jobs": [
						{
							"id": 101,
							"name": "Backend Engineer",
							"company_name": "Acme",
							"description": "Build APIs",
							"slug": "backend-engineer",
							"created_at": "2026-03-15T00:00:00Z",
							"company": {"code": "acme"},
							"google_location": {"address_components": {"city": "Jakarta", "region": "DKI Jakarta"}},
							"minimum_salary": `+testCase.minimumSalaryJSON+`,
							"maximum_salary": `+testCase.maximumSalaryJSON+`
						}
					]
				}`), nil
			})}

			adapter := NewKalibrrAdapter(client)
			adapter.Endpoint = "https://www.kalibrr.id/kjs/job_board/search"

			result, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
				Keyword: "backend",
				Page:    1,
				Limit:   20,
			})
			if err != nil {
				t.Fatalf("fetch: %v", err)
			}
			if len(result.Jobs) != 1 {
				t.Fatalf("expected 1 job, got %d", len(result.Jobs))
			}
			if result.Jobs[0].SalaryMin == nil || *result.Jobs[0].SalaryMin != testCase.expectedSalaryMin {
				t.Fatalf("expected salary_min %d, got %+v", testCase.expectedSalaryMin, result.Jobs[0].SalaryMin)
			}
			if result.Jobs[0].SalaryMax == nil || *result.Jobs[0].SalaryMax != testCase.expectedSalaryMax {
				t.Fatalf("expected salary_max %d, got %+v", testCase.expectedSalaryMax, result.Jobs[0].SalaryMax)
			}
		})
	}
}

func TestKalibrrAdapter_Fetch_AllowsEmptySalaryStrings(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return jsonHTTPResponse(request, `{
			"count": 1,
			"jobs": [
				{
					"id": 202,
					"name": "Backend Engineer",
					"company_name": "Acme",
					"description": "Build APIs",
					"slug": "backend-engineer",
					"created_at": "2026-03-15T00:00:00Z",
					"company": {"code": "acme"},
					"google_location": {"address_components": {"city": "Jakarta", "region": "DKI Jakarta"}},
					"minimum_salary": "",
					"maximum_salary": ""
				}
			]
		}`), nil
	})}

	adapter := NewKalibrrAdapter(client)
	result, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
		Keyword: "backend",
		Page:    1,
		Limit:   20,
	})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(result.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(result.Jobs))
	}
	if result.Jobs[0].SalaryMin != nil {
		t.Fatalf("expected salary_min nil, got %v", *result.Jobs[0].SalaryMin)
	}
	if result.Jobs[0].SalaryMax != nil {
		t.Fatalf("expected salary_max nil, got %v", *result.Jobs[0].SalaryMax)
	}
}

func TestKalibrrAdapter_Fetch_IncludesUpstreamBodySnippetOnBadRequest(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		response := jsonHTTPResponse(request, `{"message":"upstream blocked request"}`)
		response.StatusCode = http.StatusBadRequest
		return response, nil
	})}

	adapter := NewKalibrrAdapter(client)

	_, err := adapter.Fetch(context.Background(), scraper.FetchRequest{
		Keyword: "backend",
		Page:    1,
		Limit:   20,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "upstream blocked request") {
		t.Fatalf("expected upstream response snippet in error, got %v", err)
	}
}
