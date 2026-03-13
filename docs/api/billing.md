# Billing API (Internal Bisakerja)

Dokumen ini menjelaskan endpoint billing internal Bisakerja yang mengorkestrasi integrasi ke Mayar.

## 1) Create Checkout Session

- **Method**: `POST`
- **Path**: `/billing/checkout-session`
- **Auth**: Bearer Token (user)
- **Ownership**: user hanya dapat membuat checkout untuk dirinya sendiri (`JWT.sub`).
- **Rate limit**:
  - inbound user: maksimal 1 request / 10 detik per user.
  - outbound ke Mayar: target aman <= 18 request/menit/IP.

Endpoint ini:
1. memastikan customer tersedia di Mayar,
2. membuat invoice via Mayar,
3. menyimpan transaksi `pending` di Bisakerja.

### Idempotency

- Client direkomendasikan mengirim header `Idempotency-Key`.
- Jika key sama dipakai ulang oleh user yang sama (payload sama, window 15 menit), API mengembalikan checkout pending yang sudah ada tanpa membuat invoice baru.

### Request Body

```json
{
  "plan_code": "pro_monthly",
  "redirect_url": "https://app.bisakerja.com/billing/success"
}
```

### Validation

| Field | Rules |
|---|---|
| `plan_code` | wajib, enum MVP: `pro_monthly`. |
| `redirect_url` | wajib, URL HTTPS valid, host harus ada di allowlist backend. |

### Response `201 Created`

```json
{
  "meta": {
    "code": 201,
    "status": "success",
    "message": "Checkout session created",
    "request_id": "req_01J..."
  },
  "data": {
    "provider": "mayar",
    "invoice_id": "f774034d-d9cc-43a0-97d8-a2520c127f03",
    "transaction_id": "23fa41c5-c6ed-45d4-8302-5fac4a165dfa",
    "checkout_url": "https://andiak.myr.id/invoices/ibzfrf4880",
    "expired_at": "2026-03-20T10:00:00Z",
    "subscription_state": "pending_payment",
    "transaction_status": "pending"
  }
}
```

### Error

- `400 BAD_REQUEST` (`INVALID_PLAN_CODE`, `INVALID_REDIRECT_URL`).
- `401 UNAUTHORIZED` token user invalid.
- `409 CONFLICT` (`ALREADY_PREMIUM`) user masih premium aktif.
- `429 TOO_MANY_REQUESTS` (`TOO_MANY_REQUESTS`) rate limit user terlampaui.
- `502 BAD_GATEWAY` (`MAYAR_UPSTREAM_ERROR`) upstream Mayar gagal merespons valid.
- `503 SERVICE_UNAVAILABLE` (`MAYAR_RATE_LIMITED`/`SERVICE_UNAVAILABLE`) retry exhausted atau dependency down.

### Retry Contract ke Mayar

- Retry hanya untuk `429/5xx`.
- Max retry: 3 kali dengan exponential backoff + jitter (200ms, 400ms, 800ms).
- Timeout outbound per request: 5 detik.

## 2) Get My Billing Status

- **Method**: `GET`
- **Path**: `/billing/status`
- **Auth**: Bearer Token (user)
- **Ownership**: selalu menggunakan user dari token.

### `subscription_state` Rules (Canonical)

- `free`: user belum pernah premium aktif.
- `pending_payment`: ada transaksi terbaru status `pending`/`reminder` dan belum sukses.
- `premium_active`: transaksi sukses terakhir ada dan `premium_expired_at > now()`.
- `premium_expired`: pernah sukses, tetapi `premium_expired_at <= now()`.

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Billing status retrieved",
    "request_id": "req_01J..."
  },
  "data": {
    "plan_code": "pro_monthly",
    "subscription_state": "premium_active",
    "is_premium": true,
    "premium_expired_at": "2026-04-20T10:00:00Z",
    "last_transaction_status": "success"
  }
}
```

## 3) Get My Billing History

- **Method**: `GET`
- **Path**: `/billing/transactions`
- **Auth**: Bearer Token (user)
- **Ownership**: hanya transaksi milik user token.

### Query Parameters

| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| `page` | integer | no | 1 | Nomor halaman (`>=1`). |
| `limit` | integer | no | 20 | Jumlah data per halaman (`1..100`). |
| `status` | string | no | - | Filter enum `pending`, `reminder`, `success`, `failed`. |

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Transactions retrieved",
    "request_id": "req_01J...",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total_pages": 1,
      "total_records": 2
    }
  },
  "data": [
    {
      "id": "0f8fad5b-d9cb-469f-a165-70867728950e",
      "provider": "mayar",
      "mayar_transaction_id": "23fa41c5-c6ed-45d4-8302-5fac4a165dfa",
      "amount": 49000,
      "status": "success",
      "created_at": "2026-03-13T18:20:00Z"
    }
  ]
}
```

## Kontrak Integrasi ke Mayar

- Referensi endpoint resmi: [`mayar-headless.md`](./mayar-headless.md).
- Webhook inbound: [`webhooks.md`](./webhooks.md).
