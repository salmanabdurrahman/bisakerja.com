# Scraping & Matching Flow

## 1) Scraper Ingestion Flow

```mermaid
sequenceDiagram
  participant CRON as Scheduler
  participant S as Scraper Worker
  participant P as Job Portal
  participant DB as PostgreSQL
  participant Q as Redis Queue

  CRON->>S: Trigger schedule
  loop each source
    S->>P: HTTP request jobs page/API
    P-->>S: HTML/JSON response
    S->>S: Parse + normalize data
    loop each job item
      S->>DB: INSERT ... ON CONFLICT DO NOTHING
      alt inserted (new row)
        S->>Q: Publish JobCreated
      else duplicate
        S->>S: skip publish
      end
    end
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

## Failure Path

| Kondisi | Mitigasi |
|---|---|
| Portal timeout/error | retry terbatas per source, source lain tetap lanjut |
| Duplicate job dari source sama | ditahan oleh `UNIQUE(source, original_job_id)` |
| SMTP down | set `failed`, retry policy worker |
| Redis queue down sementara | fallback ke retry queue / alert operasional |
