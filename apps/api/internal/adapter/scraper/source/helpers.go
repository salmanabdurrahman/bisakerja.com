package source

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultJSONAcceptHeader = "application/json, text/plain, */*"
	defaultWildcardAccept   = "*/*"
	defaultBrowserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
)

var salaryComponentPattern = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)*)(?:\s*(jt|juta|rb|ribu|k|m)\b)?`)

type nullableSalaryAmount struct {
	value *int64
}

func (a *nullableSalaryAmount) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		a.value = nil
		return nil
	}

	if amount, err := parseSalaryAmount(trimmed); err == nil {
		a.value = amount
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode salary amount: %w", err)
	}

	amount, err := parseSalaryAmount(strings.TrimSpace(raw))
	if err != nil {
		return err
	}

	a.value = amount
	return nil
}

func (a nullableSalaryAmount) Ptr() *int64 {
	if a.value == nil {
		return nil
	}

	value := *a.value
	return &value
}

func parseSalaryAmount(raw string) (*int64, error) {
	normalizedRaw := strings.TrimSpace(raw)
	if normalizedRaw == "" {
		return nil, nil
	}

	value, err := strconv.ParseFloat(normalizedRaw, 64)
	if err != nil {
		return nil, fmt.Errorf("parse salary amount %q: %w", raw, err)
	}

	normalized := int64(math.Floor(value))
	return &normalized, nil
}

func parseOptionalTime(raw string) *time.Time {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}

	utc := parsed.UTC()
	return &utc
}

func normalizeDescription(raw string) string {
	normalized := strings.ReplaceAll(raw, "\u00a0", " ")
	normalized = strings.ReplaceAll(normalized, "&nbsp;", " ")
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.TrimSpace(normalized)
}

func normalizeSalaryFields(minAmount, maxAmount *int64, salaryLabel string) (*int64, *int64, string) {
	normalizedMin := cloneInt64Ptr(minAmount)
	normalizedMax := cloneInt64Ptr(maxAmount)
	normalizedLabel := normalizeWhitespace(salaryLabel)

	parsedMin, parsedMax := parseSalaryBounds(normalizedLabel)
	if normalizedMin == nil && parsedMin != nil {
		normalizedMin = parsedMin
	}
	if normalizedMax == nil && parsedMax != nil {
		normalizedMax = parsedMax
	}
	if normalizedMin != nil && normalizedMax != nil && *normalizedMin > *normalizedMax {
		normalizedMin, normalizedMax = normalizedMax, normalizedMin
	}

	normalizedRange := normalizedLabel
	if normalizedRange == "" {
		normalizedRange = formatSalaryRange(normalizedMin, normalizedMax)
	}

	return normalizedMin, normalizedMax, normalizedRange
}

func parseSalaryBounds(rawLabel string) (*int64, *int64) {
	label := strings.TrimSpace(strings.ToLower(rawLabel))
	if label == "" {
		return nil, nil
	}

	tokens := extractSalaryTokens(label)
	if len(tokens) == 0 {
		return nil, nil
	}

	if len(tokens) >= 2 {
		minimum := tokens[0]
		maximum := tokens[1]
		if minimum > maximum {
			minimum, maximum = maximum, minimum
		}
		return &minimum, &maximum
	}

	value := tokens[0]
	switch {
	case containsAny(label, "up to", "maximum", "max", "maks", "hingga", "sampai", "<="):
		return nil, &value
	case containsAny(label, "minimum", "min", "mulai dari", "from", "at least", ">="):
		return &value, nil
	default:
		return &value, &value
	}
}

func extractSalaryTokens(rawLabel string) []int64 {
	matches := salaryComponentPattern.FindAllStringSubmatch(rawLabel, -1)
	if len(matches) == 0 {
		return nil
	}
	implicitMultiplier := inferImplicitSalaryMultiplier(rawLabel)

	values := make([]int64, 0, len(matches))
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		numberPart := strings.TrimSpace(match[1])
		unitPart := strings.TrimSpace(strings.ToLower(match[2]))
		digitsOnly := strings.Map(func(char rune) rune {
			if char >= '0' && char <= '9' {
				return char
			}
			return -1
		}, numberPart)
		if unitPart == "" && len(digitsOnly) < 3 && implicitMultiplier == 1 {
			continue
		}

		appliedImplicitMultiplier := float64(1)
		if unitPart == "" && len(digitsOnly) < 3 {
			appliedImplicitMultiplier = implicitMultiplier
		}

		value, ok := parseSalaryToken(numberPart, unitPart, appliedImplicitMultiplier)
		if !ok {
			continue
		}
		values = append(values, value)
	}

	return values
}

func parseSalaryToken(numberPart, unitPart string, implicitMultiplier float64) (int64, bool) {
	numeric, err := parseFlexibleNumber(numberPart)
	if err != nil {
		return 0, false
	}

	multiplier := salaryUnitMultiplier(unitPart)
	if multiplier == 1 && implicitMultiplier > 1 {
		multiplier = implicitMultiplier
	}

	normalized := numeric * multiplier
	if normalized < 0 {
		return 0, false
	}

	return int64(math.Floor(normalized)), true
}

func parseFlexibleNumber(raw string) (float64, error) {
	normalized := strings.TrimSpace(raw)
	normalized = strings.ReplaceAll(normalized, "\u00a0", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	if normalized == "" {
		return 0, fmt.Errorf("parse salary amount %q: empty", raw)
	}

	dotCount := strings.Count(normalized, ".")
	commaCount := strings.Count(normalized, ",")

	switch {
	case dotCount > 0 && commaCount > 0:
		if strings.LastIndex(normalized, ".") > strings.LastIndex(normalized, ",") {
			normalized = strings.ReplaceAll(normalized, ",", "")
		} else {
			normalized = strings.ReplaceAll(normalized, ".", "")
			normalized = strings.ReplaceAll(normalized, ",", ".")
		}
	case commaCount > 0:
		if commaCount == 1 && digitsAfterSeparator(normalized, ",") <= 2 {
			normalized = strings.ReplaceAll(normalized, ",", ".")
		} else {
			normalized = strings.ReplaceAll(normalized, ",", "")
		}
	case dotCount > 0:
		if dotCount > 1 || digitsAfterSeparator(normalized, ".") > 2 {
			normalized = strings.ReplaceAll(normalized, ".", "")
		}
	}

	value, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, fmt.Errorf("parse salary amount %q: %w", raw, err)
	}
	return value, nil
}

func digitsAfterSeparator(value, separator string) int {
	if strings.Count(value, separator) != 1 {
		return 0
	}
	parts := strings.SplitN(value, separator, 2)
	if len(parts) != 2 {
		return 0
	}
	return len(parts[1])
}

func salaryUnitMultiplier(rawUnit string) float64 {
	switch strings.TrimSpace(strings.ToLower(rawUnit)) {
	case "jt", "juta", "m":
		return 1_000_000
	case "rb", "ribu", "k":
		return 1_000
	default:
		return 1
	}
}

func inferImplicitSalaryMultiplier(rawLabel string) float64 {
	label := strings.ToLower(strings.TrimSpace(rawLabel))
	if label == "" {
		return 1
	}

	hasCurrency := containsAny(label, "rp", "idr")
	hasMonthlyInterval := containsAny(label, "per month", "/month", "per bulan", "/bulan")
	if hasCurrency && hasMonthlyInterval {
		return 1_000_000
	}

	return 1
}

func formatSalaryRange(minAmount, maxAmount *int64) string {
	if minAmount == nil && maxAmount == nil {
		return ""
	}
	if minAmount != nil && maxAmount != nil {
		if *minAmount == *maxAmount {
			return strconv.FormatInt(*minAmount, 10)
		}
		return strconv.FormatInt(*minAmount, 10) + " - " + strconv.FormatInt(*maxAmount, 10)
	}
	if minAmount != nil {
		return ">= " + strconv.FormatInt(*minAmount, 10)
	}
	return "<= " + strconv.FormatInt(*maxAmount, 10)
}

func cloneInt64Ptr(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func containsAny(value string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(value, keyword) {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func slugifyPathSegment(value string) string {
	cleaned := strings.Builder{}
	lastDash := false
	for _, char := range strings.TrimSpace(strings.ToLower(value)) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			cleaned.WriteRune(char)
			lastDash = false
			continue
		}
		if lastDash {
			continue
		}
		cleaned.WriteRune('-')
		lastDash = true
	}

	result := strings.Trim(cleaned.String(), "-")
	if result == "" {
		return "jobs"
	}

	return result
}

func readBodySnippet(reader io.Reader, limit int64) string {
	if reader == nil || limit <= 0 {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(reader, limit))
	if err != nil {
		return ""
	}

	return strings.Join(strings.Fields(strings.TrimSpace(string(body))), " ")
}

func extractCookieValue(rawCookie, key string) string {
	targetKey := strings.TrimSpace(key)
	if targetKey == "" {
		return ""
	}

	for _, segment := range strings.Split(rawCookie, ";") {
		name, value, ok := strings.Cut(strings.TrimSpace(segment), "=")
		if !ok {
			continue
		}
		if strings.TrimSpace(name) == targetKey {
			return strings.TrimSpace(value)
		}
	}

	return ""
}
