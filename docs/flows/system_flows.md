# System Flows (Ringkas)

Dokumen flow sudah dipecah agar lebih detail.

## Daftar Flow Lengkap

1. [Auth & Subscription Flow](./auth-subscription-flow.md)
2. [Scraping & Matching Flow](./scraping-matching-flow.md)
3. [Search Serving Flow](./search-serving-flow.md)
4. [Admin Operations Flow](./admin-ops-flow.md)

## Ringkasan End-to-End

1. Scraper mengambil lowongan dari portal.
2. Job baru disimpan ke database dengan deduplikasi `source+original_job_id`.
3. Event job baru dipublish ke queue.
4. Matcher mengecek user `premium_active` + preference.
5. Notifier mengirim email dan update status.
6. User mencari/membaca lowongan via API jobs.
7. Upgrade premium diproses via checkout + webhook idempotent.
