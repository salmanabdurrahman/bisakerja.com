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

## Referensi Detail per Source

- Playbook scraping source: [`./source-scraping-playbook.md`](./source-scraping-playbook.md)
- Desain adapter + token boundary: [`../architecture/scraper-source-adapters.md`](../architecture/scraper-source-adapters.md)

## Sumber Data Awal & Status Integrasi

| Source | Tipe endpoint | Auth | Snapshot referensi | Catatan implementasi |
|---|---|---|---|---|
| `glints` | GraphQL `POST` | Tidak wajib bearer token | ✅ sukses (`item_count: 90`) | Wajib set header `x-glints-country-code`; gunakan pagination berbasis `page`. |
| `kalibrr` | REST `GET` | Tidak wajib bearer token | ✅ sukses (`item_count: 45`) | Gunakan pagination `limit+offset`; response utama berupa daftar jobs. |
| `jobstreet` | GraphQL `POST` | **Wajib Bearer token** | ⚠️ gagal tanpa token (`item_count: 0`) | Token bersifat short-lived; butuh token provider + rotasi aman. |
| `linkedin` | TBD | TBD | 📝 belum divalidasi pada referensi saat ini | Tetap non-blocking untuk MVP sampai ada kontrak teknis valid. |

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
2. Preflight capability per source (`requires_auth`, `pagination_mode`, `enabled`).
3. Resolve token jika source memerlukan auth (contoh: JobStreet).
4. Fetch halaman/endpoint source.
5. Parse item lowongan.
6. Normalize fields.
7. Check duplicate.
8. Insert jobs baru (`ON CONFLICT DO NOTHING`).
9. Publish event `JobCreated` hanya untuk row baru.

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
- **JobStreet token tidak valid/expired**: tandai source `auth_missing/auth_failed`, lakukan rotasi token, source lain tetap lanjut.

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
- Kegagalan auth pada satu source tidak menghentikan scraping source lain dalam batch yang sama.
