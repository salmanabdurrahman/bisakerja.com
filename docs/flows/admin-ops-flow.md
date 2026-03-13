# Admin Operations Flow

## 1) Manual Scraper Trigger

```mermaid
sequenceDiagram
  participant A as Admin User
  participant API as API Service
  participant R as Redis Queue
  participant S as Scraper Worker

  A->>API: POST /admin/scraper/trigger
  API->>API: Verify token + role=admin
  API->>R: enqueue scrape task (dedupe window 60s)
  API-->>A: 202 Accepted + job_id
  R->>S: consume task
  S->>S: run scraping job
```

## 2) Dashboard Stats

```mermaid
sequenceDiagram
  participant A as Admin User
  participant API as API Service
  participant DB as PostgreSQL

  A->>API: GET /admin/stats
  API->>API: Verify token + role=admin
  API->>DB: aggregate stats queries
  DB-->>API: totals + daily metrics
  API-->>A: 200 OK stats payload
```

## Failure Path

| Kondisi | Respons |
|---|---|
| Non-admin request | `403 FORBIDDEN` |
| Invalid token | `401 UNAUTHORIZED` |
| Queue unavailable saat trigger | `503 SERVICE_UNAVAILABLE` |
