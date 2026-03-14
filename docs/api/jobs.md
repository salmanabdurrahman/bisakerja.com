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

### Error

- `404 NOT_FOUND` jika job tidak ditemukan.

## Catatan Implementasi

- Search memanfaatkan PostgreSQL Full Text Search + `pg_trgm`.
- Endpoint list jobs direkomendasikan memakai Redis caching untuk query populer (`jobs:search:{hash_query}`) dengan TTL 1 jam.
- Query timeout rekomendasi: 1 detik untuk read path publik.
