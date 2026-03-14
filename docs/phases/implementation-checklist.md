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

## 3.1) Rencana Lanjutan (Document-First)

Sebelum implementasi lanjutan dimulai, urutan kerja dikunci agar perubahan dieksekusi satu per satu dan tervalidasi.

| Milestone | Fokus | Status |
| --------- | ----- | ------ |
| M0 | Dokumentasi rencana perubahan menyeluruh | ✅ Complete |
| M1 | Migrasi fondasi persistence ke PostgreSQL | ✅ Complete |
| M2 | Comments/docstrings pass sesuai standar | ✅ Complete |
| M3 | English migration (UI + API user-facing messages) | ✅ Complete |
| M4 | Frontend redesign + growth hardening | ✅ Complete |
| M5 | Phase 4 backend execution | 🟡 In Progress |

Catatan progress M1 saat ini:

- adapter persistence PostgreSQL untuk identity/jobs/preferences/billing/notifications/growth sudah ditambahkan,
- wiring runtime `cmd/api`, `cmd/scraper`, `cmd/notifier`, `cmd/billing-worker` sudah pindah ke PostgreSQL pool + fail-fast connection,
- migration cutover `000003_phase3_persistence_cutover` sudah ditambahkan,
- persistence queue notifier sudah dipindah ke PostgreSQL (`000004_phase3_notification_queue_postgres` + adapter queue PostgreSQL),
- hardening migrasi drift schema sudah ditambahkan di `000002` dan `000003` untuk kasus histori `schema_migrations` kosong pada database yang sudah pernah dicutover,
- bootstrap runtime lokal memakai `make -C apps/api check-migrations` + `make -C apps/api migrate-up` sebelum `make -C apps/api run-api`.

Catatan progress M2 saat ini:

- doc comment untuk simbol exported Go sudah ditambahkan pada layer adapter/app/domain/platform,
- TSDoc untuk exported API frontend pada `apps/web/src/services/*` dan `apps/web/src/lib/*` sudah ditambahkan,
- validasi non-fungsional pass: `make -C apps/api lint test build check-migrations` dan root `pnpm lint test build`.

Catatan progress M3 saat ini:

- seluruh copy user-facing di `apps/web/src` sudah dimigrasikan ke English (page headings, CTA, error/success/help text),
- pesan user-facing backend notifier (`apps/api/internal/app/notification/notifier.go`) sudah dimigrasikan ke English,
- test assertion frontend yang terdampak copy migration sudah disinkronkan,
- validasi pass: `make -C apps/api lint test build` dan root `pnpm lint test build`.

Catatan progress M4 saat ini:

