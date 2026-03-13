# Upgrade Billing Flow (Frontend)

## Tujuan

Menggambarkan journey upgrade user dari `free` menjadi `premium_active`, sinkron dengan kontrak Billing API dan event webhook Mayar.

## Happy Path

```mermaid
sequenceDiagram
  participant U as User
  participant FE as Frontend (Web)
  participant API as Bisakerja API
  participant M as Mayar

  U->>FE: Klik CTA "Upgrade ke Pro"
  FE->>API: GET /api/v1/auth/me
  API-->>FE: 200 (is_premium=false)
  FE-->>U: Render billing_ready

  U->>FE: Konfirmasi paket Pro
  FE->>API: POST /api/v1/billing/checkout-session
  API-->>FE: 201 (checkout_url, transaction_id, subscription_state=pending_payment, transaction_status=pending)
  FE-->>U: Render checkout_redirecting
  FE->>M: Redirect ke checkout_url

  U->>M: Selesaikan pembayaran
  M->>API: POST /api/v1/webhook/mayar (event=payment.received, transactionStatus=paid)
  API-->>M: 200 Webhook processed

  U->>FE: Kembali ke redirect_url aplikasi
  FE->>API: GET /api/v1/billing/status
  API-->>FE: 200 (subscription_state=premium_active, last_transaction_status=success)
  FE->>API: GET /api/v1/auth/me
  API-->>FE: 200 (is_premium=true, premium_expired_at=...)
  FE-->>U: Render upgrade_success (badge Pro aktif)
```

## Failure Path

```mermaid
sequenceDiagram
  participant U as User
  participant FE as Frontend (Web)
  participant API as Bisakerja API
  participant M as Mayar

  U->>FE: Konfirmasi upgrade
  FE->>API: POST /api/v1/billing/checkout-session

  alt Checkout request invalid
    API-->>FE: 400 BAD_REQUEST
    FE-->>U: Render checkout_error_validation
  else Sudah premium aktif
    API-->>FE: 409 CONFLICT
    FE-->>U: Render already_premium
  else Upstream Mayar bermasalah/rate limited
    API-->>FE: 502/503
    FE-->>U: Render checkout_error_retry + CTA coba lagi
  end

  U->>M: Pembayaran belum selesai
  M->>API: POST /api/v1/webhook/mayar (event=payment.reminder)
  API-->>M: 200 Webhook processed
  FE->>API: GET /api/v1/billing/status
  API-->>FE: 200 (subscription_state=pending_payment)
  FE-->>U: Render upgrade_pending (menunggu pembayaran)

  U->>FE: Kembali lagi setelah pembayaran gagal
  FE->>API: GET /api/v1/billing/status
  API-->>FE: 200 (subscription_state=free/premium_expired, last_transaction_status=failed)
  FE-->>U: Render upgrade_reoffer + CTA checkout baru
```

## UI State Transitions

| Current State | Trigger | Next State | Catatan UI |
|---|---|---|---|
| `billing_ready` | User klik "Lanjut bayar" | `checkout_creating` | Disable tombol submit selama request |
| `checkout_creating` | `POST /api/v1/billing/checkout-session` sukses (`201`) | `checkout_redirecting` | Simpan `transaction_id`, lalu redirect ke `checkout_url` |
| `checkout_creating` | `400 BAD_REQUEST` | `checkout_error_validation` | Tampilkan validasi plan/redirect URL |
| `checkout_creating` | `409 CONFLICT` | `already_premium` | Tampilkan status premium aktif, tanpa redirect |
| `checkout_creating` | `502/503` | `checkout_error_retry` | Tampilkan retry action |
| `checkout_redirecting` | User kembali dari gateway | `payment_verifying` | Mulai cek status via `/api/v1/billing/status` |
| `payment_verifying` | `subscription_state=premium_active` | `upgrade_success` | Refresh `/api/v1/auth/me` untuk sinkronisasi UI global |
| `payment_verifying` | `subscription_state=pending_payment` | `upgrade_pending` | Tampilkan instruksi selesaikan pembayaran |
| `payment_verifying` | `subscription_state=free` atau `premium_expired` + `last_transaction_status=failed` | `upgrade_reoffer` | Tampilkan alasan gagal + CTA checkout baru |
| `payment_verifying` | timeout/5xx berulang | `upgrade_pending_manual_check` | Sediakan refresh manual + kontak support |

## Backend/API Touchpoints

- `POST /api/v1/billing/checkout-session` — create checkout session ([Billing API](../../api/billing.md)).
- `GET /api/v1/billing/status` — baca state subscription (`free`, `pending_payment`, `premium_active`, `premium_expired`) ([Billing API](../../api/billing.md)).
- `GET /api/v1/auth/me` — sinkronisasi flag `is_premium` di UI ([Auth API](../../api/auth.md)).
- `POST /api/v1/webhook/mayar` — asynchronous event processing (`payment.received`, `payment.reminder`) ([Webhooks API](../../api/webhooks.md)).
- Aturan bisnis subscription dan idempotency webhook: [Subscription & Billing](../../features/subscription-billing.md), [Mayar Integration](../../architecture/mayar-integration.md).

## Acceptance Criteria Flow

- Frontend tidak pernah menandai sukses premium hanya dari redirect/callback URL; verifikasi wajib lewat `GET /api/v1/billing/status`.
- `subscription_state` menjadi penentu final entitlement UI meskipun `last_transaction_status` masih pending/reminder.
- Error checkout (`400`, `409`, `502/503`) selalu memiliki next action yang jelas (perbaiki input, lihat status, retry).
- Pada status pembayaran gagal (`last_transaction_status=failed`), UI menampilkan re-offer checkout tanpa dead-end.
