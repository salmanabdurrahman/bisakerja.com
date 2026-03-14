package source

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	defaultJSONAcceptHeader = "application/json, text/plain, */*"
	defaultWildcardAccept   = "*/*"
	defaultBrowserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
)

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
