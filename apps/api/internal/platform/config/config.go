package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName                    string
	Environment                string
	HTTPPort                   string
	DatabaseURL                string
	DatabaseMaxOpenConns       int
	DatabaseMinOpenConns       int
	DatabaseMaxConnLifetime    time.Duration
	DatabaseMaxConnIdleTime    time.Duration
	DatabaseConnectTimeout     time.Duration
	ShutdownTimeout            time.Duration
	WorkerTick                 time.Duration
	ScraperPageSize            int
	ScraperMaxPages            int
	AuthJWTSecret              string
	AuthAccessTokenTTL         time.Duration
	AuthRefreshTokenTTL        time.Duration
	MayarBaseURL               string
	MayarAPIKey                string
	MayarRequestTimeout        time.Duration
	MayarMaxRetries            int
	BillingWebhookToken        string
	BillingRedirectAllowlist   []string
	BillingIdempotencyWindow   time.Duration
	BillingUserRateLimitWindow time.Duration
}

func Load() Config {
	return Config{
		AppName:                    getenv("APP_NAME", "bisakerja-api"),
		Environment:                getenv("APP_ENV", "development"),
		HTTPPort:                   strings.TrimPrefix(getenv("HTTP_PORT", "8080"), ":"),
		DatabaseURL:                getenv("DATABASE_URL", ""),
		DatabaseMaxOpenConns:       parseInt(getenv("DATABASE_MAX_OPEN_CONNS", "20"), 20),
		DatabaseMinOpenConns:       parseNonNegativeInt(getenv("DATABASE_MIN_OPEN_CONNS", "2"), 2),
		DatabaseMaxConnLifetime:    parseDuration(getenv("DATABASE_MAX_CONN_LIFETIME", "30m"), 30*time.Minute),
		DatabaseMaxConnIdleTime:    parseDuration(getenv("DATABASE_MAX_CONN_IDLE_TIME", "5m"), 5*time.Minute),
		DatabaseConnectTimeout:     parseDuration(getenv("DATABASE_CONNECT_TIMEOUT", "5s"), 5*time.Second),
		ShutdownTimeout:            parseDuration(getenv("SHUTDOWN_TIMEOUT", "10s"), 10*time.Second),
		WorkerTick:                 parseDuration(getenv("WORKER_TICK_INTERVAL", "15s"), 15*time.Second),
		ScraperPageSize:            parseInt(getenv("SCRAPER_PAGE_SIZE", "30"), 30),
		ScraperMaxPages:            parseInt(getenv("SCRAPER_MAX_PAGES", "1"), 1),
		AuthJWTSecret:              getenv("AUTH_JWT_SECRET", "bisakerja-dev-secret"),
		AuthAccessTokenTTL:         parseDuration(getenv("AUTH_ACCESS_TOKEN_TTL", "15m"), 15*time.Minute),
		AuthRefreshTokenTTL:        parseDuration(getenv("AUTH_REFRESH_TOKEN_TTL", "168h"), 168*time.Hour),
		MayarBaseURL:               getenv("MAYAR_BASE_URL", "https://api.mayar.id/hl/v1"),
		MayarAPIKey:                getenv("MAYAR_API_KEY", ""),
		MayarRequestTimeout:        parseDuration(getenv("MAYAR_REQUEST_TIMEOUT", "5s"), 5*time.Second),
		MayarMaxRetries:            parseInt(getenv("MAYAR_MAX_RETRIES", "3"), 3),
		BillingWebhookToken:        getenv("BILLING_WEBHOOK_TOKEN", "bisakerja-dev-webhook-token"),
		BillingRedirectAllowlist:   parseCSVList(getenv("BILLING_REDIRECT_ALLOWLIST", "app.bisakerja.com,localhost:3000")),
		BillingIdempotencyWindow:   parseDuration(getenv("BILLING_IDEMPOTENCY_WINDOW", "15m"), 15*time.Minute),
		BillingUserRateLimitWindow: parseDuration(getenv("BILLING_USER_RATE_LIMIT_WINDOW", "10s"), 10*time.Second),
	}
}

func (c Config) HTTPAddress() string {
	return ":" + c.HTTPPort
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}

	return fallback
}

func parseDuration(raw string, fallback time.Duration) time.Duration {
	if duration, err := time.ParseDuration(raw); err == nil {
		return duration
	}

	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return fallback
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseNonNegativeInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func parseCSVList(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		result = append(result, normalized)
	}
	return result
}