- redesign SaaS style telah diterapkan di halaman inti frontend (auth, jobs, account area, pricing/subscription, billing verification) dengan fondasi komponen UI reusable,
- refinement visual pass ala Paper sudah diterapkan untuk hierarchy/spacing global (hero homepage, nav/footer shell, dan card surfaces),
- observability web vitals untuk halaman kritikal sudah aktif melalui `WebVitalsObserver` + endpoint collector same-origin `/api/web-vitals`,
- e2e-like growth journey coverage sudah ditambahkan pada `apps/web/tests/e2e/growth-engagement-flow.test.tsx`,
- validasi pass: `pnpm --filter web lint test build` dan root `pnpm lint test build`.

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
| scraper worker source utama            | ✅     | `apps/api/cmd/scraper/main.go`, `apps/api/internal/app/scraper/orchestrator.go`, `apps/api/internal/platform/worker/runner.go` | `apps/api/internal/app/scraper/orchestrator_test.go`, `apps/api/internal/platform/worker/runner_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/flows/scraping-matching-flow.md`, `docs/features/job-aggregation.md` |
| source adapters (`glints/kalibrr/jobstreet`) | ✅     | `apps/api/internal/adapter/scraper/source/glints_adapter.go`, `apps/api/internal/adapter/scraper/source/kalibrr_adapter.go`, `apps/api/internal/adapter/scraper/source/jobstreet_adapter.go` | `apps/api/internal/adapter/scraper/source/jobstreet_adapter_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/features/source-scraping-playbook.md`, `docs/architecture/scraper-source-adapters.md` |
| token provider untuk source auth-required | ✅     | `apps/api/internal/adapter/scraper/token/env_provider.go`, `apps/api/cmd/scraper/main.go` | `apps/api/internal/adapter/scraper/token/env_provider_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/features/source-scraping-playbook.md`, `docs/flows/scraping-matching-flow.md` |
| deduplikasi jobs                       | ✅     | `apps/api/internal/adapter/persistence/memory/jobs_repository.go`, `apps/api/internal/domain/job/job.go`, `apps/api/migrations/000002_phase11_scrape_runs.up.sql` | `apps/api/internal/adapter/persistence/memory/jobs_repository_test.go`, `make -C apps/api test`, `make -C apps/api check-migrations` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/features/job-aggregation.md`, `docs/architecture/database.md` |
| endpoint `GET /jobs` + `GET /jobs/:id` | ✅     | `apps/api/internal/adapter/http/handler/jobs_handler.go`, `apps/api/internal/adapter/http/router/router.go`, `apps/api/cmd/api/main.go` | `apps/api/internal/adapter/http/handler/jobs_handler_test.go`, `apps/api/test/integration/jobs_route_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/jobs.md`, `docs/flows/search-serving-flow.md` |
| telemetry scrape run per source        | ✅     | `apps/api/internal/domain/job/job.go` (`ScrapeRun`), `apps/api/internal/app/scraper/orchestrator.go`, `apps/api/migrations/000002_phase11_scrape_runs.up.sql` | `apps/api/internal/app/scraper/orchestrator_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/architecture/scraper-source-adapters.md`, `docs/flows/scraping-matching-flow.md` |
| regression test search                 | ✅     | `apps/api/internal/adapter/http/handler/jobs_handler.go`, `apps/api/internal/adapter/persistence/memory/jobs_repository.go` | `apps/api/internal/adapter/http/handler/jobs_handler_test.go`, `apps/api/internal/adapter/http/router/router_test.go`, `apps/api/test/integration/jobs_route_test.go` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/features/job-search.md`, `docs/phases/implementation-roadmap.md` |

### Iteration 1.2 - Identity & Preferences

| Item                                          | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| endpoint auth (`register/login/refresh/me`)   | ✅     | `apps/api/internal/adapter/http/handler/auth_handler.go`, `apps/api/internal/app/identity/service.go`, `apps/api/internal/platform/auth/token_manager.go`, `apps/api/cmd/api/main.go` | `apps/api/internal/adapter/http/handler/auth_handler_test.go`, `apps/api/internal/app/identity/service_test.go`, `apps/api/internal/platform/auth/token_manager_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/auth.md`, `docs/flows/auth-subscription-flow.md`, `docs/api/errors.md` |
| endpoint preferences (`GET/PUT /preferences`) | ✅     | `apps/api/internal/adapter/http/handler/preferences_handler.go`, `apps/api/internal/app/identity/service.go`, `apps/api/internal/adapter/persistence/memory/identity_repository.go`, `apps/api/internal/adapter/http/router/router.go` | `apps/api/internal/adapter/http/handler/preferences_handler_test.go`, `apps/api/internal/app/identity/service_test.go`, `apps/api/internal/adapter/persistence/memory/identity_repository_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/preferences.md`, `docs/features/smart-notification.md`, `docs/api/errors.md` |
| middleware auth + role dasar                  | ✅     | `apps/api/internal/adapter/http/middleware/auth.go`, `apps/api/internal/adapter/http/router/router.go`, `apps/api/cmd/api/main.go` | `apps/api/internal/adapter/http/middleware/auth_test.go`, `apps/api/internal/adapter/http/router/router_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/README.md`, `docs/architecture/system_architecture.md`, `docs/standards/security-observability-standards.md` |
| test auth flow                                | ✅     | `apps/api/internal/adapter/http/handler/auth_handler.go`, `apps/api/internal/adapter/http/handler/preferences_handler.go`, `apps/api/internal/adapter/http/router/router.go` | `apps/api/test/integration/auth_preferences_flow_test.go`, `apps/api/internal/adapter/http/handler/auth_handler_test.go`, `apps/api/internal/adapter/http/handler/preferences_handler_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/phases/implementation-roadmap.md`, `docs/flows/auth-subscription-flow.md` |

