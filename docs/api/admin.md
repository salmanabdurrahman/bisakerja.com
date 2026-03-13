# Admin API

Semua endpoint di dokumen ini memerlukan:

- Token Bearer valid.
- Role user = `admin`.

## 1) Trigger Scraper

- **Method**: `POST`
- **Path**: `/admin/scraper/trigger`
- **Auth**: Admin

Digunakan untuk menjalankan scraping manual.

### Request Body

```json
{
  "source": "glints"
}
```

- `source` opsional.
- Jika kosong, sistem men-trigger semua source MVP (`glints`, `jobstreet`, `kalibrr`, `linkedin`).

### Validation & Idempotency

- `source` harus ada dalam enum source yang didukung.
- Trigger identik dalam jendela 60 detik boleh dide-duplikasi oleh queue untuk mencegah job storm.

### Response `202 Accepted`

```json
{
  "meta": {
    "code": 202,
    "status": "success",
    "message": "Scraper job queued",
    "request_id": "req_01J..."
  },
  "data": {
    "job_id": "scrape-20260313-001",
    "source": "glints",
    "queued_at": "2026-03-13T18:10:00Z"
  }
}
```

### Error

- `400 BAD_REQUEST` jika `source` tidak dikenali.
- `401 UNAUTHORIZED` jika token invalid.
- `403 FORBIDDEN` jika role bukan admin.
- `503 SERVICE_UNAVAILABLE` jika queue backend tidak tersedia.

## 2) Get Dashboard Stats

- **Method**: `GET`
- **Path**: `/admin/stats`
- **Auth**: Admin

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Stats retrieved",
    "request_id": "req_01J..."
  },
  "data": {
    "total_users": 1500,
    "premium_users": 236,
    "total_jobs": 28500,
    "jobs_scraped_today": 480,
    "notifications_sent_today": 311,
    "notifications_failed_today": 9
  }
}
```

### Operational Notes

- Endpoint stats boleh memakai cache pendek (30-60 detik) untuk menekan query agregasi berat.
- Timestamp cut-off `*_today` mengikuti zona waktu UTC agar konsisten lintas service.

### Error

- `401 UNAUTHORIZED`
- `403 FORBIDDEN`
