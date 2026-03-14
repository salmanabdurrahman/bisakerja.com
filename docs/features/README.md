# Fitur Produk Bisakerja API

Dokumen ini adalah katalog fitur backend Bisakerja.

## Daftar Fitur Inti

1. [Job Aggregation Engine](./job-aggregation.md)
2. [Source Scraping Playbook (Glints, Kalibrr, JobStreet)](./source-scraping-playbook.md)
3. [Job Search Engine](./job-search.md)
4. [Smart Notification Engine](./smart-notification.md)
5. [Subscription & Billing](./subscription-billing.md)
6. [Admin Operations](./admin-operations.md)
7. [Optional Features (Post-MVP)](./optional-features.md)

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
