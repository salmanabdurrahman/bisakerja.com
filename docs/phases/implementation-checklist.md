# Implementation Checklist (Phase Audit)

Checklist ini dipakai untuk memastikan status implementasi berbasis evidence, bukan asumsi.

## 1) Definisi Status

- `✅ Implemented`: kode production siap pakai + test + CI + docs/ADR evidence lengkap.
- `🟡 Partial`: implementasi sebagian; ada gap pada wiring, test, CI, atau docs.
- `⬜ Not Implemented`: belum ada implementasi kode.
- `📝 Documented`: requirement/arsitektur tersedia, kode belum ada.
- `⛔ Blocked`: tidak bisa lanjut karena blocker eksternal/internal yang jelas.

## 2) Aturan Evidence (Wajib untuk `✅`)

Setiap item `✅` harus mengisi seluruh kolom evidence:

1. **Code Evidence**: path kode terkait (repo-relative).
2. **Test Evidence**: path test + jenis test/suite.
3. **CI Evidence**: nama workflow/job yang pass.
4. **Docs/ADR Evidence**: path dokumen/ADR yang diperbarui, atau `N/A` bila tidak relevan.

Jika salah satu kolom kosong, status item harus `🟡`, bukan `✅`.

## 3) Snapshot Audit Saat Ini

Berdasarkan struktur repository saat ini, konten masih didominasi dokumentasi (`docs/`) dan belum ditemukan source folder aplikasi (`apps/api`, `apps/web`, `cmd`, `internal`, dll).

- Implementasi kode backend/frontend: **⬜ Not Implemented**
- Kesiapan dokumen standar + phase: **📝 Documented**

## 4) Checklist per Phase & Iteration

## Phase 0 - Foundation & Baseline

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
|---|---|---|---|---|---|
| `apps/api` skeleton (`cmd`, `internal`, `pkg`) | ⬜ | - | - | - | - |
| `apps/web` baseline terintegrasi ke backend | ⬜ | - | - | - | - |
| konfigurasi environment/bootstrap | ⬜ | - | - | - | - |
| migrasi database awal | ⬜ | - | - | - | - |
| healthcheck API/worker | ⬜ | - | - | - | - |
| standards hub + gate policy terdefinisi | 📝 | - | - | - | `docs/standards/README.md`, `docs/standards/ci-quality-gates.md` |

## Phase 1 - MVP Core Delivery

### Iteration 1.1 - Aggregation & Search

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
|---|---|---|---|---|---|
| scraper worker source utama | ⬜ | - | - | - | - |
| deduplikasi jobs | ⬜ | - | - | - | - |
| endpoint `GET /jobs` + `GET /jobs/:id` | ⬜ | - | - | - | - |
| regression test search | ⬜ | - | - | - | - |

### Iteration 1.2 - Identity & Preferences

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
|---|---|---|---|---|---|
| endpoint auth (`register/login/refresh/me`) | ⬜ | - | - | - | - |
| endpoint preferences (`GET/PUT /preferences`) | ⬜ | - | - | - | - |
| middleware auth + role dasar | ⬜ | - | - | - | - |
| test auth flow | ⬜ | - | - | - | - |

### Iteration 1.3 - Notification MVP

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
|---|---|---|---|---|---|
| matcher worker rule-based | ⬜ | - | - | - | - |
| email notifier worker | ⬜ | - | - | - | - |
| status notifikasi tersimpan | ⬜ | - | - | - | - |
| test notification flow | ⬜ | - | - | - | - |

## Phase 2 - Billing Hardening (Mayar)

### Iteration 2.1 - Checkout Orchestration

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
|---|---|---|---|---|---|
| endpoint `POST /billing/checkout-session` | ⬜ | - | - | - | - |
| adapter Mayar `customer/create` + `invoice/create` | ⬜ | - | - | - | - |
| transaksi pending tersimpan | ⬜ | - | - | - | - |

### Iteration 2.2 - Webhook Reliability

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
|---|---|---|---|---|---|
| endpoint `POST /webhook/mayar` | ⬜ | - | - | - | - |
| idempotency key transaksi | ⬜ | - | - | - | - |
| audit tabel `webhook_deliveries` | ⬜ | - | - | - | - |
| test duplicate webhook | ⬜ | - | - | - | - |

### Iteration 2.3 - Reconciliation & Recovery

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
|---|---|---|---|---|---|
| endpoint/status billing user | ⬜ | - | - | - | - |
| rekonsiliasi ke Mayar | ⬜ | - | - | - | - |
| retry workflow error upstream | ⬜ | - | - | - | - |
| alert billing anomaly | ⬜ | - | - | - | - |

## Phase 3 - Growth Features (Optional Set A)

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
|---|---|---|---|---|---|
| saved searches + scheduled alerts | ⬜ | - | - | - | - |
| in-app notification center | ⬜ | - | - | - | - |
| company watchlist | ⬜ | - | - | - | - |
| frequency/digest control | ⬜ | - | - | - | - |

## Phase 4 - Advanced Monetization & Intelligence (Optional Set B)

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
|---|---|---|---|---|---|
| advanced billing (coupon/installment/qrcode) | ⬜ | - | - | - | - |
| salary normalization lanjutan | ⬜ | - | - | - | - |
| semantic recommendation | ⬜ | - | - | - | - |
| B2B team subscription | ⬜ | - | - | - | - |
| ADR untuk keputusan besar phase ini | ⬜ | - | - | - | - |

## 5) Cara Pakai

1. Ubah status item sesuai progres (`⬜` → `🟡` → `✅`).
2. Isi seluruh kolom evidence saat status `✅`.
3. Validasi gate mengacu ke `docs/standards/ci-quality-gates.md`.
4. Update checklist ini di PR yang sama dengan implementasi.
