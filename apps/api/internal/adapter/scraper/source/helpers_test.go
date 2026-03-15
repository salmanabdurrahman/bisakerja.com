package source

import "testing"

func TestParseSalaryBounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		label       string
		expectedMin *int64
		expectedMax *int64
	}{
		{
			name:        "range with currency symbols",
			label:       "Rp 8.000.000 - Rp 12.000.000",
			expectedMin: int64Ptr(8_000_000),
			expectedMax: int64Ptr(12_000_000),
		},
		{
			name:        "minimum only label",
			label:       "Mulai dari Rp 10.000.000",
			expectedMin: int64Ptr(10_000_000),
			expectedMax: nil,
		},
		{
			name:        "maximum only label",
			label:       "Up to Rp 15,000,000",
			expectedMin: nil,
			expectedMax: int64Ptr(15_000_000),
		},
		{
			name:        "exact salary label",
			label:       "Rp 9.500.000 per month",
			expectedMin: int64Ptr(9_500_000),
			expectedMax: int64Ptr(9_500_000),
		},
		{
			name:        "short unit notation",
			label:       "Rp 12.5 jt - 15 jt",
			expectedMin: int64Ptr(12_500_000),
			expectedMax: int64Ptr(15_000_000),
		},
		{
			name:        "implicit monthly million range",
			label:       "Rp 8 – Rp 12 per month",
			expectedMin: int64Ptr(8_000_000),
			expectedMax: int64Ptr(12_000_000),
		},
		{
			name:        "comparator maximum fallback",
			label:       "<= 2999998",
			expectedMin: nil,
			expectedMax: int64Ptr(2_999_998),
		},
		{
			name:        "non numeric label",
			label:       "Competitive salary",
			expectedMin: nil,
			expectedMax: nil,
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			minimum, maximum := parseSalaryBounds(testCase.label)
			assertInt64PtrEqual(t, minimum, testCase.expectedMin, "salary_min")
			assertInt64PtrEqual(t, maximum, testCase.expectedMax, "salary_max")
		})
	}
}

func TestNormalizeSalaryFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		minimum             *int64
		maximum             *int64
		label               string
		expectedMin         *int64
		expectedMax         *int64
		expectedSalaryRange string
	}{
		{
			name:                "min max fallback range",
			minimum:             int64Ptr(9_000_000),
			maximum:             int64Ptr(15_000_000),
			label:               "",
			expectedMin:         int64Ptr(9_000_000),
			expectedMax:         int64Ptr(15_000_000),
			expectedSalaryRange: "9000000 - 15000000",
		},
		{
			name:                "exact salary fallback range",
			minimum:             int64Ptr(7_000_000),
			maximum:             int64Ptr(7_000_000),
			label:               "",
			expectedMin:         int64Ptr(7_000_000),
			expectedMax:         int64Ptr(7_000_000),
			expectedSalaryRange: "7000000",
		},
		{
			name:                "label fills missing minimum",
			minimum:             nil,
			maximum:             int64Ptr(12_000_000),
			label:               "Rp 8.000.000 - Rp 12.000.000",
			expectedMin:         int64Ptr(8_000_000),
			expectedMax:         int64Ptr(12_000_000),
			expectedSalaryRange: "Rp 8.000.000 - Rp 12.000.000",
		},
		{
			name:                "maximum only fallback range",
			minimum:             nil,
			maximum:             int64Ptr(12_000_000),
			label:               "",
			expectedMin:         nil,
			expectedMax:         int64Ptr(12_000_000),
			expectedSalaryRange: "<= 12000000",
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			minimum, maximum, salaryRange := normalizeSalaryFields(testCase.minimum, testCase.maximum, testCase.label)
			assertInt64PtrEqual(t, minimum, testCase.expectedMin, "salary_min")
			assertInt64PtrEqual(t, maximum, testCase.expectedMax, "salary_max")
			if salaryRange != testCase.expectedSalaryRange {
				t.Fatalf("expected salary_range %q, got %q", testCase.expectedSalaryRange, salaryRange)
			}
		})
	}
}

func TestNormalizeDescription(t *testing.T) {
	t.Parallel()

	raw := "\r\n <p>Hello&nbsp;World</p>\r\n\r\n"
	normalized := normalizeDescription(raw)

	if normalized != "<p>Hello World</p>" {
		t.Fatalf("expected normalized description %q, got %q", "<p>Hello World</p>", normalized)
	}
}

func assertInt64PtrEqual(t *testing.T, actual, expected *int64, field string) {
	t.Helper()

	switch {
	case actual == nil && expected == nil:
		return
	case actual == nil && expected != nil:
		t.Fatalf("expected %s %d, got nil", field, *expected)
	case actual != nil && expected == nil:
		t.Fatalf("expected %s nil, got %d", field, *actual)
	default:
		if *actual != *expected {
			t.Fatalf("expected %s %d, got %d", field, *expected, *actual)
		}
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}
