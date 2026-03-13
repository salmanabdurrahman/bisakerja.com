# Fitur Produk Bisakerja API

Dokumen ini adalah katalog fitur backend Bisakerja.

## Daftar Fitur Inti

1. [Job Aggregation Engine](./job-aggregation.md)
2. [Job Search Engine](./job-search.md)
3. [Smart Notification Engine](./smart-notification.md)
4. [Subscription & Billing](./subscription-billing.md)
5. [Admin Operations](./admin-operations.md)
6. [Optional Features (Post-MVP)](./optional-features.md)

## Cara Membaca

- Mulai dari dokumen fitur untuk memahami requirement bisnis dan aturan utama.
- Lanjutkan ke `docs/api/` untuk kontrak endpoint.
- Lanjutkan ke `docs/architecture/` untuk detail komponen teknis.
- Lanjutkan ke `docs/flows/` untuk urutan proses end-to-end.

## Definisi Status Prioritas

- **MVP**: Wajib ada di rilis awal.
- **Post-MVP**: Direncanakan setelah MVP stabil.

## Canonical Enum Referensi

- `subscription_state`: `free`, `pending_payment`, `premium_active`, `premium_expired`.
- `transactions.status`: `pending`, `reminder`, `success`, `failed`.

Rujukan utama: [`../api/README.md`](../api/README.md).
