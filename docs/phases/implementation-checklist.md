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

Berdasarkan struktur repository saat ini, implementasi **Phase 0** sudah dimulai.

- Source folder aplikasi sudah tersedia: `apps/api` dan `apps/web`.
- Baseline check lokal backend + frontend sudah berjalan dan tervalidasi ulang.
- Evidence CI sudah tersedia melalui eksekusi workflow lokal menggunakan `act` untuk `ci-api` dan `ci-web`.
- Implementasi saat ini: **✅ Phase 0 foundation baseline complete** (dengan catatan hardening lanjutan tetap direkomendasikan).

## 4) Checklist per Phase & Iteration

## Phase 0 - Foundation & Baseline

| Item                                           | Status | Code Evidence                                                                                                                                                             | Test Evidence                                                                                                                                                                                   | CI Evidence                                                                                     | Docs/ADR Evidence                                                                                    |
| ---------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `apps/api` skeleton (`cmd`, `internal`, `pkg`) | ✅     | `apps/api/cmd/*`, `apps/api/internal/*`, `apps/api/pkg/*`, `apps/api/Makefile`                                                                                            | `make -C apps/api test` + `apps/api/internal/platform/config/config_test.go`, `apps/api/internal/adapter/http/handler/health_handler_test.go`, `apps/api/test/integration/health_route_test.go` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded)           | `docs/architecture/monorepo-project-layout.md`                                                       |
| `apps/web` baseline terintegrasi ke backend    | ✅     | `apps/web/src/app/*`, `apps/web/src/services/*`, `apps/web/src/lib/*`, `package.json`, `pnpm-workspace.yaml`                                                              | `pnpm --filter web test`, `pnpm --filter web test:coverage`, `apps/web/tests/unit/env.test.ts`, `apps/web/tests/components/home-page.test.tsx`                                                  | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded)      | `docs/frontend/phases/implementation-roadmap.md`, `docs/frontend/phases/implementation-checklist.md` |
| konfigurasi environment/bootstrap              | ✅     | `apps/api/.env.example`, `apps/api/internal/platform/config/config.go`, `apps/web/.env.example`, `apps/web/src/lib/config/env.ts`                                         | `apps/api/internal/platform/config/config_test.go`, `apps/web/tests/unit/env.test.ts`                                                                                                           | `ci-api` + `ci-web` job pass via `act` local workflow run                                       | `docs/phases/implementation-kickoff.md`                                                              |
| migrasi database awal                          | ✅     | `apps/api/migrations/000001_init.up.sql`, `apps/api/migrations/000001_init.down.sql`, `apps/api/cmd/migrate/main.go`, `apps/api/internal/platform/migration/validator.go` | `make -C apps/api check-migrations`, `apps/api/internal/platform/migration/validator_test.go`                                                                                                   | `ci-api` job pass via `act` (`Validate migrations` step success)                                | `docs/architecture/database.md`                                                                      |
| healthcheck API/worker                         | ✅     | `apps/api/internal/adapter/http/handler/health_handler.go`, `apps/api/internal/platform/worker/runner.go`, `apps/api/internal/adapter/http/router/router.go`              | `make -C apps/api test`, `go run ./apps/api/cmd/scraper -healthcheck`, `go run ./apps/api/cmd/notifier -healthcheck`, `go run ./apps/api/cmd/billing-worker -healthcheck`                       | `ci-api` job pass via `act` (`Test`, `Build`, `gosec` success; `govulncheck` non-blocking note) | `docs/phases/implementation-roadmap.md`                                                              |
| standards hub + gate policy terdefinisi        | 📝     | -                                                                                                                                                                         | -                                                                                                                                                                                               | -                                                                                               | `docs/standards/README.md`, `docs/standards/ci-quality-gates.md`                                     |

## Phase 1 - MVP Core Delivery

### Iteration 1.1 - Aggregation & Search

