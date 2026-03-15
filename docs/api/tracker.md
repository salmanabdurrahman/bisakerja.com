# Tracker API (Bookmarks & Application Tracking)

Dokumen ini merangkum endpoint backend Phase 6 untuk Application Tracker & Bookmark.

Semua endpoint di dokumen ini:

- user-scoped dan wajib Bearer token valid,
- menggunakan ownership `JWT.sub`,
- runtime path berprefix `/api/v1/*`.

## 1) Bookmarks

### 1.1 Create Bookmark

- **Method**: `POST`
- **Path**: `/bookmarks`

#### Request Body

```json
{
  "job_id": "job_abc123"
}
```

#### Validation

| Field | Rules |
|---|---|
| `job_id` | wajib, 1..100 karakter |

#### Response `201 Created`

```json
{
  "meta": {
    "code": 201,
    "status": "success",
    "message": "Bookmark created",
    "request_id": "req_01J..."
  },
  "data": {
    "id": "bm_abc123",
    "job_id": "job_abc123",
    "created_at": "2026-03-14T13:30:00Z"
  }
}
```

### 1.2 Delete Bookmark

- **Method**: `DELETE`
- **Path**: `/bookmarks/{job_id}`

#### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Bookmark deleted",
    "request_id": "req_01J..."
  },
  "data": {
    "job_id": "job_abc123"
  }
}
```

### 1.3 List Bookmarks

- **Method**: `GET`
- **Path**: `/bookmarks`

#### Response `200 OK`

`data` berupa array bookmark milik user login, diurutkan terbaru lebih dulu.

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Bookmarks retrieved",
    "request_id": "req_01J..."
  },
  "data": [
    {
      "id": "bm_abc123",
      "job_id": "job_abc123",
      "created_at": "2026-03-14T13:30:00Z"
    }
  ]
}
```

## 2) Application Tracking

### 2.1 Create Tracked Application

- **Method**: `POST`
- **Path**: `/applications`

#### Request Body

```json
{
  "job_id": "job_abc123",
  "notes": "Applied via company portal"
}
```

#### Validation

| Field | Rules |
|---|---|
| `job_id` | wajib, 1..100 karakter |
| `notes` | opsional, <= 2000 karakter |

#### Free Tier Limit

User free tier hanya boleh memiliki maksimal **5 active tracked applications** (status bukan `rejected` atau `withdrawn`). Jika limit tercapai, endpoint mengembalikan `403 FORBIDDEN` dengan error code `TRACKER_LIMIT_EXCEEDED`.

#### Response `201 Created`

```json
{
  "meta": {
    "code": 201,
    "status": "success",
    "message": "Tracked application created",
    "request_id": "req_01J..."
  },
  "data": {
    "id": "app_xyz789",
    "job_id": "job_abc123",
    "status": "applied",
    "notes": "Applied via company portal",
    "created_at": "2026-03-14T13:30:00Z",
    "updated_at": "2026-03-14T13:30:00Z"
  }
}
```

### 2.2 Update Application Status

- **Method**: `PATCH`
- **Path**: `/applications/{id}/status`

#### Request Body

```json
{
  "status": "interview"
}
```

#### Validation

| Field | Rules |
|---|---|
| `status` | wajib, enum: `applied`, `interview`, `offer`, `rejected`, `withdrawn` |

#### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Application status updated",
    "request_id": "req_01J..."
  },
  "data": {
    "id": "app_xyz789",
    "job_id": "job_abc123",
    "status": "interview",
    "notes": "Applied via company portal",
    "created_at": "2026-03-14T13:30:00Z",
    "updated_at": "2026-03-14T14:00:00Z"
  }
}
```

### 2.3 Delete Tracked Application

- **Method**: `DELETE`
- **Path**: `/applications/{id}`

#### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Tracked application deleted",
    "request_id": "req_01J..."
  },
  "data": {
    "id": "app_xyz789"
  }
}
```

### 2.4 List Tracked Applications

- **Method**: `GET`
- **Path**: `/applications`

#### Response `200 OK`

`data` berupa array tracked application milik user login.

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Tracked applications retrieved",
    "request_id": "req_01J..."
  },
  "data": [
    {
      "id": "app_xyz789",
      "job_id": "job_abc123",
      "status": "interview",
      "notes": "Applied via company portal",
      "created_at": "2026-03-14T13:30:00Z",
      "updated_at": "2026-03-14T14:00:00Z"
    }
  ]
}
```

## Error Ringkas

- `400 BAD_REQUEST`: payload/path invalid atau field tidak memenuhi validasi.
- `401 UNAUTHORIZED`: token invalid/tidak ada.
- `403 TRACKER_LIMIT_EXCEEDED`: user free tier mencapai batas 5 active tracked applications.
- `404 NOT_FOUND`: bookmark atau tracked application milik user tidak ditemukan.
- `409 CONFLICT`: duplikasi bookmark atau tracked application pada job yang sama.
- `500 INTERNAL_SERVER_ERROR`: kegagalan internal.

## Enum Referensi

| Domain | Field | Allowed Values |
|---|---|---|
| Application status | `status` | `applied`, `interview`, `offer`, `rejected`, `withdrawn` |

Catatan: status `rejected` dan `withdrawn` dianggap **tidak aktif** untuk keperluan perhitungan free tier limit.
