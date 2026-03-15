# Webhooks API

## 1) Midtrans Webhook Receiver

- **Method**: `POST`
- **Path**: `/webhook/midtrans`
- **Auth**: validasi signature Midtrans (SHA512) + idempotency validation

Endpoint ini menerima notifikasi pembayaran dari Midtrans untuk sinkronisasi status transaksi.

## Event yang Diproses (MVP)

Midtrans mengirim notifikasi flat JSON dengan field `transaction_status` dan `fraud_status`. Event yang memicu perubahan status:

- `settlement` → `success`
- `capture` + `fraud_status=accept` → `success`
- `pending` → `pending`
- `cancel` / `expire` / `deny` → `failed`

## Request Body Example (Midtrans)

```json
{
  "order_id": "pay-a1b2c3d4e5f6a7b8-c9d0e1f2",
  "transaction_status": "settlement",
  "fraud_status": "accept",
  "gross_amount": "49000.00",
  "status_code": "200",
  "signature_key": "<sha512hash>"
}
```

## Validasi Keamanan

Midtrans menggunakan signature validation berbasis SHA512:

1. **Signature formula**: `SHA512(order_id + status_code + gross_amount + server_key)`.
2. Tolak payload jika signature tidak cocok.
3. Tolak payload tanpa field inti: `order_id`, `transaction_status`, `gross_amount`, `status_code`, `signature_key`.
4. Simpan payload raw untuk audit.
5. Tambahkan `request_id` ke log untuk traceability.

## Idempotency Contract

- Kunci idempotency: `midtrans:{order_id}:{transaction_status}`.
- Simpan/cek di `webhook_deliveries.idempotency_key`.
- Jika event dengan key yang sama sudah `processed`, kembalikan `200 OK` tanpa side effect.
- Semua mutasi `transactions` + `users` dilakukan dalam **satu transaksi database**.

## Normalisasi Status Transaksi

| `transaction_status` + `fraud_status` | `transactions.status` | Dampak ke premium |
|---|---|---|
| `capture` + `accept` / `settlement` | `success` | aktifkan/extend premium |
| `pending` | `pending` | belum aktif premium |
| `cancel` / `expire` / `deny` | `failed` | pastikan premium tidak aktif |

## Perilaku Processing

- `settlement` / `capture+accept` → update transaksi `success` + update user premium.
- `pending` → update transaksi `pending`, tanpa aktivasi premium.
- Event tidak dikenali → simpan sebagai `rejected`/ignored, kembalikan response aman.
- User tidak ditemukan dari `order_id` → `422 UNPROCESSABLE_ENTITY` + catat `error_message`.

## Order ID Format

Format `order_id`: `pay-{16hex}-{8hex}` (29 karakter, selalu ≤ 50 karakter).

Contoh: `pay-a1b2c3d4e5f6a7b8-c9d0e1f2`.

Backend me-lookup transaksi berdasarkan `order_id` (= `ProviderTransactionID`) di database untuk menemukan `UserID` terkait.
Format tidak mengandung `userID` secara eksplisit — resolusi owner dilakukan lewat DB lookup.
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
    "provider": "midtrans",
    "processed": true,
    "idempotent": false
  }
}
```

Untuk duplicate idempotent, set `idempotent: true`.

### Error

- `400 BAD_REQUEST` (`INVALID_WEBHOOK_PAYLOAD`) jika payload tidak valid atau signature salah.
- `422 UNPROCESSABLE_ENTITY` (`WEBHOOK_USER_NOT_FOUND`) jika payload valid format tapi gagal aturan bisnis.
- `503 SERVICE_UNAVAILABLE` jika dependency internal gagal (agar Midtrans dapat retry).

## Retry Expectations

- Bisakerja mengembalikan status non-2xx jika proses gagal transient agar Midtrans dapat retry.
- Midtrans akan retry webhook otomatis pada status non-2xx.

## Referensi

- Dokumentasi Midtrans Snap: [`./midtrans-snap.md`](./midtrans-snap.md).
- Untuk replay manual, gunakan Midtrans dashboard → Notification → Resend.
