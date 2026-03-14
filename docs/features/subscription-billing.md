# Subscription & Billing

## Tujuan

Mengelola upgrade user dari Free ke Pro melalui integrasi Mayar yang aman, idempotent, dan mudah diaudit.

## Prioritas

**MVP** dengan hardening integrasi di phase berikutnya.

## Scope Fitur

### In Scope (MVP)

- Endpoint internal checkout: `POST /api/v1/billing/checkout-session`.
- Endpoint read billing user: `GET /api/v1/billing/status` dan `GET /api/v1/billing/transactions`.
- Sinkronisasi customer dan invoice ke Mayar.
- Proses webhook pembayaran dari Mayar (`payment.received`, `payment.reminder`).
- Reconciliation periodik status invoice Mayar ke transaksi internal.
- Aktivasi status premium otomatis setelah pembayaran valid.
- Pencatatan transaksi + raw webhook untuk audit.

### Out of Scope (MVP)

- Multi-plan dengan proration kompleks.
- Refund/chargeback automation.
- Kupon, cicilan, QR checkout lanjutan (opsional post-MVP).

## Kontrak Mayar yang Dipakai (MVP)

| Use Case | Endpoint Mayar |
|---|---|
| Create customer | `POST /hl/v1/customer/create` |
| Create invoice | `POST /hl/v1/invoice/create` |
| Invoice detail | `GET /hl/v1/invoice/{id}` |
| Paid transactions | `GET /hl/v1/transactions?page=1&pageSize=10` |
| Register webhook URL | `GET /hl/v1/webhook/register` |
| Test webhook URL | `POST /hl/v1/webhook/test` |
| Webhook history/retry | `GET /hl/v1/webhook/history`, `POST /hl/v1/webhook/retry` |

Lihat detail: [`../api/mayar-headless.md`](../api/mayar-headless.md).

## Aturan Bisnis

1. Premium aktif hanya jika transaksi internal berstatus `success`.
2. Event `payment.reminder` mengubah status transaksi ke `reminder`, tanpa aktivasi premium.
3. Webhook wajib idempotent (`mayar:{event}:{transactionId}`).
4. Semua perubahan status pembayaran harus terekam di `transactions`.
5. Integrasi outbound ke Mayar harus mengikuti limit 20 request/menit/IP (target internal <=18).
6. Endpoint billing user-scoped hanya boleh mengakses data user dari `JWT.sub`.

## State & Enum Canonical

### Subscription State (read model)

- `free` -> `pending_payment` -> `premium_active` -> `premium_expired`

### Transaction Status (write/audit model)

- `pending`, `reminder`, `success`, `failed`

## Dependensi Endpoint Internal

- `POST /api/v1/billing/checkout-session`
- `GET /api/v1/billing/status`
- `GET /api/v1/billing/transactions`
- `POST /api/v1/webhook/mayar`
- `GET /api/v1/auth/me`

Lihat detail: [`../api/billing.md`](../api/billing.md), [`../api/webhooks.md`](../api/webhooks.md), [`../api/auth.md`](../api/auth.md).

## Dampak Database

- `users`: sinkron field premium + `mayar_customer_id`.
- `transactions`: simpan transaksi gateway + payload.
- `webhook_deliveries`: audit event webhook masuk.

Lihat detail: [`../architecture/database.md`](../architecture/database.md), [`../architecture/mayar-integration.md`](../architecture/mayar-integration.md).

## Edge Cases & Handling

- Event duplikat -> response sukses idempotent tanpa update ganda.
- User tidak ditemukan saat webhook masuk -> `422` + tandai untuk rekonsiliasi.
- Upstream Mayar `429/5xx` -> retry max 3x + backoff; gagal akhir -> `503`.
- User double-click upgrade -> dibatasi rate limit 1 request / 10 detik + idempotency key.
- Transaksi `pending/reminder` stale >24 jam -> terdeteksi anomaly summary pada billing worker.

## NFR Minimum (MVP)

| Metric | Target |
|---|---|
| Checkout API p95 latency | < 2 detik |
| Webhook processing p95 latency | < 500 ms |
| Webhook success ratio | >= 99% |
| Mayar rate-limit hit ratio | < 1% per jam |

## Acceptance Criteria

- Checkout valid menghasilkan `checkout_url` dan transaksi `pending`.
- Pembayaran sukses (`payment.received`) mengaktifkan premium dalam <= 30 detik.
- Event reminder/gagal tidak memberi akses premium.
- Duplikasi webhook tidak menyebabkan double activation.
- Endpoint status/history menampilkan entitlement premium dan riwayat transaksi user-scoped.
- Reconciliation worker dapat mengoreksi mismatch status transaksi berdasarkan invoice Mayar.
- Histori transaksi dan webhook dapat ditelusuri end-to-end via `request_id` + tabel audit.
