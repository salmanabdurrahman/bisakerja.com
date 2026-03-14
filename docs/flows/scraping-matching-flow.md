# Scraping & Matching Flow

## 1) Scraper Ingestion Flow (Source-Aware)

```mermaid
sequenceDiagram
  participant CRON as Scheduler
  participant S as Scraper Worker
  participant A as Source Adapter
  participant T as Token Provider
  participant P as Job Portal
  participant DB as PostgreSQL
  participant Q as Redis Queue

  CRON->>S: Trigger schedule
  S->>S: Load source capability matrix
  loop each source
    S->>A: Build request plan(source, keyword, page)
    alt source requires token (JobStreet)
      A->>T: Resolve bearer token
      alt token unavailable/expired
        T-->>A: auth_error
        A-->>S: mark source run failed_auth
        S->>DB: Save scrape run result (failed_auth)
        S->>S: skip source
      else token valid
        T-->>A: token
        A->>P: HTTP request with source headers + token
        P-->>A: JSON/GraphQL response
        A->>S: Parse + normalize jobs
        loop each job item
          S->>DB: INSERT ... ON CONFLICT DO NOTHING
          alt inserted (new row)
            S->>Q: Publish JobCreated
          else duplicate
            S->>S: skip publish
          end
        end
      end
    else source without token (Glints/Kalibrr)
      A->>P: HTTP request with source headers
      P-->>A: JSON/GraphQL response
      A->>S: Parse + normalize jobs
      loop each job item
        S->>DB: INSERT ... ON CONFLICT DO NOTHING
        alt inserted (new row)
          S->>Q: Publish JobCreated
        else duplicate
          S->>S: skip publish
        end
      end
    end
    S->>DB: Persist per-source scrape metrics
  end
```

## 2) Matcher + Notification Flow

```mermaid
sequenceDiagram
  participant Q as Redis Queue
  participant M as Matcher Worker
  participant DB as PostgreSQL
  participant N as Notifier
  participant SMTP as SMTP Provider

  Q->>M: Consume JobCreated
  M->>DB: Get job detail
  M->>DB: Get users with subscription_state=premium_active + preferences
  loop each user
    M->>M: Evaluate matching rules
    alt match
      M->>DB: Insert notifications(status=pending)
      M->>N: enqueue/send
      N->>SMTP: send email
      alt success
        N->>DB: Update notifications(status=sent)
      else failure
        N->>DB: Update notifications(status=failed)
      end
    end
  end
```

## 3) Preflight Checklist per Batch Scrape

1. Validasi source aktif (`glints`, `kalibrr`, `jobstreet`) dan keyword batch.
2. Cek `requires_auth` per source.
3. Untuk source auth-required, validasi token tersedia dan belum expired.
4. Set rate-limit budget per source (request concurrency + delay).
5. Inisialisasi run ID untuk observability dan audit.

## 4) Failure Path

| Kondisi | Mitigasi |
|---|---|
| Portal timeout/error | retry terbatas per source, source lain tetap lanjut |
| Missing/expired bearer token (JobStreet) | tandai `failed_auth`, lanjut source lain, trigger token refresh/runbook |
| `401/403` dari source auth-required | invalidasi token cache, rotate token, retry terbatas |
| Duplicate job dari source sama | ditahan oleh `UNIQUE(source, original_job_id)` |
| SMTP down | set `failed`, retry policy worker |
| Redis queue down sementara | fallback ke retry queue / alert operasional |