### Iteration 1.3 - Notification MVP

| Item                        | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| matcher worker rule-based   | ✅     | `apps/api/internal/app/notification/matcher.go`, `apps/api/internal/domain/notification/notification.go`, `apps/api/internal/adapter/queue/memory/queue.go` | `apps/api/internal/app/notification/matcher_test.go`, `apps/api/internal/adapter/queue/memory/queue_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/features/smart-notification.md`, `docs/flows/scraping-matching-flow.md`, `docs/architecture/redis_queue_cache.md` |
| email notifier worker       | ✅     | `apps/api/internal/app/notification/notifier.go`, `apps/api/internal/adapter/notifier/email/logger_sender.go`, `apps/api/cmd/notifier/main.go` | `apps/api/internal/app/notification/notifier_test.go`, `apps/api/test/integration/notification_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/features/smart-notification.md`, `docs/architecture/system_architecture.md`, `docs/flows/scraping-matching-flow.md` |
| status notifikasi tersimpan | ✅     | `apps/api/internal/adapter/persistence/memory/notification_repository.go`, `apps/api/internal/domain/notification/notification.go`, `apps/api/internal/app/notification/notifier.go` | `apps/api/internal/adapter/persistence/memory/notification_repository_test.go`, `apps/api/internal/app/notification/notifier_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/architecture/database.md`, `docs/features/smart-notification.md`, `docs/api/errors.md` |
| test notification flow      | ✅     | `apps/api/internal/app/notification/matcher.go`, `apps/api/internal/app/notification/notifier.go`, `apps/api/cmd/notifier/main.go` | `apps/api/test/integration/notification_flow_test.go`, `apps/api/internal/app/notification/matcher_test.go`, `apps/api/internal/app/notification/notifier_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/phases/implementation-roadmap.md`, `docs/flows/scraping-matching-flow.md` |

## Phase 2 - Billing Hardening (Mayar)

### Iteration 2.1 - Checkout Orchestration

| Item                                               | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| -------------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| endpoint `POST /billing/checkout-session`          | ✅     | `apps/api/internal/adapter/http/handler/billing_handler.go`, `apps/api/internal/adapter/http/router/router.go`, `apps/api/cmd/api/main.go` | `apps/api/internal/adapter/http/handler/billing_handler_test.go`, `apps/api/internal/adapter/http/router/router_test.go`, `apps/api/test/integration/billing_checkout_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/billing.md`, `docs/phases/implementation-roadmap.md`, `docs/api/errors.md` |
| adapter Mayar `customer/create` + `invoice/create` | ✅     | `apps/api/internal/adapter/billing/mayar/client.go`, `apps/api/internal/app/billing/service.go`, `apps/api/internal/platform/config/config.go` | `apps/api/internal/adapter/billing/mayar/client_test.go`, `apps/api/internal/app/billing/service_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/mayar-headless.md`, `docs/architecture/mayar-integration.md`, `docs/features/subscription-billing.md` |
| transaksi pending tersimpan                        | ✅     | `apps/api/internal/domain/billing/billing.go`, `apps/api/internal/adapter/persistence/memory/billing_repository.go`, `apps/api/internal/app/billing/service.go` | `apps/api/internal/adapter/persistence/memory/billing_repository_test.go`, `apps/api/internal/app/billing/service_test.go`, `apps/api/test/integration/billing_checkout_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/architecture/database.md`, `docs/features/subscription-billing.md`, `docs/phases/implementation-roadmap.md` |

### Iteration 2.2 - Webhook Reliability

