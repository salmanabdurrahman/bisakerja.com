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
  "coupon_code": "SAVE10",
  "redirect_url": "https://app.bisakerja.com/billing/success"
}
```

### Validation

| Field | Rules |
|---|---|
| `plan_code` | wajib, enum MVP: `pro_monthly`. |
| `coupon_code` | opsional, jika diisi harus alfanumerik (`A-Z`, `0-9`, `-`, `_`) panjang `3..64`, dan valid di Mayar. |
| `redirect_url` | wajib, host harus ada di allowlist backend; skema `https` wajib untuk host non-local, sementara `http` hanya diizinkan untuk local development (`localhost`, `127.0.0.1`, `::1`). |

Jika `coupon_code` dikirim:

1. backend memvalidasi kode ke Mayar (`GET /hl/v1/coupon/validate`),
2. nominal invoice checkout memakai `final_amount` setelah diskon,
3. response tetap canonical dengan tambahan metadata amount.

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
    "plan_code": "pro_monthly",
    "invoice_id": "f774034d-d9cc-43a0-97d8-a2520c127f03",
    "transaction_id": "23fa41c5-c6ed-45d4-8302-5fac4a165dfa",
    "checkout_url": "https://andiak.myr.id/invoices/ibzfrf4880",
    "original_amount": 49000,
    "discount_amount": 10000,
    "final_amount": 39000,
    "coupon_code": "SAVE10",
    "expired_at": "2026-03-20T10:00:00Z",
    "subscription_state": "pending_payment",
    "transaction_status": "pending"
  }
}
```

### Response `200 OK` (Idempotent Reuse)

Jika request mengirim `Idempotency-Key` yang sama (masih dalam window 15 menit), API dapat mengembalikan checkout pending sebelumnya tanpa membuat invoice baru:

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Checkout session reused",
    "request_id": "req_01J..."
  },
  "data": {
    "provider": "mayar",
    "plan_code": "pro_monthly",
    "invoice_id": "f774034d-d9cc-43a0-97d8-a2520c127f03",
    "transaction_id": "23fa41c5-c6ed-45d4-8302-5fac4a165dfa",
    "checkout_url": "https://andiak.myr.id/invoices/ibzfrf4880",
    "original_amount": 49000,
    "discount_amount": 10000,
    "final_amount": 39000,
    "coupon_code": "SAVE10",
    "expired_at": "2026-03-20T10:00:00Z",
    "subscription_state": "pending_payment",
    "transaction_status": "pending"
  }
}
```

### Error

- `400 BAD_REQUEST` (`INVALID_PLAN_CODE`, `INVALID_COUPON_CODE`, `INVALID_REDIRECT_URL`).
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

### Error

- `401 UNAUTHORIZED` token user invalid atau tidak tersedia.
- `500 INTERNAL_SERVER_ERROR` gagal membaca profil/riwayat billing user.

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

### Error

- `400 BAD_REQUEST` query `page`, `limit`, atau `status` tidak valid.
- `401 UNAUTHORIZED` token user invalid atau tidak tersedia.
- `500 INTERNAL_SERVER_ERROR` gagal membaca histori transaksi user.

## 4) Reconciliation & Recovery (Internal Worker)

Reconciliation dijalankan periodik oleh `billing-worker` untuk memastikan status transaksi lokal tetap sinkron dengan status invoice di Mayar.

### Input Reconciliation

- Sumber transaksi: semua transaksi internal status `pending`/`reminder`.
- Sumber status upstream: `GET /hl/v1/invoice/{id}` ke Mayar.

### Status Normalization

| Status invoice Mayar | Status internal Bisakerja |
|---|---|
| `paid`, `success`, `completed` | `success` |
| `reminder` | `reminder` |
| `pending`, `unpaid`, `open`, `waiting` | `pending` |
| `failed`, `expired`, `cancelled`, `canceled`, `void` | `failed` |

Jika rekonsiliasi menemukan status berubah menjadi `success`, sistem akan mengaktifkan premium user terkait sesuai kontrak `subscription_state`.

### Retry & Failure Contract

- Outbound call rekonsiliasi mengikuti retry policy yang sama dengan checkout:
  - retry hanya untuk `429/5xx`,
  - max retry 3 kali,
  - exponential backoff + jitter (`200ms`, `400ms`, `800ms`).
- Jika tetap gagal setelah retry, transaksi dihitung sebagai `retryable_failure` dan diproses pada tick worker berikutnya.

### Anomaly Signal

- Worker menandai anomali jika transaksi masih `pending`/`reminder` lebih dari 24 jam.
- Ringkasan reconciliation setiap tick:
  - `scanned_transactions`,
  - `reconciled`,
  - `retryable_failures`,
  - `anomaly_count`.

## 5) Kontrak Integrasi ke Mayar

- Referensi endpoint resmi: [`mayar-headless.md`](./mayar-headless.md).
- Webhook inbound: [`webhooks.md`](./webhooks.md).
