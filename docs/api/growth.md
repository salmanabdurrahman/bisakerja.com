# Growth API (Saved Searches, Watchlist, Notifications)

Dokumen ini merangkum endpoint backend Phase 3 untuk retensi user.

Catatan implementasi saat ini: endpoint growth menyediakan konfigurasi/rule user-scoped (saved search, watchlist, notification center). Pengiriman alert terjadwal batch dari saved search/watchlist dapat dihubungkan ke worker scheduler pada iterasi lanjutan tanpa perubahan kontrak HTTP.

Semua endpoint di dokumen ini:

- user-scoped dan wajib Bearer token valid,
- menggunakan ownership `JWT.sub`,
- runtime path berprefix `/api/v1/*`.

## 1) Saved Searches

### 1.1 Create Saved Search

- **Method**: `POST`
- **Path**: `/saved-searches`

#### Request Body

```json
{
  "query": "golang backend",
  "location": "jakarta",
  "source": "glints",
  "salary_min": 12000000,
  "frequency": "daily_digest",
  "is_active": true
}
```

#### Validation

| Field | Rules |
|---|---|
| `query` | wajib, 2..200 karakter |
| `location` | opsional, <= 100 karakter |
| `source` | opsional, enum `glints`, `kalibrr`, `jobstreet` |
| `salary_min` | opsional, integer `>= 0` |
| `frequency` | opsional, enum `instant`, `daily_digest`, `weekly_digest` (default `instant`) |
| `is_active` | opsional, boolean (default `true`) |

#### Response `201 Created`

```json
{
  "meta": {
    "code": 201,
    "status": "success",
    "message": "Saved search created",
    "request_id": "req_01J..."
  },
  "data": {
    "id": "ss_abc123",
    "query": "golang backend",
    "location": "jakarta",
    "source": "glints",
    "salary_min": 12000000,
    "frequency": "daily_digest",
    "is_active": true,
    "created_at": "2026-03-14T13:30:00Z",
    "updated_at": "2026-03-14T13:30:00Z"
  }
}
```

### 1.2 List Saved Searches

- **Method**: `GET`
- **Path**: `/saved-searches`

#### Response `200 OK`

`data` berupa array saved search milik user login, diurutkan terbaru lebih dulu.

### 1.3 Delete Saved Search

- **Method**: `DELETE`
- **Path**: `/saved-searches/:id`

#### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Saved search deleted",
    "request_id": "req_01J..."
  },
  "data": {
    "id": "ss_abc123"
  }
}
```

## 2) Company Watchlist

### 2.1 Add Company to Watchlist

- **Method**: `POST`
- **Path**: `/watchlist/companies`

#### Request Body

```json
{
  "company_slug": "acme-group"
}
```

#### Validation

- `company_slug` wajib, slug format (`a-z`, `0-9`, `-`) panjang 2..80 karakter.

### 2.2 List Watchlist Companies

- **Method**: `GET`
- **Path**: `/watchlist/companies`

#### Response `200 OK`

`data` berupa array company watchlist user login.

### 2.3 Remove Company from Watchlist

- **Method**: `DELETE`
- **Path**: `/watchlist/companies/:company_slug`

## 3) Notification Center

### 3.1 List Notifications

- **Method**: `GET`
- **Path**: `/notifications`

#### Query Parameters

| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| `page` | integer | no | 1 | halaman (`>=1`) |
| `limit` | integer | no | 20 | item per halaman (`1..100`) |
| `unread_only` | boolean | no | false | filter notifikasi belum dibaca |

#### Response `200 OK`

`data[]` berisi notifikasi user (`id`, `job_id`, `channel`, `status`, `error_message`, `sent_at`, `read_at`, `created_at`) dengan pagination standar.

### 3.2 Mark Notification as Read

- **Method**: `PATCH`
- **Path**: `/notifications/:id/read`

#### Response `200 OK`

Mengembalikan notifikasi yang sudah memiliki `read_at`.

## Error Ringkas

- `400 BAD_REQUEST`: payload/query/path invalid.
- `401 UNAUTHORIZED`: token invalid/tidak ada.
- `404 NOT_FOUND`: resource growth milik user tidak ditemukan.
- `409 CONFLICT`: duplikasi saved search/watchlist.
- `500 INTERNAL_SERVER_ERROR`: kegagalan internal.
