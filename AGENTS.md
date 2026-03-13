# AGENTS Playbook — Bisakerja

Dokumen ini adalah panduan operasional agar output agent konsisten, dan agar konteks kerja backend/frontend selalu lengkap sebelum implementasi.

## 1) Tujuan Utama

1. Menjaga konsistensi kualitas output lintas task.
2. Memastikan context loading cukup sebelum mengerjakan backend atau frontend.
3. Menjaga keselarasan implementasi terhadap kontrak di `docs/`.
4. Mengurangi mismatch antar tim (API, UI, data, flow, phase).

## 2) Output Contract (Wajib Konsisten)

Gunakan struktur output berikut untuk setiap task yang selesai:

1. **Outcome**: hasil utama yang dicapai.
2. **Perubahan**: file/area yang diubah + alasan.
3. **Validasi**: test/check yang dijalankan dan hasilnya.
4. **Risiko/Asumsi**: hal yang perlu dipantau.
5. **Next Step**: langkah lanjutan paling relevan.

Jika task belum selesai, laporkan:

1. status saat ini,
2. blocker spesifik,
3. langkah berikutnya yang akan dilakukan.

## 3) Context Intake Wajib Sebelum Eksekusi

### 3.1 Baseline (semua task)

- `docs/README.md`
- `docs/standards/README.md`
- `docs/prd/bisakerja-api-prd.md`
- `docs/prd/bisakerja-frontend-prd.md`
- `docs/phases/implementation-roadmap.md`
- `docs/phases/implementation-checklist.md`
- `docs/frontend/phases/implementation-roadmap.md`
- `docs/frontend/phases/implementation-checklist.md`

### 3.2 Jika Task Backend (Go API/Worker/Scraper)

Wajib baca konteks domain terkait:

- `docs/api/*.md`
- `docs/architecture/*.md`
- `docs/features/*.md`
- `docs/flows/*.md`
- `docs/standards/go-coding-standards.md`
- `docs/standards/testing-strategy.md`
- `docs/standards/ci-quality-gates.md`
- `docs/standards/security-observability-standards.md`

### 3.3 Jika Task Frontend (Next.js User App)

Wajib baca konteks domain terkait:

- `docs/frontend/features/*.md`
- `docs/frontend/architecture/*.md`
- `docs/frontend/flows/*.md`
- `docs/frontend/traceability/frontend-backend-traceability.md`
- endpoint backend yang dipakai UI di `docs/api/*.md`
- `docs/standards/nextjs-coding-standards.md`
- `docs/standards/testing-strategy.md`
- `docs/standards/ci-quality-gates.md`
- `docs/standards/security-observability-standards.md`

### 3.4 Jika Task Lintas FE-BE

Wajib sinkronkan:

- kontrak endpoint + payload (`docs/api/`),
- traceability FE-BE (`docs/frontend/traceability/...`),
- flow end-to-end (`docs/flows/` + `docs/frontend/flows/`),
- checklist phase backend + frontend.

## 4) Canonical Contract (Tidak Boleh Menyimpang)

1. Runtime API path: `/api/v1/*`; dokumen endpoint boleh menulis resource path tanpa prefix.
2. `subscription_state`: `free`, `pending_payment`, `premium_active`, `premium_expired`.
3. `transactions.status` dan `billing/status.last_transaction_status`: `pending`, `reminder`, `success`, `failed`.
4. Entitlement premium UI harus mengikuti `GET /billing/status -> subscription_state`.
5. Preferences user: `GET /preferences` (bootstrap/read) dan `PUT /preferences` (update/write).
6. Endpoint user-scoped wajib pakai ownership `JWT.sub`.

## 5) Alur Kerja Eksekusi

1. Identifikasi scope: backend, frontend, atau lintas domain.
2. Load context wajib sesuai matriks di atas.
3. Buat rencana kerja singkat dan petakan ke phase/iteration.
4. Implementasi bertahap dengan perubahan kecil dan terukur.
5. Update dokumen/checklist yang terdampak dalam PR yang sama.
6. Jalankan validasi teknis + validasi kontrak.
7. Kirim handoff menggunakan format output konsisten (lihat Bagian 2).

## 6) Validation Gate Sebelum Selesai

Task belum dianggap selesai jika salah satu poin ini belum terpenuhi:

1. Gate wajib sesuai scope (G1-G8) terpenuhi.
2. Test relevan pass (unit/integration/e2e sesuai perubahan).
3. Tidak ada mismatch kontrak FE-BE pada endpoint/field/status.
4. Untuk perubahan docs: markdown local links harus valid.
5. Checklist phase terdampak sudah diperbarui dengan evidence yang sesuai.

## 7) Definition of Done (Praktis)

Sebuah task dianggap done jika:

- perilaku yang diminta user sudah terimplementasi,
- validasi relevan sudah dijalankan dan lulus,
- dokumentasi/traceability/checklist yang terkait sudah sinkron,
- tidak ada blocker terbuka pada scope task.

## 8) Prinsip Eksekusi

- **Simplicity first**: solusi paling sederhana yang tetap robust.
- **Root cause mindset**: hindari patch sementara tanpa akar masalah.
- **Minimal blast radius**: ubah hanya area yang dibutuhkan.
- **Evidence over assumption**: klaim “selesai” harus berbasis hasil validasi.
