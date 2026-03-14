package scraper

import (
	"errors"
	"fmt"
	"strings"
)

// SourceError carries structured metadata for upstream source failures.
type SourceError struct {
	Operation  string
	StatusCode int
	Err        error
}

// Error returns the source error message.
func (e *SourceError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return "source error"
	}

	operation := strings.TrimSpace(e.Operation)
	if operation == "" {
		return e.Err.Error()
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s (status=%d): %v", operation, e.StatusCode, e.Err)
	}

	return fmt.Sprintf("%s: %v", operation, e.Err)
}

// Unwrap returns the wrapped error.
func (e *SourceError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// WrapSourceError wraps an upstream source error with structured metadata.
func WrapSourceError(operation string, statusCode int, err error) error {
	if err == nil {
		return nil
	}

	return &SourceError{
		Operation:  strings.TrimSpace(operation),
		StatusCode: statusCode,
		Err:        err,
	}
}

// SourceErrorDetails extracts structured metadata from a wrapped source error.
func SourceErrorDetails(err error) (operation string, statusCode int) {
	var sourceErr *SourceError
	if !errors.As(err, &sourceErr) || sourceErr == nil {
		return "", 0
	}

	return strings.TrimSpace(sourceErr.Operation), sourceErr.StatusCode
}
