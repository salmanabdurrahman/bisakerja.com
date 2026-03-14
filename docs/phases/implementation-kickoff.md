# Implementation Kickoff Guide (Backend + Frontend)

Dokumen ini menjadi panduan operasional terakhir sebelum coding dimulai bertahap.

## 1) Tujuan

1. Menyamakan kontrak FE-BE sebelum implementasi.
2. Memastikan setiap iterasi punya output terukur dan evidence jelas.
3. Mengurangi rework akibat mismatch requirement, API, dan flow.

## 2) Prasyarat Kickoff

Semua poin berikut harus siap:

- PRD backend: [`../prd/bisakerja-api-prd.md`](../prd/bisakerja-api-prd.md)
- PRD frontend: [`../prd/bisakerja-frontend-prd.md`](../prd/bisakerja-frontend-prd.md)
- Source scraping playbook: [`../features/source-scraping-playbook.md`](../features/source-scraping-playbook.md)
- Standards hub: [`../standards/README.md`](../standards/README.md)
- Roadmap backend: [`./implementation-roadmap.md`](./implementation-roadmap.md)
- Checklist backend: [`./implementation-checklist.md`](./implementation-checklist.md)
- Roadmap frontend: [`../frontend/phases/implementation-roadmap.md`](../frontend/phases/implementation-roadmap.md)
- Checklist frontend: [`../frontend/phases/implementation-checklist.md`](../frontend/phases/implementation-checklist.md)
- Traceability FE-BE: [`../frontend/traceability/frontend-backend-traceability.md`](../frontend/traceability/frontend-backend-traceability.md)

## 3) Langkah Kickoff (Wajib Berurutan)

### 3.1 Konfirmasi Canonical Contract

- Kunci enum canonical:
  - `subscription_state`: `free`, `pending_payment`, `premium_active`, `premium_expired`
  - `transactions.status`: `pending`, `reminder`, `success`, `failed`
- Kunci ownership: endpoint user-scoped selalu dari `JWT.sub`.
- Kunci source of truth entitlement: `GET /billing/status -> subscription_state`.
- Kunci preferences contract: `GET /preferences` (read/bootstrap), `PUT /preferences` (write/update).

### 3.2 Pecah Scope per Phase/Iteration

- Turunkan roadmap menjadi task implementasi yang kecil dan independen.
- Konfirmasi capability matrix source (`requires_auth`, pagination mode, source readiness) sebelum mulai Iteration 1.1.
- Tandai dependency antar task backend/frontend (khusus endpoint yang dipakai UI).
- Tetapkan evidence minimum per task: code, test, CI, docs/ADR.

### 3.3 Setup Fondasi Implementasi

- Siapkan struktur monorepo sesuai:
  [`../architecture/monorepo-project-layout.md`](../architecture/monorepo-project-layout.md)
- Siapkan baseline CI dan gate merge sesuai:
  [`../standards/ci-quality-gates.md`](../standards/ci-quality-gates.md)
- Pastikan observability minimum tersedia untuk flow kritikal:
  [`../standards/security-observability-standards.md`](../standards/security-observability-standards.md)

### 3.4 Eksekusi PR Bertahap

- Setiap PR harus punya scope kecil dan terukur.
- Setiap PR harus update checklist phase yang terdampak.
- Status checklist `✅` hanya boleh dipakai jika evidence lengkap.

### 3.5 Validasi Sebelum Merge

- Jalankan gate sesuai scope (G1-G8).
- Verifikasi tidak ada mismatch FE-BE untuk endpoint/field/status.
- Jika ada perubahan kontrak atau arsitektur, update docs terkait di PR yang sama.

## 4) Urutan Implementasi yang Direkomendasikan

1. **Phase 0 Backend Foundation**: fondasi service, schema, auth envelope, observability dasar.
2. **Phase 0 Frontend Foundation**: bootstrap `apps/web`, data/state boundary, baseline test/build.
3. **Phase 1 Backend Core**: jobs + auth + preferences + notification engine.
4. **Phase 1 Frontend MVP**: auth, discovery, preferences, upgrade flow.
5. **Phase 2 Billing Hardening**: webhook reliability, reconciliation, recovery.
6. **Growth/Advanced**: hanya setelah core stabil dan gate quality konsisten hijau.

## 5) Definition of Kickoff Complete

Kickoff dinyatakan selesai jika:

- roadmap + checklist backend/frontend siap dipakai eksekusi,
- canonical contract sudah disepakati lintas tim,
- gate kualitas sudah enforceable di CI,
- owner task awal Phase 0 sudah terdefinisi.
