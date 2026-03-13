# Admin Operations

## Tujuan

Menyediakan endpoint operasional untuk tim internal agar dapat mengontrol proses ingestion dan memantau kesehatan sistem.

## Prioritas

**MVP**

## Scope Fitur

### In Scope (MVP)

- Trigger scraping manual seluruh source atau source tertentu.
- Melihat statistik ringkas platform.
- Akses dibatasi role admin.

### Out of Scope (MVP)

- UI admin lengkap (role matrix detail, audit panel kompleks).
- Bulk moderation jobs/users dari dashboard.

## Aturan Bisnis

1. Endpoint admin wajib token valid + role `admin`.
2. Trigger scraper harus bersifat asynchronous (`202 Accepted`).
3. Request trigger dicatat untuk observability/audit operasional.
4. Trigger identik beruntun boleh dide-duplikasi untuk mencegah queue flooding.

## Endpoint Terkait

- `POST /api/v1/admin/scraper/trigger`
- `GET /api/v1/admin/stats`

Lihat detail: [`../api/admin.md`](../api/admin.md).

## Metrik Operasional Minimum

- total_users
- total_jobs
- jobs_scraped_today
- notifications_sent_today
- scraper_failures_today
- queue_depth_scraper

## NFR Minimum

| Metric | Target |
|---|---|
| Admin API p95 latency | < 500 ms |
| `POST /admin/scraper/trigger` success ratio | >= 99% |
| Stats freshness | <= 60 detik |

## Acceptance Criteria

- User non-admin tidak bisa akses endpoint admin (`403`).
- Admin bisa trigger scrape source tertentu dan menerima `job_id`.
- Statistik yang ditampilkan konsisten dengan data database (toleransi delay <= 60 detik).
