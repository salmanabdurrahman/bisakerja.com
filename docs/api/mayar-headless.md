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

Catatan progress saat ini: increment awal Phase 4 sudah mulai memakai `GET /coupon/validate` pada flow internal `POST /api/v1/billing/checkout-session` ketika request menyertakan `coupon_code`.

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

## 8) Variasi Payload yang Harus Ditoleransi

Agar flow checkout tetap stabil lintas versi payload Mayar, parser Bisakerja wajib toleran pada variasi berikut:

### 8.1 Struktur root `data`

- format object: `data: { ... }`
- format array: `data: [{ ... }]`

### 8.2 Customer ID (hasil `POST /customer/create`)

Parser menerima beberapa field:

- `data.id`
- `data.customer_id` / `data.customerId`
- `data.response` (varian docs lama)
- fallback root: `id` / `response`

### 8.3 Invoice fields (hasil `POST /invoice/create`)

Parser menerima beberapa field:

- invoice ID: `data.id`, `data.invoice_id`, `data.invoice.id`
- transaction ID: `data.transactionId`, `data.transaction_id`, `data.transaction.id`
- checkout link: `data.invoiceUrl`, `data.checkoutUrl`, `data.link`, `data.paymentLink`
- expiry:
  - RFC3339 string (`2026-03-20T10:00:00Z`), atau
  - unix epoch timestamp (millisecond/second)

### 8.4 Invoice reconciliation fields (hasil `GET /invoice/{id}`)

Parser reconciliation menerima beberapa variasi field:

- transaction ID: `data.transactionId`, `data.transaction_id`, `data.transactions[0].id`
- customer email: `data.customerEmail`, `data.customer_email`, `data.customer.email`
- updated time:
  - RFC3339 string (`updatedAt` / `updated_at`), atau
  - unix epoch timestamp (millisecond/second)

### 8.5 Outbound request compatibility (Bisakerja -> Mayar)

Agar kompatibel dengan variasi dokumentasi/sandbox, payload outbound checkout dikirim dengan field modern **dan** alias legacy:

- `POST /customer/create`
  - modern: `name`, `email`, `mobile` (sandbox saat ini mewajibkan `mobile`)
- `POST /invoice/create`
  - modern:
    - customer: `name`, `email`, `mobile`
    - redirect: `redirectUrl`
    - amount/items: `items[].quantity`, `items[].rate`, `items[].description`
    - expiry: `expiredAt` (RFC3339)
    - metadata: `extraData`
  - alias legacy (tetap dikirim untuk kompatibilitas):
    - `customer_id`, `description`, `item`, `success_redirect_url`, `external_id`, `extra_data`
- `GET /coupon/validate`
  - query tetap mengirim alias umum (`coupon_code` + `code`, `amount`)
  - parser response menerima varian:
    - `discount_amount`/`final_amount`, atau
    - `coupon.discountType` + `coupon.discountValue` (percentage/fixed amount)

## 9) Troubleshooting `MAYAR_UPSTREAM_ERROR` (502)

Jika API internal mengembalikan:

- `code`: `MAYAR_UPSTREAM_ERROR`
- `message`: `mayar upstream returned invalid response ...`

lakukan checklist berikut:

1. Verifikasi env backend:
   - `MAYAR_BASE_URL` sesuai environment (`.id` prod / `.club` sandbox),
   - `MAYAR_API_KEY` valid untuk environment tersebut.
2. Pastikan endpoint checkout memakai payload kompatibel (lihat bagian 8).
3. Validasi respons Mayar yang diterima:
   - status code non-2xx (selain 429/5xx) akan dipetakan sebagai upstream error,
   - payload JSON tanpa field minimal (`invoice_id`, `transaction_id`, `checkout link`) juga dipetakan upstream error.
4. Jika `GET /invoice/{id}` tidak punya `transactionId` root, pastikan respons tetap memiliki `transactions[0].id`.
5. Cek log backend:
   - `mayar upstream non-success response` untuk status non-2xx (`status_code`, `request_id`, `response_body`),
   - `mayar upstream invalid json response` untuk body yang tidak valid JSON.
6. Jika `response_body` berisi `Duplicate request detected`, gunakan checkout pending yang sudah ada (reuse) atau tunggu sekitar 1 menit sebelum create invoice baru.
7. Gunakan `request_id` response Bisakerja untuk korelasi log API saat investigasi.