| Item                                   | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| -------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| scraper worker source utama            | ⬜     | -             | -             | -           | -                 |
| source adapters (`glints/kalibrr/jobstreet`) | ⬜     | -             | -             | -           | `docs/features/source-scraping-playbook.md`, `docs/architecture/scraper-source-adapters.md` |
| token provider untuk source auth-required | ⬜     | -             | -             | -           | `docs/features/source-scraping-playbook.md`, `docs/flows/scraping-matching-flow.md` |
| deduplikasi jobs                       | ⬜     | -             | -             | -           | -                 |
| endpoint `GET /jobs` + `GET /jobs/:id` | ⬜     | -             | -             | -           | -                 |
| telemetry scrape run per source        | ⬜     | -             | -             | -           | `docs/architecture/scraper-source-adapters.md` |
| regression test search                 | ⬜     | -             | -             | -           | -                 |

### Iteration 1.2 - Identity & Preferences

| Item                                          | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| endpoint auth (`register/login/refresh/me`)   | ⬜     | -             | -             | -           | -                 |
| endpoint preferences (`GET/PUT /preferences`) | ⬜     | -             | -             | -           | -                 |
| middleware auth + role dasar                  | ⬜     | -             | -             | -           | -                 |
| test auth flow                                | ⬜     | -             | -             | -           | -                 |

### Iteration 1.3 - Notification MVP

| Item                        | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| matcher worker rule-based   | ⬜     | -             | -             | -           | -                 |
| email notifier worker       | ⬜     | -             | -             | -           | -                 |
| status notifikasi tersimpan | ⬜     | -             | -             | -           | -                 |
| test notification flow      | ⬜     | -             | -             | -           | -                 |

## Phase 2 - Billing Hardening (Mayar)

### Iteration 2.1 - Checkout Orchestration

| Item                                               | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| -------------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| endpoint `POST /billing/checkout-session`          | ⬜     | -             | -             | -           | -                 |
| adapter Mayar `customer/create` + `invoice/create` | ⬜     | -             | -             | -           | -                 |
| transaksi pending tersimpan                        | ⬜     | -             | -             | -           | -                 |

### Iteration 2.2 - Webhook Reliability

| Item                             | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| -------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| endpoint `POST /webhook/mayar`   | ⬜     | -             | -             | -           | -                 |
| idempotency key transaksi        | ⬜     | -             | -             | -           | -                 |
| audit tabel `webhook_deliveries` | ⬜     | -             | -             | -           | -                 |
| test duplicate webhook           | ⬜     | -             | -             | -           | -                 |

### Iteration 2.3 - Reconciliation & Recovery

| Item                          | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| ----------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| endpoint/status billing user  | ⬜     | -             | -             | -           | -                 |
| rekonsiliasi ke Mayar         | ⬜     | -             | -             | -           | -                 |
| retry workflow error upstream | ⬜     | -             | -             | -           | -                 |
| alert billing anomaly         | ⬜     | -             | -             | -           | -                 |

## Phase 3 - Growth Features (Optional Set A)

| Item                              | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| saved searches + scheduled alerts | ⬜     | -             | -             | -           | -                 |
| in-app notification center        | ⬜     | -             | -             | -           | -                 |
| company watchlist                 | ⬜     | -             | -             | -           | -                 |
| frequency/digest control          | ⬜     | -             | -             | -           | -                 |

## Phase 4 - Advanced Monetization & Intelligence (Optional Set B)

| Item                                         | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| -------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| advanced billing (coupon/installment/qrcode) | ⬜     | -             | -             | -           | -                 |
| salary normalization lanjutan                | ⬜     | -             | -             | -           | -                 |
| semantic recommendation                      | ⬜     | -             | -             | -           | -                 |
| B2B team subscription                        | ⬜     | -             | -             | -           | -                 |
| ADR untuk keputusan besar phase ini          | ⬜     | -             | -             | -           | -                 |

## 5) Cara Pakai

1. Ubah status item sesuai progres (`⬜` → `🟡` → `✅`).
2. Isi seluruh kolom evidence saat status `✅`.
3. Validasi gate mengacu ke `docs/standards/ci-quality-gates.md`.
4. Update checklist ini di PR yang sama dengan implementasi.
