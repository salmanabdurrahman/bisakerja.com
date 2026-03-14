package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName         string
	Environment     string
	HTTPPort        string
	ShutdownTimeout time.Duration
	WorkerTick      time.Duration
	ScraperPageSize int
	ScraperMaxPages int
}

func Load() Config {
	return Config{
		AppName:         getenv("APP_NAME", "bisakerja-api"),
		Environment:     getenv("APP_ENV", "development"),
		HTTPPort:        strings.TrimPrefix(getenv("HTTP_PORT", "8080"), ":"),
		ShutdownTimeout: parseDuration(getenv("SHUTDOWN_TIMEOUT", "10s"), 10*time.Second),
		WorkerTick:      parseDuration(getenv("WORKER_TICK_INTERVAL", "15s"), 15*time.Second),
		ScraperPageSize: parseInt(getenv("SCRAPER_PAGE_SIZE", "30"), 30),
		ScraperMaxPages: parseInt(getenv("SCRAPER_MAX_PAGES", "1"), 1),
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
