# Jobs API

## 1) List/Search Jobs

- **Method**: `GET`
- **Path**: `/jobs`
- **Auth**: Public

### Query Parameters

| Name | Type | Required | Default | Validation |
|---|---|---|---|---|
| `q` | string | no | - | max 200 chars; trim whitespace. |
| `location` | string | no | - | max 100 chars. |
| `salary_min` | integer | no | - | `>= 0`. |
| `page` | integer | no | 1 | `>= 1`. |
| `limit` | integer | no | 20 | `1..100`. |
| `sort` | string | no | `-posted_at` | enum: `-posted_at`, `posted_at`, `-created_at`, `created_at`. |
| `source` | string | no | - | enum Phase 1.1: `glints`, `jobstreet`, `kalibrr` (`linkedin` reserved untuk tahap lanjut). |

### Behavior Notes

- Query kosong (`q` tidak ada) = list lowongan terbaru.
- `page` di atas `total_pages` -> `200 OK` dengan `data: []`.
- `source` multi-value (mis. `source=glints,kalibrr`) belum didukung di MVP.
- `salary_range` selalu diprioritaskan dari label source; jika label kosong backend akan membentuk fallback dari angka numerik:
  - exact salary (`salary_min == salary_max`) -> `"10000000"`
  - min-only -> `">= 10000000"`
  - max-only -> `"<= 12000000"`
  - range -> `"10000000 - 15000000"`
- Untuk label shorthand bulanan seperti `Rp 8 – Rp 12 per month`, parser backend menurunkan `salary_min=8000000` dan `salary_max=12000000` agar filter salary tetap konsisten.

### Example Request

`GET /api/v1/jobs?q=golang&location=jakarta&page=1&limit=20&sort=-posted_at&source=glints`

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Jobs retrieved",
    "request_id": "req_01J...",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total_pages": 3,
      "total_records": 56
    }
  },
  "data": [
    {
      "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "title": "Backend Engineer (Golang)",
      "company": "Contoh Teknologi",
      "location": "Jakarta",
      "salary_range": "12000000 - 22000000",
      "source": "glints",
      "posted_at": "2026-03-13T12:00:00Z"
    }
  ]
}
```

### Error

- `400 BAD_REQUEST` (`INVALID_LIMIT`, `INVALID_SORT`, `INVALID_SOURCE`) untuk query tidak valid.

## 2) Get Job Detail

- **Method**: `GET`
- **Path**: `/jobs/:id`
- **Auth**: Public

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Job detail retrieved",
    "request_id": "req_01J..."
  },
  "data": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "title": "Backend Engineer (Golang)",
    "company": "Contoh Teknologi",
    "location": "Jakarta",
    "description": "Membangun API dan worker dengan Go...",
    "salary_range": "12000000 - 22000000",
    "source": "glints",
    "url": "https://glints.com/id/opportunities/jobs/123456",
    "posted_at": "2026-03-13T12:00:00Z",
    "created_at": "2026-03-13T12:03:04Z"
  }
}
```

### Detail Notes

- `data.description` dapat berisi rich text/HTML dari source (terutama dari Kalibrr/detail source lain).
- Consumer frontend wajib melakukan sanitasi HTML sebelum render ke DOM.

### Error

- `404 NOT_FOUND` jika job tidak ditemukan.

## Catatan Implementasi

- Search memanfaatkan PostgreSQL LIKE multi-kata pada `title`, `company`, dan `description`.
- Multi-kata didukung: query `q` dipecah per spasi, setiap kata menjadi kondisi `AND` tersendiri (contoh: "backend intern" → `title LIKE '%backend%' AND title LIKE '%intern%'`).
- Indeks `idx_jobs_title_lower` dan `idx_jobs_description_lower` (operator `text_pattern_ops`) ditambahkan via migrasi `000006_jobs_search_index` untuk performa pencarian case-insensitive.
- Endpoint list jobs direkomendasikan memakai Redis caching untuk query populer (`jobs:search:{hash_query}`) dengan TTL 1 jam.
- Query timeout rekomendasi: 1 detik untuk read path publik.

## 3) Search Job Titles (Autocomplete)

- **Method**: `GET`
- **Path**: `/jobs/titles`
- **Auth**: Public

Endpoint publik untuk autocomplete judul lowongan. Dipakai UI AI Tools saat user mengetik title untuk job-fit summary dan cover letter draft.

### Query Parameters

| Name | Type | Required | Default | Validation |
|---|---|---|---|---|
| `q` | string | yes | - | min 1 karakter, max 200 karakter. |

### Behavior Notes

- Prefix match case-insensitive pada kolom `title` (`LOWER(title) LIKE $1%`).
- Distinct titles, diurutkan ascending, maksimum 10 hasil.
- Jika `q` kosong atau tidak ada, endpoint mengembalikan `400 BAD_REQUEST`.

### Example Request

`GET /api/v1/jobs/titles?q=backend`

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Job titles retrieved",
    "request_id": "req_01J..."
  },
  "data": {
    "titles": [
      "Backend Engineer",
      "Backend Engineer (Golang)",
      "Backend Intern",
      "Backend Software Engineer"
    ]
  }
}
```

### Error

- `400 BAD_REQUEST` (`BAD_REQUEST`) jika `q` kosong atau tidak ada.
