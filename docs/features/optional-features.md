# Optional Features (Post-MVP)

## Tujuan

Dokumen ini merinci fitur tambahan yang tidak wajib untuk MVP, tetapi bernilai tinggi untuk retensi user, monetisasi, dan efisiensi operasional.

## Prinsip Prioritas

1. Dampak langsung ke user aktif dan conversion premium.
2. Reuse komponen yang sudah ada (jobs, preference, notification, billing).
3. Tidak mengganggu stabilitas pipeline scraper dan search.

## Status Implementasi Backend (Phase 3)

- ✅ Saved Search & Alert Rules
- ✅ Notification Frequency Control (Instant/Digest)
- ✅ In-App Notification Center
- ✅ Company Watchlist
- 🟡 Phase 4 increment started: coupon-enabled checkout (`coupon_code` -> Mayar `/coupon/validate`)

Catatan implementasi saat ini:

- backend sudah menyediakan kontrak CRUD/read model dan preference control untuk fitur growth,
- delivery terjadwal berbasis saved-search/watchlist masih mengikuti orchestrasi matcher/worker yang ada (hook scheduler batch lanjutan bisa dilanjutkan di phase berikutnya).
- advanced billing baru masuk increment pertama (coupon validation + discounted invoice amount); installment/payment request/qrcode belum diaktifkan.

## Daftar Fitur Opsional

### 1) Saved Search & Alert Rules

- **Nilai bisnis**: user bisa menyimpan query favorit dan menerima alert lebih presisi.
- **API tambahan**:
  - `POST /api/v1/saved-searches`
  - `GET /api/v1/saved-searches`
  - `DELETE /api/v1/saved-searches/:id`
- **Dampak DB**: tabel `saved_searches` (`user_id`, `query`, `filters`, `frequency`, `is_active`).

### 2) Notification Frequency Control (Instant/Digest)

- **Nilai bisnis**: mengurangi noise notifikasi dan meningkatkan open rate.
- **API tambahan**:
  - `PUT /api/v1/preferences/notification`
- **Dampak DB**: tambah kolom pada `user_preferences` seperti `alert_mode` dan `digest_hour`.

### 3) In-App Notification Center

- **Nilai bisnis**: histori notifikasi bisa dibaca ulang di aplikasi.
- **API tambahan**:
  - `GET /api/v1/notifications`
  - `PATCH /api/v1/notifications/:id/read`
- **Dampak DB**: perlu kolom `read_at` pada `notifications`.

### 4) Company Watchlist & Hiring Signals

- **Nilai bisnis**: user dapat follow perusahaan dan mendapat update lowongan baru dari perusahaan tersebut.
- **API tambahan**:
  - `POST /api/v1/watchlist/companies`
  - `GET /api/v1/watchlist/companies`
  - `DELETE /api/v1/watchlist/companies/:company_slug`
- **Dampak DB**: tabel `company_watchlists`.

### 5) Salary Normalization Service

- **Nilai bisnis**: hasil pencarian salary lebih konsisten lintas source.
- **Perubahan teknis**:
  - tambah normalizer worker.
  - kolom `salary_min`, `salary_max`, `salary_currency`, `salary_period` pada `jobs`.

### 6) Application Tracker & Bookmark

- **Nilai bisnis**: meningkatkan stickiness produk karena user melacak status lamaran.
- **API tambahan**:
  - `POST /api/v1/jobs/:id/bookmark`
  - `DELETE /api/v1/jobs/:id/bookmark`
  - `POST /api/v1/applications`
  - `PATCH /api/v1/applications/:id/status`
- **Dampak DB**: tabel `bookmarks` dan `applications`.

### 7) Recommendation Layer (Semantic Matching)

- **Nilai bisnis**: relevansi rekomendasi meningkat untuk user premium.
- **Perubahan teknis**:
  - embedding pipeline (batch).
  - vector index terpisah (service tambahan, post-MVP lanjut).

### 8) Advanced Billing menggunakan Mayar

- **Nilai bisnis**: fleksibilitas pembayaran untuk meningkatkan konversi checkout.
- **Endpoint Mayar yang bisa dipakai**:
  - `GET /hl/v1/coupon/validate`
  - `POST /hl/v1/installment/create`
  - `POST /hl/v1/payment/create`
  - `POST /hl/v1/qrcode/create`
- **Dampak Bisakerja**:
  - endpoint internal billing perlu mode `invoice`, `payment_request`, dan `installment`.
  - tabel transaksi perlu menyimpan `payment_type` dan metadata tambahan.

### 9) B2B Team Subscription (Opsional Lanjut)

- **Nilai bisnis**: membuka monetisasi per perusahaan/tim.
- **API tambahan**:
  - `POST /api/v1/orgs`
  - `POST /api/v1/orgs/:id/invites`
  - `GET /api/v1/orgs/:id/billing`
- **Dampak DB**: tabel `organizations`, `organization_members`, `organization_subscriptions`.

## Exit Criteria Fitur Opsional

- Setiap fitur opsional harus punya:
  - kontrak API terdokumentasi,
  - dampak schema DB terdokumentasi,
  - flow end-to-end terdokumentasi,
  - metrik keberhasilan minimum (adopsi, error rate, latency).
