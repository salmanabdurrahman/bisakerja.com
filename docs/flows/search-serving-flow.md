# Search Serving Flow

## Search Request Flow (Public)

```mermaid
sequenceDiagram
  participant U as User
  participant API as API Service
  participant R as Redis
  participant DB as PostgreSQL

  U->>API: GET /jobs?q=golang&location=jakarta&page=1&limit=20
  API->>API: Validate query params (limit/sort/source/page)
  API->>R: GET jobs:search:{hash}
  alt cache hit
    R-->>API: cached result
  else cache miss
    API->>DB: execute search query
    DB-->>API: rows + count
    API->>R: SET jobs:search:{hash} (TTL 1h)
  end
  API-->>U: JSON response with pagination
```

## Job Detail Flow

```mermaid
sequenceDiagram
  participant U as User
  participant API as API Service
  participant R as Redis
  participant DB as PostgreSQL

  U->>API: GET /jobs/:id
  API->>R: GET jobs:detail:{id}
  alt cache hit
    R-->>API: cached detail
  else cache miss
    API->>DB: select job by id
    DB-->>API: job row
    API->>R: SET jobs:detail:{id} (TTL 24h)
  end
  API-->>U: JSON detail
```

## Failure Path

| Kondisi | Respons |
|---|---|
| Query invalid | `400 BAD_REQUEST` |
| Job not found | `404 NOT_FOUND` |
| Redis unavailable | fallback langsung ke DB (tanpa cache) |
| DB timeout | `503 SERVICE_UNAVAILABLE` + `request_id` untuk tracing |