| Item                             | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| -------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| endpoint `POST /webhook/mayar`   | ✅     | `apps/api/internal/adapter/http/handler/billing_handler.go`, `apps/api/internal/adapter/http/router/router.go`, `apps/api/cmd/api/main.go` | `apps/api/internal/adapter/http/handler/billing_handler_test.go`, `apps/api/internal/adapter/http/router/router_test.go`, `apps/api/test/integration/billing_webhook_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/webhooks.md`, `docs/flows/auth-subscription-flow.md`, `docs/phases/implementation-roadmap.md` |
| idempotency key transaksi        | ✅     | `apps/api/internal/app/billing/webhook.go` (`mayar:{event}:{transactionId}`), `apps/api/internal/adapter/persistence/memory/billing_repository.go`, `apps/api/internal/domain/billing/billing.go` | `apps/api/internal/app/billing/webhook_test.go`, `apps/api/internal/adapter/persistence/memory/billing_repository_test.go`, `apps/api/test/integration/billing_webhook_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/webhooks.md`, `docs/architecture/mayar-integration.md`, `docs/features/subscription-billing.md` |
| audit tabel `webhook_deliveries` | ✅     | `apps/api/internal/domain/billing/billing.go` (`WebhookDelivery`), `apps/api/internal/adapter/persistence/memory/billing_repository.go` (`RecordWebhookDelivery`), `apps/api/internal/app/billing/webhook.go` | `apps/api/internal/adapter/persistence/memory/billing_repository_test.go`, `apps/api/internal/app/billing/webhook_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/architecture/database.md`, `docs/api/webhooks.md`, `docs/phases/implementation-roadmap.md` |
| test duplicate webhook           | ✅     | `apps/api/internal/app/billing/webhook.go`, `apps/api/internal/adapter/http/handler/billing_handler.go`, `apps/api/internal/adapter/http/router/router.go` | `apps/api/internal/app/billing/webhook_test.go`, `apps/api/internal/adapter/http/handler/billing_handler_test.go`, `apps/api/test/integration/billing_webhook_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/phases/implementation-roadmap.md`, `docs/api/webhooks.md` |

### Iteration 2.3 - Reconciliation & Recovery

| Item                          | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| ----------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| endpoint/status billing user  | ✅     | `apps/api/internal/adapter/http/handler/billing_handler.go`, `apps/api/internal/app/billing/read_models.go`, `apps/api/internal/adapter/http/router/router.go` | `apps/api/internal/adapter/http/handler/billing_handler_test.go`, `apps/api/internal/app/billing/read_models_test.go`, `apps/api/test/integration/billing_read_reconcile_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/billing.md`, `docs/features/subscription-billing.md`, `docs/phases/implementation-roadmap.md` |
| rekonsiliasi ke Mayar         | ✅     | `apps/api/internal/app/billing/reconciliation.go`, `apps/api/internal/adapter/billing/mayar/client.go` (`GetInvoiceByID`), `apps/api/internal/adapter/persistence/memory/billing_repository.go` | `apps/api/internal/app/billing/reconciliation_test.go`, `apps/api/internal/adapter/billing/mayar/client_test.go`, `apps/api/test/integration/billing_read_reconcile_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/billing.md`, `docs/architecture/mayar-integration.md`, `docs/phases/implementation-roadmap.md` |
| retry workflow error upstream | ✅     | `apps/api/internal/adapter/billing/mayar/client.go` (retry 429/5xx + backoff), `apps/api/internal/app/billing/reconciliation.go` (retryable failure accounting), `apps/api/internal/platform/config/config.go` | `apps/api/internal/adapter/billing/mayar/client_test.go`, `apps/api/internal/app/billing/reconciliation_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/billing.md`, `docs/api/mayar-headless.md`, `docs/standards/security-observability-standards.md` |
| alert billing anomaly         | ✅     | `apps/api/cmd/billing-worker/main.go`, `apps/api/internal/app/billing/reconciliation.go` | `apps/api/internal/app/billing/reconciliation_test.go`, `apps/api/internal/platform/worker/runner_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/architecture/mayar-integration.md`, `docs/phases/implementation-roadmap.md`, `docs/api/billing.md` |

## Phase 3 - Growth Features (Optional Set A)

