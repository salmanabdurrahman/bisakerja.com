# Smart Notification Engine

## Tujuan

Mengirim notifikasi lowongan baru yang relevan ke user premium secepat mungkin dengan biaya komputasi terkontrol.

## Prioritas

**MVP** (email), **Post-MVP** (WhatsApp)

## Scope Fitur

### In Scope (MVP)

- Trigger saat ada job baru yang valid tersimpan.
- Matching berbasis keyword, lokasi, tipe kerja, dan salary minimum.
- Pengiriman notifikasi via email.
- Penyimpanan status notifikasi (`pending`, `sent`, `failed`).

### Out of Scope (MVP)

- AI semantic matching.
- Pengiriman multi-channel paralel.
- Digest harian/mingguan sebagai channel pengiriman penuh (phase 3 saat ini sudah menambahkan kontrol mode/frekuensi di preference).

## Aturan Bisnis

1. Hanya user dengan `subscription_state = premium_active` yang menerima notifikasi.
2. Satu notifikasi dibuat per pasangan `user_id + job_id + channel`.
3. Kegagalan kirim email tidak menghapus data notifikasi; status menjadi `failed`.
4. Notifikasi tidak dikirim ulang untuk item yang sudah `sent` kecuali ada mekanisme replay admin.

## Algoritma Matching (MVP)

1. Konsumsi event `JobCreated`.
2. Ambil job detail.
3. Ambil daftar user premium aktif + preference.
4. Cek:
   - keyword match (judul/deskripsi),
   - lokasi cocok,
   - tipe kerja cocok,
   - salary memenuhi minimum (jika tersedia).
5. Buat record notifikasi.
6. Kirim email.
7. Update status.

## Dependensi API

Fitur ini berjalan asynchronous, tetapi bergantung pada endpoint:

- `PUT /api/v1/preferences`
- `POST /api/v1/webhook/mayar`
- `GET /api/v1/billing/status`

Lihat detail: [`../api/preferences.md`](../api/preferences.md), [`../api/webhooks.md`](../api/webhooks.md), [`../api/billing.md`](../api/billing.md).

## Dampak Data

- Tabel utama:
  - `user_preferences`
  - `notifications`
  - `users`
  - `jobs`

Lihat detail: [`../architecture/database.md`](../architecture/database.md).

## Monitoring Minimum

- Jumlah event job baru.
- Jumlah notifikasi terkirim.
- Jumlah notifikasi gagal.
- Latensi dari insert job ke send email.

## NFR Minimum

| Metric | Target |
|---|---|
| Match-to-email latency p95 | < 60 detik |
| Delivery success ratio | >= 98% / hari |
| Duplicate notification ratio | 0 untuk pasangan `user_id+job_id+channel` |

## Acceptance Criteria

- User premium aktif menerima notifikasi untuk lowongan yang match.
- User non-premium tidak menerima notifikasi.
- Status notifikasi akurat (`sent`/`failed`).
- Tidak ada duplicate notification untuk user-job-channel yang sama.
