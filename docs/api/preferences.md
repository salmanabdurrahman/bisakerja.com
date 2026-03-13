# User Preferences API

## 1) Get My Preferences

- **Method**: `GET`
- **Path**: `/preferences`
- **Auth**: Bearer Token (User login)
- **Ownership**: hanya boleh membaca preference user dari `JWT.sub`.

Endpoint ini dipakai untuk bootstrap halaman preferences di frontend dan read model untuk fitur yang membutuhkan preferensi user.

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Preferences retrieved",
    "request_id": "req_01J..."
  },
  "data": {
    "user_id": "2f2d5b4f-4ad8-4379-88df-d03790d1e9df",
    "keywords": ["golang", "backend", "software engineer"],
    "locations": ["jakarta", "remote"],
    "job_types": ["fulltime", "contract"],
    "salary_min": 10000000,
    "updated_at": "2026-03-13T18:00:00Z"
  }
}
```

### Catatan Response untuk User Baru

- Jika user belum pernah menyimpan preferensi, endpoint tetap mengembalikan `200 OK` dengan default:
  - `keywords: []`
  - `locations: []`
  - `job_types: []`
  - `salary_min: 0`
  - `updated_at: null`

### Error

- `401 UNAUTHORIZED` jika token tidak valid.

## 2) Update Preferences

- **Method**: `PUT`
- **Path**: `/preferences`
- **Auth**: Bearer Token (User login)
- **Ownership**: hanya boleh mengubah preference user dari `JWT.sub`.

Endpoint ini menyimpan preferensi user untuk mesin notifikasi.

### Request Body

```json
{
  "keywords": ["golang", "backend", "software engineer"],
  "locations": ["jakarta", "remote"],
  "job_types": ["fulltime", "contract"],
  "salary_min": 10000000
}
```

### Validation

| Field | Rules |
|---|---|
| `keywords` | wajib, array string, min 1 item, max 10 item, panjang tiap item 2-50 karakter. |
| `locations` | opsional, array string, max 5 item, panjang tiap item 2-100 karakter. |
| `job_types` | opsional, array string enum: `fulltime`, `parttime`, `contract`, `internship`, max 4 item. |
| `salary_min` | opsional, integer `0..999000000` (default 0 jika kosong). |

### Normalization

- String list di-trim whitespace.
- Nilai duplikat pada array dihapus (case-insensitive) sebelum disimpan.
- Field yang tidak dikirim tetap menggunakan nilai sebelumnya kecuali dikirim eksplisit sebagai array kosong.

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Preferences updated",
    "request_id": "req_01J..."
  },
  "data": {
    "user_id": "2f2d5b4f-4ad8-4379-88df-d03790d1e9df",
    "keywords": ["golang", "backend", "software engineer"],
    "locations": ["jakarta", "remote"],
    "job_types": ["fulltime", "contract"],
    "salary_min": 10000000,
    "updated_at": "2026-03-13T18:00:00Z"
  }
}
```

### Error

- `400 BAD_REQUEST` (`INVALID_JOB_TYPE`, `BAD_REQUEST`) jika payload tidak valid.
- `401 UNAUTHORIZED` jika token tidak valid.
- `403 FORBIDDEN` jika terjadi pelanggaran ownership (mis-konfigurasi middleware).

## Catatan

- Endpoint ini boleh diakses user free/premium, tetapi notifikasi hanya dipakai untuk user `premium_active`.
- Frontend sebaiknya bootstrap form dari `GET /preferences`, lalu kirim perubahan ke `PUT /preferences`.
