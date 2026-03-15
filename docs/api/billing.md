# Billing API (Internal Bisakerja)

Dokumen ini menjelaskan endpoint billing internal Bisakerja yang mengorkestrasi integrasi ke Midtrans.

## 1) Create Checkout Session

- **Method**: `POST`
- **Path**: `/billing/checkout-session`
- **Auth**: Bearer Token (user)
- **Ownership**: user hanya dapat membuat checkout untuk dirinya sendiri (`JWT.sub`).
- **Rate limit**:
  - inbound user: maksimal 1 request / 10 detik per user.
  - outbound ke Midtrans: dikelola oleh Midtrans SDK.
  - inbound user: maksimal 1 request / 10 detik per user.
  - outbound ke Mayar: target aman <= 18 request/menit/IP.

Endpoint ini:
1. membuat Snap transaction via Midtrans,
2. menyimpan transaksi `pending` di Bisakerja.

### Idempotency

- Client direkomendasikan mengirim header `Idempotency-Key`.
- Jika key sama dipakai ulang oleh user yang sama (payload sama, window 15 menit), API mengembalikan checkout pending yang sudah ada tanpa membuat invoice baru.
- Untuk retry cepat tanpa mengganti payload, API juga dapat me-return checkout `pending` terbaru (`200 Checkout session reused`) selama masih valid, agar user tidak terblokir oleh rate limit saat melanjutkan pembayaran yang sama.

### Request Body

```json
{
  "plan_code": "pro_monthly",
  "customer_mobile": "08123456789",
  "redirect_url": "https://app.bisakerja.com/billing/success"
}
```

### Validation

| Field | Rules |
|---|---|
| `plan_code` | wajib, enum MVP: `pro_monthly`. |
| `customer_mobile` | wajib, nomor telepon customer untuk detail Midtrans; karakter yang diizinkan angka dengan panjang `9..15` digit (spasi, `+`, `-`, `(`, `)` pada input akan dinormalisasi). |
| `redirect_url` | wajib, host harus ada di allowlist backend; skema `https` wajib untuk host non-local, sementara `http` hanya diizinkan untuk local development (`localhost`, `127.0.0.1`, `::1`). |


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
    "provider": "midtrans",
    "plan_code": "pro_monthly",
    "invoice_id": "snap-token-xyz",
    "transaction_id": "checkout:user-uuid:random-key",
    "checkout_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-token-xyz",
    "snap_token": "snap-token-xyz",
    "original_amount": 49000,
    "final_amount": 49000,
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
    "provider": "midtrans",
    "plan_code": "pro_monthly",
    "invoice_id": "snap-token-xyz",
    "transaction_id": "checkout:user-uuid:random-key",
    "checkout_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-token-xyz",
    "snap_token": "snap-token-xyz",
    "original_amount": 49000,
    "final_amount": 49000,
    "expired_at": "2026-03-20T10:00:00Z",
    "subscription_state": "pending_payment",
    "transaction_status": "pending"
  }
}
```

### Error

- `400 BAD_REQUEST` (`INVALID_PLAN_CODE`, `INVALID_CUSTOMER_MOBILE`, `INVALID_REDIRECT_URL`).
- `401 UNAUTHORIZED` token user invalid.
- `409 CONFLICT` (`ALREADY_PREMIUM`) user masih premium aktif.
- `429 TOO_MANY_REQUESTS` (`TOO_MANY_REQUESTS`) rate limit user terlampaui **dan** tidak ada checkout pending valid yang bisa direuse.
- `502 BAD_GATEWAY` (`MIDTRANS_UPSTREAM_ERROR`) upstream Midtrans mengembalikan respons non-kompatibel (verifikasi `MIDTRANS_SERVER_KEY`, `MIDTRANS_CLIENT_KEY`, dan `MIDTRANS_ENV`).
- `503 SERVICE_UNAVAILABLE` (`MIDTRANS_RATE_LIMITED`/`SERVICE_UNAVAILABLE`) retry exhausted atau dependency down.

### Retry Contract ke Midtrans

- Retry hanya untuk `429/5xx`.
- Max retry: 3 kali dengan exponential backoff + jitter (200ms, 400ms, 800ms).
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
      "provider": "midtrans",
      "provider_transaction_id": "checkout:user-uuid:random-key",
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

Reconciliation dijalankan periodik oleh `billing-worker` untuk memastikan status transaksi lokal tetap sinkron dengan status invoice di Midtrans.

### Input Reconciliation

- Sumber transaksi: semua transaksi internal status `pending`/`reminder`.
- Sumber status upstream: `CheckTransaction(orderID)` via Midtrans Core API SDK.

### Status Normalization

| Status Midtrans | Status internal Bisakerja |
|---|---|
| `capture` + fraud `accept` / `settlement` | `success` |
| `pending` | `pending` |
| `cancel` / `expire` / `deny` | `failed` |

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

## 5) Kontrak Integrasi ke Midtrans

- Referensi integrasi Snap: [`midtrans-snap.md`](./midtrans-snap.md).
- Webhook inbound: [`webhooks.md`](./webhooks.md).
