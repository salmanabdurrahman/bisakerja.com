# API Documentation Hub

Dokumen ini adalah pintu masuk untuk seluruh kontrak endpoint Bisakerja API.

## Base URL

- Production: `https://api.bisakerja.com/api/v1`
- Development: `http://localhost:8080/api/v1`

## Konvensi Umum

### Endpoint Path Convention

- Field `Path` di dokumen endpoint **tidak** menyertakan prefix `/api/v1`.
- Rumus URL final: `{base_url}{path}`.
- Contoh: base URL `https://api.bisakerja.com/api/v1` + path `/jobs` -> `https://api.bisakerja.com/api/v1/jobs`.

### Authentication

- Header: `Authorization: Bearer <access_token>`.
- Endpoint public tidak memerlukan token.
- Endpoint admin memerlukan token valid + role `admin`.

### Ownership Boundaries

- Endpoint user-scoped **wajib** memakai `user_id` dari `JWT.sub`, bukan dari request body/query/path.
- Jika user mencoba mengakses resource user lain -> `403 FORBIDDEN`.
- Endpoint admin boleh akses lintas user hanya di namespace `/admin/*`.

### Content Type

- Request: `application/json`
- Response: `application/json`

### Success Envelope

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "OK",
    "request_id": "req_01J..."
  },
  "data": {}
}
```

### Error Envelope

```json
{
  "meta": {
    "code": 400,
    "status": "error",
    "message": "Validation error",
    "request_id": "req_01J..."
  },
  "errors": [
    {
      "field": "email",
      "code": "INVALID_EMAIL",
      "message": "email is required"
    }
  ]
}
```

### Canonical Enums

| Domain | Field | Allowed Values | Notes |
|---|---|---|---|
| Subscription state | `subscription_state` | `free`, `pending_payment`, `premium_active`, `premium_expired` | Canonical source: `GET /billing/status`. |
| Transaction status | `transactions.status` | `pending`, `reminder`, `success`, `failed` | Status internal Bisakerja (bukan raw status Mayar). |
| Webhook processing | `webhook_deliveries.processing_status` | `processed`, `ignored_duplicate`, `rejected` | Untuk audit idempotency inbound webhook. |

### Pagination

- Query: `page`, `limit`
- Default: `page=1`, `limit=20`
- Max: `limit=100`

Contoh:

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
      "total_pages": 12,
      "total_records": 236
    }
  },
  "data": []
}
```

### Idempotency & Retry Contract

- `POST /billing/checkout-session` menerima header `Idempotency-Key` (direkomendasikan, unik per aksi checkout user).
- `POST /webhook/mayar` wajib idempotent berdasarkan `mayar:{event}:{transactionId}`.
- Duplicate webhook yang sudah diproses harus mengembalikan `200 OK` tanpa side effect ulang.
- Outbound call ke Mayar untuk `429/5xx`: retry max 3 kali (exponential backoff + jitter), lalu fail dengan `503 SERVICE_UNAVAILABLE`.

## Dokumen Endpoint

- [Authentication](./auth.md)
- [Jobs](./jobs.md)
- [User Preferences](./preferences.md)
- [Billing](./billing.md)
- [Admin](./admin.md)
- [Webhooks](./webhooks.md)
- [Error Codes](./errors.md)

## Integrasi Payment Gateway

- [Mayar Headless API Mapping](./mayar-headless.md)
