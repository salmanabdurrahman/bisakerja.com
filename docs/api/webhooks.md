# Webhooks API

## 1) Mayar Webhook Receiver

- **Method**: `POST`
- **Path**: `/webhook/mayar`
- **Auth**: token endpoint webhook + idempotency validation

Endpoint ini menerima callback event dari Mayar untuk sinkronisasi status pembayaran/subscription.

## Event yang Diproses (MVP)

- `payment.received`
- `payment.reminder`

Event membership (`membership.*`) disiapkan untuk post-MVP.

## Request Body Example (Mayar)

```json
{
  "event": "payment.received",
  "data": {
    "id": "43b2f0ce-03f2-4f59-a341-299ea3ef19b6",
    "transactionId": "43b2f0ce-03f2-4f59-a341-299ea3ef19b6",
    "transactionStatus": "paid",
    "createdAt": "2025-04-29T09:35:43.635Z",
    "updatedAt": "2025-04-29T09:35:43.635Z",
    "customerEmail": "user@example.com",
    "customerName": "Budi",
    "amount": 49000,
    "productId": "688b6a9f-2893-4b8a-a637-a008d91d0cfc",
    "productType": "membership"
  }
}
```

## Validasi Keamanan

Karena dokumentasi webhook Mayar tidak mendefinisikan signature header wajib, validasi minimum di Bisakerja:

1. Verifikasi token endpoint (`X-Bisakerja-Webhook-Token` atau secret path/query yang ekuivalen).
2. Tolak payload tanpa field inti: `event`, `data.transactionId`, `data.customerEmail`.
3. Simpan payload raw untuk audit.
4. Tambahkan `request_id` ke log untuk traceability.

## Idempotency Contract

- Kunci idempotency: `mayar:{event}:{transactionId}`.
- Simpan/cek di `webhook_deliveries.idempotency_key`.
- Jika event dengan key yang sama sudah `processed`, kembalikan `200 OK` tanpa side effect.
- Semua mutasi `transactions` + `users` dilakukan dalam **satu transaksi database**.

## Normalisasi Status Transaksi

| Event / Kondisi | `transactions.status` | Dampak ke premium |
|---|---|---|
| `payment.received` + status paid/success | `success` | aktifkan/extend premium |
| `payment.reminder` | `reminder` | tidak mengaktifkan premium |
| invoice dibuat (sebelum webhook) | `pending` | belum aktif premium |
| status gagal/expired final | `failed` | pastikan premium tidak aktif |

## Perilaku Processing

- `payment.received` + status paid/success -> update transaksi `success` + update user premium.
- `payment.reminder` -> update transaksi `reminder`, tanpa aktivasi premium.
- Event tidak dikenali -> simpan sebagai `rejected`/ignored, kembalikan response aman.
- User tidak ditemukan dari `customerEmail` -> `422 UNPROCESSABLE_ENTITY` + catat `error_message`.

## Response

### `200 OK` (processed atau duplicate)

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Webhook processed",
    "request_id": "req_01J..."
  },
  "data": {
    "provider": "mayar",
    "processed": true,
    "idempotent": false
  }
}
```

Untuk duplicate idempotent, set `idempotent: true`.

### Error

- `400 BAD_REQUEST` (`INVALID_WEBHOOK_PAYLOAD`) jika payload tidak valid.
- `401 UNAUTHORIZED` (`INVALID_WEBHOOK_TOKEN`) jika token webhook tidak valid.
- `422 UNPROCESSABLE_ENTITY` (`WEBHOOK_USER_NOT_FOUND`) jika payload valid format tapi gagal aturan bisnis.
- `503 SERVICE_UNAVAILABLE` jika dependency internal gagal (agar Mayar dapat retry).

## Retry Expectations

- Bisakerja mengembalikan status non-2xx jika proses gagal transient agar Mayar dapat retry.
- Untuk replay manual, gunakan endpoint operasional webhook Mayar (lihat di bawah).

## Operasional Webhook di Mayar

Endpoint resmi Mayar untuk operasional webhook:

- Register URL: `GET /hl/v1/webhook/register`
- Test URL: `POST /hl/v1/webhook/test`
- History: `GET /hl/v1/webhook/history`
- Retry: `POST /hl/v1/webhook/retry`

Lihat mapping lengkap: [`./mayar-headless.md`](./mayar-headless.md).
