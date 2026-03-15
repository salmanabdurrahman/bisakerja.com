# Premium Upgrade

## Objective

Mengonversi user dari `free`/`premium_expired` menjadi `premium_active` melalui alur checkout yang jelas, aman, dan transparan terhadap status pembayaran.

## Scope MVP

- Menampilkan status langganan aktif user.
- Menampilkan CTA upgrade untuk user `free` dan `premium_expired`.
- Membuat checkout session dan redirect user ke `checkout_url`.
- Menangani state `pending_payment` setelah checkout dibuat.
- Refresh status langganan setelah user kembali dari halaman pembayaran.

## Out of Scope

- Cicilan, multi-plan kompleks, refund/chargeback.
- Riwayat invoice detail dengan drill-down penuh.
- Webhook handling di frontend (tetap tanggung jawab backend).

## UI Behavior Rules

- Label state subscription harus mengikuti nilai canonical:
  - `free`
  - `pending_payment`
  - `premium_active`
  - `premium_expired`
- Tombol upgrade:
  - aktif untuk `free` dan `premium_expired`,
  - nonaktif untuk `premium_active`,
  - berubah menjadi "Lanjutkan Pembayaran" saat `pending_payment` dan checkout URL masih valid.
- Setelah `POST /api/v1/billing/checkout-session` sukses, frontend wajib:
  1. menyimpan `checkout_url` + `expired_at`,
  2. redirect user ke provider checkout.
- Frontend boleh mengirim `coupon_code` opsional saat membuat checkout; jika backend mengembalikan `INVALID_COUPON_CODE`, UI wajib menampilkan error yang spesifik.
- Halaman callback/success tidak boleh langsung menganggap pembayaran sukses; wajib konfirmasi via `GET /api/v1/billing/status`.
- Entitlement premium **hanya** diputuskan oleh `subscription_state`; `last_transaction_status` dipakai sebagai informasi UX tambahan.

## Subscription & Payment State Mapping

| Sumber backend | Nilai | Dampak UI |
|---|---|---|
| `billing/status.subscription_state` | `free` | CTA upgrade aktif |
| `billing/status.subscription_state` | `pending_payment` | CTA "Lanjutkan Pembayaran" + instruksi selesaikan pembayaran |
| `billing/status.subscription_state` | `premium_active` | CTA upgrade nonaktif, tampil badge premium aktif |
| `billing/status.subscription_state` | `premium_expired` | CTA re-upgrade aktif |
| `billing/status.last_transaction_status` | `success` | Tampilkan info pembayaran berhasil terakhir |
| `billing/status.last_transaction_status` | `reminder` / `pending` | Tampilkan info pembayaran belum selesai |
| `billing/status.last_transaction_status` | `failed` | Tampilkan info gagal + CTA buat checkout baru |

## Edge Cases

- **Double click CTA upgrade**: hanya satu request checkout yang boleh diproses.
- **`409 CONFLICT` saat checkout**: sinkron ulang status; jika ternyata `premium_active`, update UI tanpa error fatal.
- **Checkout URL kadaluarsa** (`expired_at` terlewati): tampilkan CTA untuk membuat session baru.
- **Webhook delay**: jika user sudah bayar tapi status masih `pending_payment`, tampilkan status "sedang diproses" + manual refresh.
- **User meninggalkan flow lalu kembali**: status terbaru selalu ditarik dari backend sebelum menampilkan CTA.
- **Status pembayaran terakhir `failed`**: jangan ubah entitlement; tawarkan checkout ulang.
- **`premium_active` tapi histori transaksi belum termuat**: badge premium tetap tampil, histori dimuat deferred.

## API Dependencies

| Endpoint | Tujuan di Frontend | Field minimum yang dikonsumsi | Referensi |
|---|---|---|---|
| `POST /api/v1/billing/checkout-session` | Membuat sesi checkout dan mendapat `checkout_url`. | request: `plan_code`, `customer_mobile`, `redirect_url`, `coupon_code?`; response: `data.checkout_url`, `data.transaction_id`, `data.expired_at`, `data.subscription_state`, `data.transaction_status`, `data.original_amount`, `data.discount_amount`, `data.final_amount`, `data.coupon_code?` | [billing.md](../../api/billing.md) |
| `GET /api/v1/billing/status` | Mendapatkan state subscription canonical. | `data.subscription_state`, `data.last_transaction_status`, `data.premium_expired_at` | [billing.md](../../api/billing.md) |
| `GET /api/v1/billing/transactions` | Menampilkan ringkasan histori pembayaran (opsional MVP ringan). | `data[].status`, `data[].amount`, `data[].created_at` | [billing.md](../../api/billing.md) |
| `GET /api/v1/auth/me` | Sinkron badge premium di navbar/profile. | `data.is_premium`, `data.premium_expired_at` | [auth.md](../../api/auth.md) |

## Loading / Error / Empty States

- **Loading**
  - Saat fetch status billing: tampilkan status skeleton.
  - Saat create checkout: tombol upgrade tampil loading + disabled.
- **Error**
- `400 BAD_REQUEST`: tampilkan pesan validasi spesifik (`INVALID_PLAN_CODE`/`INVALID_CUSTOMER_MOBILE`/`INVALID_COUPON_CODE`/`INVALID_REDIRECT_URL`).
    - Untuk `INVALID_REDIRECT_URL`, arahkan user mengecek `BILLING_REDIRECT_ALLOWLIST`; local dev boleh `http` hanya untuk `localhost`/`127.0.0.1`/`::1`.
  - `401 UNAUTHORIZED`: minta login ulang.
  - `409 CONFLICT`: refresh status dan tampilkan pesan informatif.
  - `429 TOO_MANY_REQUESTS`: tampilkan instruksi tunggu singkat (~10 detik) lalu retry/continue pending checkout.
  - `502`/`503`: tampilkan error payment provider sementara + retry.
- **Empty**
  - Jika tidak ada histori transaksi: tampilkan empty state "Belum ada transaksi".

## Acceptance Criteria

- User `free` dapat memulai checkout sampai redirect ke provider.
- State `pending_payment` terlihat jelas setelah checkout dibuat.
- State `premium_active` hanya tampil setelah verifikasi status dari backend.
- Error pembayaran tidak menyebabkan UI stuck; user selalu punya aksi lanjut (retry/refresh).
- Terminologi state di seluruh komponen upgrade konsisten dan canonical.
- `last_transaction_status` tidak pernah mengoverride `subscription_state` untuk entitlement.

## Output Implementasi Minimum

- Route: `/pricing`, `/account/subscription`, callback success route (mis. `/billing/success`).
- Komponen: `SubscriptionStatusCard`, `UpgradeCTA`, `BillingHistoryList`.
- Service calls typed: `createCheckoutSession`, `getBillingStatus`, `getBillingTransactions`.
- Test minimum:
  - integration test checkout success -> redirect provider,
  - integration test callback verify -> `upgrade_success` vs `upgrade_pending`,
  - component test error branch (`409`, `502/503`, `failed` transaction status).

## Related Specs

- [Auth & Session](./auth-session.md)
- [Preferences & Notifications](./preferences-notifications.md)
- [Profile & Account](./profile-account.md)
