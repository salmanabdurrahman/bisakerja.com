# Discovery Flow (Frontend)

## Tujuan

Menggambarkan alur user saat menemukan lowongan: dari submit search, melihat hasil, membuka detail, lalu lanjut apply ke URL sumber.

## Happy Path

```mermaid
sequenceDiagram
  participant U as User
  participant FE as Frontend (Web)
  participant API as Bisakerja API

  U->>FE: Buka halaman pencarian lowongan
  FE-->>U: Render state search_idle

  U->>FE: Submit keyword/filter
  FE->>API: GET /api/v1/jobs?q=golang&location=jakarta&page=1&limit=20
  API-->>FE: 200 Jobs retrieved (data + pagination)
  FE-->>U: Render state search_results

  U->>FE: Klik salah satu kartu lowongan
  FE->>API: GET /api/v1/jobs/:id
  API-->>FE: 200 Job detail retrieved
  FE->>FE: Sanitasi `data.description` jika rich HTML
  FE->>FE: Normalisasi display `salary_range` (comparator/range shorthand)
  FE-->>U: Render detail_ready + tombol "Lamar"

  U->>FE: Klik "Lamar"
  FE-->>U: Redirect ke job source URL
```

## Failure Path

```mermaid
sequenceDiagram
  participant U as User
  participant FE as Frontend (Web)
  participant API as Bisakerja API

  U->>FE: Submit filter tidak valid
  FE->>API: GET /api/v1/jobs?... (invalid params)
  API-->>FE: 400 BAD_REQUEST
  FE-->>U: Render search_error_validation + hint perbaikan filter

  U->>FE: Buka detail lowongan yang sudah tidak tersedia
  FE->>API: GET /api/v1/jobs/:id
  API-->>FE: 404 NOT_FOUND
  FE-->>U: Render detail_not_found + CTA kembali ke hasil pencarian

  U->>FE: Search saat jaringan/API bermasalah
  FE->>API: GET /api/v1/jobs
  API-->>FE: 5xx / timeout
  FE-->>U: Render search_error_retry + tombol "Coba lagi"

  U->>FE: Search terlalu sering
  FE->>API: GET /api/v1/jobs
  API-->>FE: 429 TOO_MANY_REQUESTS
  FE-->>U: Render search_error_rate_limited + info jeda retry
```

## UI State Transitions

| Current State | Trigger | Next State | Catatan UI |
|---|---|---|---|
| `search_idle` | User submit query/filter | `search_loading` | Disable submit sementara request berjalan |
| `search_loading` | `GET /api/v1/jobs` sukses + `data.length > 0` | `search_results` | Tampilkan list + pagination |
| `search_loading` | `GET /api/v1/jobs` sukses + `data.length = 0` | `search_empty` | Empty state dengan saran ubah keyword/filter |
| `search_loading` | `GET /api/v1/jobs` gagal (`400`) | `search_error_validation` | Tampilkan error validasi query |
| `search_loading` | `GET /api/v1/jobs` gagal (`429`) | `search_error_rate_limited` | Tampilkan cooldown + retry terkontrol |
| `search_loading` | `GET /api/v1/jobs` gagal (`5xx`/timeout) | `search_error_retry` | Tampilkan retry action |
| `search_results` | User klik job card | `detail_loading` | Tampilkan skeleton/overlay detail |
| `detail_loading` | `GET /api/v1/jobs/:id` sukses | `detail_ready` | Render detail lengkap + tombol apply |
| `detail_loading` | `GET /api/v1/jobs/:id` gagal (`404`) | `detail_not_found` | Tampilkan pesan lowongan tidak tersedia |

## Backend/API Touchpoints

- `GET /api/v1/jobs` — list/search jobs ([Jobs API](../../api/jobs.md)).
- `GET /api/v1/jobs/:id` — detail lowongan ([Jobs API](../../api/jobs.md)).
- Caching behavior untuk endpoint jobs mengikuti backend search-serving flow ([Search Serving Flow](../../flows/search-serving-flow.md)).

## Acceptance Criteria Flow

- URL filter (`q`, `location`, `salary_min`, `sort`, `page`, `limit`, `source`) selalu sinkron dengan request API list jobs.
- Pergantian filter utama me-reset `page=1` sebelum request dikirim.
- `404` detail selalu menghasilkan state `detail_not_found` dengan CTA kembali ke hasil.
- `429` dan `5xx` menampilkan state retry yang berbeda agar user tahu tindakan lanjutan.
- Konten `description` bertipe HTML tidak pernah dirender mentah; sanitasi frontend wajib aktif sebelum tampil.
- Salary fallback comparator seperti `<= 2999998`/`<= Rp 3.500.000` ditampilkan sebagai label friendly (`Up to Rp ...`) dan shorthand `Rp 8 – Rp 12 per month` tetap terformat readable.
