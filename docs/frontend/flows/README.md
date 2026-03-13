# Frontend Flows

Dokumentasi ini memetakan journey utama user dari sisi UI, termasuk transisi state dan touchpoint ke API backend.

## Dokumen

- [Discovery Flow](./discovery-flow.md) — alur user menemukan lowongan dari search sampai detail.
- [Upgrade Billing Flow](./upgrade-billing-flow.md) — alur upgrade dari `free` ke `premium_active` via checkout Mayar.

## Konvensi Istilah

- Subscription state mengikuti backend billing: `free` -> `pending_payment` -> `premium_active` -> `premium_expired`.
- Event webhook mengikuti backend webhook: `payment.received`, `payment.reminder`.
- Path API di flow ditulis sebagai runtime path frontend: `/api/v1/*` (contoh: `/api/v1/jobs`, `/api/v1/billing/checkout-session`).
- Dokumen API backend menuliskan resource path inti (`/jobs`, `/billing/checkout-session`, dst) pada base URL `/api/v1`.

## Referensi Backend

- Jobs API: [`../../api/jobs.md`](../../api/jobs.md)
- Auth API: [`../../api/auth.md`](../../api/auth.md)
- Billing API: [`../../api/billing.md`](../../api/billing.md)
- Webhooks API: [`../../api/webhooks.md`](../../api/webhooks.md)
- Subscription & Billing rules: [`../../features/subscription-billing.md`](../../features/subscription-billing.md)
