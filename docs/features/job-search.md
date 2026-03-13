# Job Search Engine

## Tujuan

Memberikan pengalaman pencarian lowongan yang cepat, relevan, dan stabil untuk user free maupun premium.

## Prioritas

**MVP**

## Scope Fitur

### In Scope (MVP)

- Pencarian keyword pada judul, perusahaan, dan deskripsi.
- Filter `location`, `salary_min`, `source`.
- Pagination `page`, `limit`.
- Sorting by `created_at`/`posted_at`.
- Caching hasil query populer di Redis.

### Out of Scope (MVP)

- Semantic/vector search.
- Personalisasi ranking berbasis perilaku user.
- Filter multi-source dalam satu parameter.

## Aturan Bisnis

1. Endpoint pencarian bersifat publik.
2. Default sorting: `-posted_at`.
3. Batas `limit` maksimum 100 untuk melindungi performa.
4. Filter/query invalid -> `400 BAD_REQUEST`.
5. `page` di luar total halaman -> tetap `200` dengan data kosong.

## Perilaku Query

- Teknologi utama:
  - PostgreSQL Full Text Search (`tsvector`).
  - `pg_trgm` untuk fuzzy matching.
- Strategi cache:
  - Key: `jobs:search:{hash_query}`
  - TTL: 1 jam.
- Timeout query read disarankan: 1 detik.

## Kontrak API Terkait

- `GET /api/v1/jobs`
- `GET /api/v1/jobs/:id`

Lihat detail payload: [`../api/jobs.md`](../api/jobs.md).

## Dampak ke Database

- Tabel: `jobs`
- Index:
  - trigram pada `title`
  - trigram pada `company`
  - index timestamp untuk sorting terbaru

Lihat detail: [`../architecture/database.md`](../architecture/database.md).

## Edge Cases

- Query kosong: list latest jobs.
- Hasil kosong: tetap `200`, `data: []`.
- Redis unavailable: fallback langsung ke DB.

## NFR Minimum

| Metric | Target |
|---|---|
| `GET /jobs` p95 latency | < 300 ms |
| `GET /jobs/:id` p95 latency | < 200 ms |
| Jobs search error rate (`5xx`) | <= 0.1% |
| Cache hit rate query populer | >= 40% |

## Acceptance Criteria

- Query populer terlayani stabil dengan pagination konsisten.
- Hasil page tidak lompat karena sorting ambigu.
- Validasi parameter berjalan konsisten antar environment.
- Respon selalu mengikuti standar envelope API.
