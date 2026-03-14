.PHONY: fmt vet test build lint run-api run-scraper run-notifier run-billing-worker check-migrations migrate-up migrate-down

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

build:
	go build ./cmd/api ./cmd/scraper ./cmd/notifier ./cmd/billing-worker ./cmd/migrate

lint: vet

run-api:
	go run ./cmd/api

run-scraper:
	go run ./cmd/scraper

run-notifier:
	go run ./cmd/notifier

run-billing-worker:
	go run ./cmd/billing-worker

check-migrations:
	go run ./cmd/migrate -mode validate

migrate-up:
	go run ./cmd/migrate -mode up -direction up

migrate-down:
	go run ./cmd/migrate -mode down -direction down
