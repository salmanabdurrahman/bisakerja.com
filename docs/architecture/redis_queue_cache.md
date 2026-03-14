# Redis, Queue, and Cache Architecture

## Peran Redis di Bisakerja

Redis dipakai untuk 2 kebutuhan utama saat ini:

1. **Cache layer** untuk mempercepat read endpoint.
2. **Rate limiting store** untuk API protection.

Catatan implementasi terkini:

- Queue state notification (`job events` dan `delivery tasks`) saat ini dipersistenkan di PostgreSQL (`notification_job_events`, `notification_delivery_tasks`) agar state tetap konsisten lintas proses worker.
- Queue berbasis Redis tetap opsional untuk scale-out lanjutan (post-MVP hardening).

## 1) Cache Design

### Search Cache

- Key: `jobs:search:{hash_query}`
- Value: JSON hasil list jobs.
- TTL: `3600` detik.
- Invalidation: TTL-based, optional manual clear saat terjadi perubahan besar.

### Job Detail Cache

- Key: `jobs:detail:{job_id}`
- Value: JSON detail job.
- TTL: `86400` detik.

## 2) Queue Design

### Notification Queue (Current: PostgreSQL)

- Storage: tabel `notification_delivery_tasks`.
- Producer: Matcher Engine.
- Consumer: Notification Worker.
- Payload contoh:

```json
{
  "notification_id": "uuid",
  "user_id": "uuid",
  "job_id": "uuid",
  "channel": "email"
}
```

### Job Event Queue (Current: PostgreSQL)

- Storage: tabel `notification_job_events`.
- Producer: Scraper Worker saat insert job baru.
- Consumer: Matcher Engine.

### Mayar Retry Queue

- Key: `queue:mayar:retry`
- Producer: Billing service ketika call Mayar gagal sementara (`429/5xx`).
- Consumer: Billing worker untuk retry dengan backoff.
- Payload contoh:

```json
{
  "request_type": "create_invoice",
  "user_id": "uuid",
  "attempt": 1,
  "idempotency_key": "billing:checkout:user_uuid:hash"
}
```

## 3) Retry & Dead-letter Strategy

### Retry

- Max retry: 3 kali.
- Backoff: eksponensial (`200ms`, `400ms`, `800ms`) + jitter.
- Timeout per attempt outbound: 5 detik.

### Dead-letter Queue (opsional post-MVP)

- Key: `queue:notifications:dlq`
- Menyimpan event yang gagal permanen untuk analisis manual.

## 4) Rate Limiting

### Inbound API

- Key: `ratelimit:{ip}:{minute}`
- Algoritma: fixed-window sederhana (MVP).
- Batas default: `100 req/minute`.

### Outbound Limit ke Mayar

- Key: `ratelimit:mayar:{minute}`
- Tujuan: menjaga call outbound tetap di bawah limit resmi Mayar (`20 req/minute/IP`).
- Praktik aman: target internal `<=18 req/minute`.

### Checkout Anti Double-Submit

- Key: `ratelimit:user:{user_id}:checkout:10s`
- Batas: 1 request checkout per 10 detik per user.

## 5) Observability Redis

Metrik yang disarankan:

- `redis_cache_hit_total`
- `redis_cache_miss_total`
- `queue_notifications_depth` (dari tabel queue PostgreSQL)
- `queue_notifications_failed_total`
- `queue_mayar_retry_depth`
- `mayar_rate_limit_block_total`

Target minimum:

- cache hit rate jobs search >= 40% (traffic normal).
- queue retry Mayar tidak menumpuk > 100 item lebih dari 5 menit.