| Item                              | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| saved searches + alert rules      | ✅     | `apps/api/internal/domain/growth/growth.go`, `apps/api/internal/app/growth/service.go`, `apps/api/internal/adapter/persistence/memory/growth_repository.go`, `apps/api/internal/adapter/http/handler/growth_handler.go`, `apps/api/internal/adapter/http/router/router.go` | `apps/api/internal/app/growth/service_test.go`, `apps/api/internal/adapter/persistence/memory/growth_repository_test.go`, `apps/api/internal/adapter/http/handler/growth_handler_test.go`, `apps/api/test/integration/growth_features_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/growth.md`, `docs/features/optional-features.md`, `docs/flows/growth-engagement-flow.md` |
| in-app notification center        | ✅     | `apps/api/internal/domain/notification/notification.go`, `apps/api/internal/app/notification/center.go`, `apps/api/internal/adapter/persistence/memory/notification_repository.go`, `apps/api/internal/adapter/http/handler/notification_handler.go`, `apps/api/internal/adapter/http/router/router.go` | `apps/api/internal/app/notification/center_test.go`, `apps/api/internal/adapter/persistence/memory/notification_repository_test.go`, `apps/api/internal/adapter/http/handler/notification_handler_test.go`, `apps/api/test/integration/growth_features_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/growth.md`, `docs/flows/growth-engagement-flow.md`, `docs/features/optional-features.md` |
| company watchlist                 | ✅     | `apps/api/internal/domain/growth/growth.go`, `apps/api/internal/app/growth/service.go`, `apps/api/internal/adapter/persistence/memory/growth_repository.go`, `apps/api/internal/adapter/http/handler/growth_handler.go` | `apps/api/internal/app/growth/service_test.go`, `apps/api/internal/adapter/persistence/memory/growth_repository_test.go`, `apps/api/internal/adapter/http/handler/growth_handler_test.go`, `apps/api/test/integration/growth_features_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/growth.md`, `docs/features/optional-features.md`, `docs/flows/growth-engagement-flow.md` |
| frequency/digest control          | ✅     | `apps/api/internal/domain/identity/identity.go`, `apps/api/internal/app/identity/service.go`, `apps/api/internal/adapter/persistence/memory/identity_repository.go`, `apps/api/internal/adapter/http/handler/preferences_handler.go`, `apps/api/internal/app/notification/matcher.go` | `apps/api/internal/app/identity/service_test.go`, `apps/api/internal/adapter/http/handler/preferences_handler_test.go`, `apps/api/internal/app/notification/matcher_test.go`, `apps/api/test/integration/growth_features_flow_test.go`, `make -C apps/api test` | `act pull_request -W .github/workflows/ci-api.yml -j lint-test-build` (job succeeded) | `docs/api/preferences.md`, `docs/api/growth.md`, `docs/features/smart-notification.md`, `docs/features/optional-features.md` |

## Phase 4 - Advanced Monetization & Intelligence (Optional Set B)

| Item                                         | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| -------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| advanced billing (coupon/installment/qrcode) | 🟡     | `apps/api/internal/app/billing/service.go`, `apps/api/internal/adapter/http/handler/billing_handler.go`, `apps/api/internal/adapter/billing/mayar/client.go`, `apps/api/internal/domain/billing/billing.go`, `apps/api/pkg/errcode/codes.go` | `apps/api/internal/app/billing/service_test.go`, `apps/api/internal/adapter/http/handler/billing_handler_test.go`, `apps/api/internal/adapter/billing/mayar/client_test.go`, `apps/api/test/integration/billing_checkout_flow_test.go` | local gate pass: `make -C apps/api lint test build check-migrations` | `docs/api/billing.md`, `docs/api/errors.md`, `docs/features/optional-features.md`, `docs/phases/implementation-roadmap.md` |
| salary normalization lanjutan                | ⬜     | -             | -             | -           | -                 |
| semantic recommendation                      | ⬜     | -             | -             | -           | -                 |
| B2B team subscription                        | ⬜     | -             | -             | -           | -                 |
| ADR untuk keputusan besar phase ini          | ⬜     | -             | -             | -           | -                 |

## 5) Cara Pakai

1. Ubah status item sesuai progres (`⬜` → `🟡` → `✅`).
2. Isi seluruh kolom evidence saat status `✅`.
3. Validasi gate mengacu ke `docs/standards/ci-quality-gates.md`.
4. Update checklist ini di PR yang sama dengan implementasi.
