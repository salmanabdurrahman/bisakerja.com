# Job Aggregation Engine

## Tujuan

Mengumpulkan lowongan kerja dari beberapa portal ke satu sumber data terstandar agar dapat dicari cepat oleh user.

## Prioritas

**MVP**

## Scope Fitur

### In Scope (MVP)

- Scraping berbasis HTTP untuk portal target yang bisa diakses tanpa browser headless.
- Parsing data lowongan: `title`, `company`, `location`, `description`, `salary_range`, `url`, `posted_at`.
- Normalisasi format data (teks, salary, lokasi).
- Deduplikasi berbasis kombinasi `source + original_job_id`.
- Scheduling otomatis 2 kali sehari.

### Out of Scope (MVP)

- Scraping portal heavy-SPA yang memerlukan browser automation.
- OCR atau ekstraksi dari gambar/PDF.
- Auto-apply ke portal pihak ketiga.
- Cross-source deduplication semantik.

## Sumber Data Awal

- Glints
- Jobstreet
- Kalibrr
- LinkedIn

## Aturan Bisnis

1. Data lowongan baru hanya disimpan jika belum ada pasangan `source + original_job_id`.
2. Jika field non-kritis kosong (contoh salary), data tetap disimpan.
3. Jika portal gagal di-scrape, proses portal lain tetap lanjut.
4. Error scraping dicatat ke log + metrik error per source.
5. Re-scrape job existing tidak mengirim event `JobCreated` ulang.

## Kontrak Teknis

### Trigger Scheduler

- Cron: `0 2,14 * * *`

### Pipeline

1. Load daftar source.
2. Fetch halaman/endpoint source.
3. Parse item lowongan.
4. Normalize fields.
5. Check duplicate.
6. Insert jobs baru (`ON CONFLICT DO NOTHING`).
7. Publish event `JobCreated` hanya untuk row baru.

## Dampak ke Database

- Tabel utama: `jobs`
- Constraint: `UNIQUE(source, original_job_id)`
- Index penting:
  - `idx_jobs_title_trgm`
  - `idx_jobs_company_trgm`
  - `idx_jobs_posted_at`

Lihat detail: [`../architecture/database.md`](../architecture/database.md).

## Dependensi API

Fitur ini tidak diekspos langsung ke user endpoint publik, tetapi berdampak pada:

- `GET /api/v1/jobs`
- `GET /api/v1/jobs/:id`
- `POST /api/v1/admin/scraper/trigger`

Lihat detail: [`../api/jobs.md`](../api/jobs.md), [`../api/admin.md`](../api/admin.md).

## Risiko & Mitigasi

- **IP blocked / rate limit portal**: concurrency limit per source + retry terbatas.
- **Perubahan struktur HTML**: parser per source terpisah + observability error per source.
- **Data duplikat tinggi**: enforce constraint DB + cek conflict insert.

## NFR Minimum

| Metric | Target |
|---|---|
| Scrape success ratio per source | >= 95% / hari |
| Duplicate insert conflict ratio | terukur, tidak menyebabkan error 5xx |
| End-to-end ingest latency (fetch -> DB) | p95 < 10 menit per batch |

## Acceptance Criteria

- Scraper berjalan sesuai jadwal dan menghasilkan data baru.
- Tidak ada duplikasi untuk kombinasi source + original id.
- Field minimum (`title`, `company`, `url`, `source`) selalu terisi.
- Event notifikasi hanya dipublish untuk row baru.
