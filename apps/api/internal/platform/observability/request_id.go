package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

type requestIDKey string

const requestIDContextKey requestIDKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = newRequestID()
		}

		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r.WithContext(WithRequestID(r.Context(), requestID)))
	})
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(requestIDContextKey).(string)
	if !ok {
		return ""
	}

	return value
}

func newRequestID() string {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err == nil {
		return hex.EncodeToString(buffer)
	}

	return time.Now().UTC().Format("20060102150405.000000000")
}
