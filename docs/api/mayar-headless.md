# Mayar Headless API Mapping

Dokumen ini menyelaraskan kebutuhan Bisakerja dengan dokumentasi resmi Mayar (`/api-reference` dan `/integration/webhook`).

## 1) Dasar Integrasi

### Base URL

- Production: `https://api.mayar.id/hl/v1`
- Sandbox: `https://api.mayar.club/hl/v1`

### Authentication

- Header: `Authorization: Bearer <MAYAR_API_KEY>`

### Status Code (dokumen resmi)

- `200`, `400`, `401`, `404`, `429`, `500`

### Rate Limit (dokumen resmi)

- **20 requests/minute per IP**
- Saat limit terlampaui: `429 Too Many Requests` + header `Retry-After`

## 2) Endpoint Mayar yang Dipakai Bisakerja (MVP)

| Domain | Method | Endpoint Mayar | Pemakaian di Bisakerja |
|---|---|---|---|
| Customer | `POST` | `/customer/create` | Sinkronisasi customer saat checkout pertama |
| Customer | `GET` | `/customer?page=1&pageSize=10` | Rekonsiliasi customer (opsional operasional) |
| Customer | `POST` | `/customer/update` | Perubahan email customer (opsional operasional) |
| Invoice | `POST` | `/invoice/create` | Membuat checkout Pro |
| Invoice | `POST` | `/invoice/edit` | Koreksi invoice pending |
| Invoice | `GET` | `/invoice` | Monitoring invoice merchant |
| Invoice | `GET` | `/invoice/{id}` | Validasi status invoice tertentu |
| Invoice | `GET` | `/invoice/close/{id}` | Menutup invoice tidak terpakai |
| Invoice | `GET` | `/invoice/open/{id}` | Membuka kembali invoice |
| Transaction | `GET` | `/transactions?page=1&pageSize=10` | Rekonsiliasi transaksi paid |
| Transaction | `GET` | `/transactions/unpaid?page=1&pageSize=10` | Monitoring unpaid transactions |
| Webhook | `GET` | `/webhook/history?page=1&pageSize=10` | Audit pengiriman webhook |
| Webhook | `GET` | `/webhook/register` | Registrasi URL webhook Bisakerja |
| Webhook | `POST` | `/webhook/test` | Uji endpoint webhook Bisakerja |
| Webhook | `POST` | `/webhook/retry` | Retry event webhook yang gagal |

## 3) Endpoint Mayar Opsional (Post-MVP)

| Use Case | Method | Endpoint Mayar |
|---|---|---|
| Single payment request | `POST` | `/payment/create` |
| Edit payment request | `POST` | `/payment/edit` |
| Close/Open payment request | `GET` | `/payment/close/{id}`, `/payment/open/{id}` |
| Installment plan | `POST` | `/installment/create` |
| Coupon validation | `GET` | `/coupon/validate` |
| Dynamic QR checkout | `POST` | `/qrcode/create` |
| Account balance | `GET` | `/balance` |
| SaaS license verify | `POST` | `https://api.mayar.id/saas/v1/license/verify` |
| SaaS license activate/deactivate | `POST` | `https://api.mayar.id/saas/v1/license/activate`, `https://api.mayar.id/saas/v1/license/deactivate` |

## 4) Webhook Event yang Relevan

Berdasarkan dokumentasi integration webhook Mayar:

- `payment.received` -> pembayaran selesai/diterima.
- `payment.reminder` -> pengingat untuk transaksi yang belum diselesaikan.
- `membership.memberUnsubscribed` (opsional).
- `membership.memberExpired` (opsional).
- `membership.changeTierMemberRegistered` (opsional).
- `membership.newMemberRegistered` (opsional).

Untuk MVP Bisakerja:
- aktivasi premium pada `payment.received` dengan `data.transactionStatus=paid`.
- `payment.reminder` tidak mengaktifkan premium.

## 5) Mapping Payload Mayar ke Data Internal

| Payload Mayar | Tabel/Kolom Internal | Catatan |
|---|---|---|
| `event` | `webhook_deliveries.event_type` | Dipakai untuk routing logic |
| `data.transactionId` | `transactions.mayar_transaction_id` | Kunci idempotensi utama |
| `data.productId` | `transactions.mayar_payment_link_id` | Referensi payment link/invoice |
| `data.amount` | `transactions.amount` | Nominal transaksi |
| `data.transactionStatus` | `transactions.status` | Dinormalisasi ke `pending/reminder/success/failed` |
| `data.customerEmail` | lookup ke `users.email` | Untuk menemukan user lokal |
| payload raw | `webhook_deliveries.payload_raw`, `transactions.raw_payload` | Audit dan debugging |

### Normalization Reference

| Raw Mayar signal | Internal `transactions.status` |
|---|---|
| invoice dibuat, belum ada event sukses | `pending` |
| `event=payment.reminder` | `reminder` |
| `event=payment.received` + `transactionStatus in (paid, success)` | `success` |
| event negatif/final failure (expired/canceled/failed) | `failed` |

## 6) Reliability Contract untuk Integrasi

- Outbound request ke Mayar **wajib di-throttle** (target aman: <=18 req/minute per IP).
- Retry policy saat `429/5xx`: max 3 kali, exponential backoff (`200ms`, `400ms`, `800ms`) + jitter.
- Timeout outbound per request: 5 detik.
- Wajib idempotent saat memproses webhook (`event + transactionId`).
- Simpan raw payload untuk audit dan replay.

## 7) Catatan Konsistensi Dokumen Mayar

Beberapa halaman contoh request memiliki inkonsistensi kecil (contoh path di curl tidak sama dengan endpoint final). Untuk implementasi Bisakerja:

1. Jadikan blok **Endpoint (Production/Sandbox)** sebagai sumber kebenaran.
2. Jadikan contoh response sebagai referensi bentuk payload.
3. Simpan fallback parser agar toleran terhadap variasi field (misal `transaction_id` vs `transactionId`).
