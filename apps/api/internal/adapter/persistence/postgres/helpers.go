package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationCode = "23505"

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == uniqueViolationCode
}

func encodeJSON(value map[string]any) ([]byte, error) {
	if value == nil {
		return []byte("{}"), nil
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal json value: %w", err)
	}
	return encoded, nil
}

func decodeJSON(value []byte) (map[string]any, error) {
	if len(value) == 0 {
		return map[string]any{}, nil
	}

	result := make(map[string]any)
	if err := json.Unmarshal(value, &result); err != nil {
		return nil, fmt.Errorf("unmarshal json value: %w", err)
	}
	return result, nil
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

// nonNilStrings returns a non-nil copy of s.
// pgx serialises a nil []string as SQL NULL, which violates NOT NULL constraints
// on array columns. Use this helper before passing string slices to queries.
func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return append([]string(nil), s...)
}
