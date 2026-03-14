# Frontend Implementation Checklist (Phase Audit)

Checklist ini memastikan status implementasi frontend berbasis evidence yang konsisten.

## 1) Definisi Status

- `✅ Implemented`: kode production siap pakai + test + CI + docs/ADR evidence lengkap.
- `🟡 Partial`: implementasi sebagian; ada gap pada wiring/test/CI/docs.
- `⬜ Not Implemented`: belum ada implementasi kode.
- `📝 Documented`: requirement/arsitektur/dokumen tersedia, kode belum ada.
- `⛔ Blocked`: tidak bisa lanjut karena blocker yang teridentifikasi.

## 2) Aturan Evidence (Wajib untuk `✅`)

Setiap item `✅` wajib mengisi:

1. **Code Evidence**: path file frontend yang relevan.
2. **Test Evidence**: path test + nama suite.
3. **CI Evidence**: workflow/job yang pass.
4. **Docs/ADR Evidence**: path dokumen/ADR yang diperbarui, atau `N/A`.

Jika salah satu kolom kosong, status tidak boleh `✅`.

## 3) Snapshot Audit Saat Ini

Berdasarkan struktur repository saat ini, implementasi **Phase 0 frontend** sudah dimulai.

- `apps/web` sudah tersedia sebagai baseline Next.js App Router.
- Baseline lint/typecheck/test/build frontend lulus pada environment lokal dan workflow `ci-web` tervalidasi via `act`.
- Status keseluruhan frontend saat ini: **✅ Phase 0 frontend baseline complete**.

## 3.1) API Contract Readiness (Frontend Critical)

| API Dependency                          | Field Kunci yang Wajib Ada                                                                 | Status | Catatan                                    |
| --------------------------------------- | ------------------------------------------------------------------------------------------ | ------ | ------------------------------------------ |
| `POST /api/v1/auth/login`               | `access_token`, `refresh_token`, `expires_in`                                              | 📝     | sudah terdokumentasi di `docs/api/auth.md` |
| `POST /api/v1/auth/refresh`             | `access_token`, `expires_in`                                                               | 📝     | diperlukan untuk single-flight refresh     |
| `GET /api/v1/auth/me`                   | `id`, `email`, `name`, `role`, `is_premium`                                                | 📝     | dipakai di profile/session bootstrap       |
| `GET /api/v1/jobs`                      | `data[]`, `meta.pagination`                                                                | 📝     | dipakai list/pagination                    |
| `GET /api/v1/jobs/:id`                  | `data.url`, `data.description`                                                             | 📝     | dipakai detail + apply handoff             |
| `PUT /api/v1/preferences`               | payload preferensi + `updated_at`                                                          | 📝     | write path tersedia                        |
| `GET /api/v1/preferences`               | preferensi tersimpan user                                                                  | 📝     | read path untuk bootstrap form tersedia    |
| `POST /api/v1/billing/checkout-session` | `checkout_url`, `transaction_id`, `expired_at`, `subscription_state`, `transaction_status` | 📝     | dipakai upgrade initiation                 |
| `GET /api/v1/billing/status`            | `subscription_state`, `last_transaction_status`                                            | 📝     | source of truth entitlement                |
| `GET /api/v1/billing/transactions`      | `data[].status`, `data[].amount`                                                           | 📝     | histori pembayaran account/subscription    |
| `POST /api/v1/saved-searches`           | `id`, `query`, `frequency`, `is_active`, `created_at`                                      | 📝     | dipakai create saved search                |
| `GET /api/v1/saved-searches`            | `data[]` saved search user                                                                  | 📝     | dipakai bootstrap list saved search        |
| `DELETE /api/v1/saved-searches/:id`     | `id`                                                                                        | 📝     | dipakai hapus saved search                 |
| `GET /api/v1/notifications`             | `data[]`, `meta.pagination`                                                                 | 📝     | dipakai notification center                |
| `PATCH /api/v1/notifications/:id/read`  | `id`, `read_at`                                                                             | 📝     | dipakai mark as read                       |
| `PUT /api/v1/preferences/notification`  | `alert_mode`, `digest_hour`, `updated_at`                                                   | 📝     | dipakai kontrol digest mode                |
| `POST /api/v1/ai/search-assistant`      | `prompt`, `suggested_query`, `suggested_filters`, `quota_remaining`                         | 🟡     | backend increment 1 tersedia (`docs/api/ai.md`) |
| `POST /api/v1/ai/job-fit-summary`       | `job_id`, `fit_score`, `strengths`, `gaps`, `next_actions`                                  | 📝     | planned premium AI insight                 |
| `POST /api/v1/ai/cover-letter-draft`    | `job_id`, `tone`, `draft`, `quota_remaining`                                                 | 📝     | planned premium AI writing aid             |
| `POST /api/v1/ai/interview-prep`        | `job_id`, `questions`, `answer_tips`, `practice_plan`                                        | 📝     | planned free+premium AI prep               |
| `GET /api/v1/ai/usage`                  | `tier`, `daily_quota`, `used`, `remaining`, `reset_at`                                       | 🟡     | backend increment 1 tersedia (`docs/api/ai.md`) |

## 3.2) Rencana Lanjutan (Document-First)

Urutan implementasi lanjutan dikunci agar frontend berjalan seirama dengan readiness backend dan kontrak API.

| Milestone | Dampak ke Frontend | Status |
| --------- | ------------------ | ------ |
| M0 | Dokumentasi rencana perubahan menyeluruh | ✅ Complete |
| M1 | Stabilitas persistence backend (PostgreSQL) sebagai fondasi | ✅ Complete |
| M2 | Comments/TSDoc pass pada area frontend kritikal | ✅ Complete |
| M3 | Migrasi seluruh UI copy + message user-facing ke English | ✅ Complete |
| M4 | Redesign SaaS + penutupan web vitals dan growth e2e | ✅ Complete |
| M5 | Follow-up kontrak frontend untuk scope Phase 4 backend | 🟡 In Progress |
| M6 | AI experience layer (free + premium value) | 📝 Documented |

## 4) Checklist per Phase & Iteration

## Phase 0 - Foundation

| Item                                                              | Status | Code Evidence                                                                                                            | Test Evidence                                               | CI Evidence                                                                                | Docs/ADR Evidence                                                                                                                      |
| ----------------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------- | ------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| dokumen phase frontend tersedia                                   | 📝     | -                                                                                                                        | -                                                           | -                                                                                          | `docs/frontend/phases/README.md`, `docs/frontend/phases/implementation-roadmap.md`, `docs/frontend/phases/implementation-checklist.md` |
| bootstrap `apps/web` (Next.js App Router)                         | ✅     | `apps/web/src/app/layout.tsx`, `apps/web/src/app/page.tsx`, `apps/web/next.config.ts`, `apps/web/package.json`           | `pnpm --filter web build`                                   | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/architecture/app-structure.md`                                                                                          |
| struktur domain (`features/`, `components/`, `services/`, `lib/`) | ✅     | `apps/web/src/features/*`, `apps/web/src/components/*`, `apps/web/src/services/*`, `apps/web/src/lib/*`                  | `pnpm --filter web typecheck`, `pnpm --filter web test`     | `ci-web` job pass via `act`                                                                | `docs/frontend/architecture/app-structure.md`                                                                                          |
| baseline env config frontend                                      | ✅     | `apps/web/.env.example`, `apps/web/src/lib/config/env.ts`                                                                | `apps/web/tests/unit/env.test.ts`, `pnpm --filter web test` | `ci-web` job pass via `act`                                                                | `docs/frontend/architecture/data-state-flow.md`                                                                                        |
| lint frontend (`eslint`)                                          | ✅     | `apps/web/eslint.config.mjs`                                                                                             | `pnpm --filter web lint`                                    | `ci-web` job pass via `act`                                                                | `docs/standards/nextjs-coding-standards.md`                                                                                            |
| type-check frontend (`tsc --noEmit`)                              | ✅     | `apps/web/tsconfig.json`                                                                                                 | `pnpm --filter web typecheck`                               | `ci-web` job pass via `act`                                                                | `docs/standards/nextjs-coding-standards.md`                                                                                            |
| unit/component test baseline                                      | ✅     | `apps/web/vitest.config.ts`, `apps/web/vitest.setup.ts`, `apps/web/tests/**`                                             | `pnpm --filter web test`, `pnpm --filter web test:coverage` | `ci-web` job pass via `act`                                                                | `docs/standards/testing-strategy.md`                                                                                                   |
| build gate frontend (`next build`)                                | ✅     | `apps/web/package.json` (`build` script), `apps/web/src/app/*`                                                           | `pnpm --filter web build`                                   | `ci-web` job pass via `act`                                                                | `docs/frontend/phases/implementation-roadmap.md`                                                                                       |
| accessibility baseline (keyboard + semantic HTML)                 | ✅     | `apps/web/src/app/page.tsx`, `apps/web/src/components/layout/app-shell.tsx`, `apps/web/src/components/ui/state-card.tsx` | `pnpm --filter web lint`, `pnpm --filter web test`          | `ci-web` job pass via `act`                                                                | `docs/frontend/architecture/accessibility-performance.md`                                                                              |

## Phase 1 - MVP Delivery

### Iteration 1.1 - Discovery & Search UX

| Item                                            | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| ----------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| halaman jobs list dengan filter/sort/pagination | ✅     | `apps/web/src/app/jobs/page.tsx`, `apps/web/src/features/jobs/components/jobs-search-form.tsx`, `apps/web/src/features/jobs/components/jobs-pagination.tsx`, `apps/web/src/services/jobs.ts`, `apps/web/src/features/jobs/search-params.ts` | `apps/web/tests/unit/jobs-search-params.test.ts`, `apps/web/tests/components/jobs-page.test.tsx`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/job-discovery.md`, `docs/frontend/flows/discovery-flow.md`, `docs/api/jobs.md` |
| halaman job detail                              | ✅     | `apps/web/src/app/jobs/[id]/page.tsx`, `apps/web/src/app/jobs/[id]/loading.tsx`, `apps/web/src/features/jobs/components/jobs-list.tsx`, `apps/web/src/services/jobs.ts` | `apps/web/tests/components/job-detail-page.test.tsx`, `apps/web/tests/components/jobs-page.test.tsx`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/job-discovery.md`, `docs/frontend/flows/discovery-flow.md`, `docs/api/jobs.md` |
| loading/error/empty states untuk search flow    | ✅     | `apps/web/src/app/jobs/loading.tsx`, `apps/web/src/app/jobs/page.tsx`, `apps/web/src/features/jobs/components/jobs-state-panel.tsx` | `apps/web/tests/components/jobs-state-panel.test.tsx`, `apps/web/tests/components/jobs-page.test.tsx`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/job-discovery.md`, `docs/frontend/architecture/data-state-flow.md`, `docs/frontend/flows/discovery-flow.md` |
| handling `429`/retry state pada search          | ✅     | `apps/web/src/app/jobs/page.tsx`, `apps/web/src/lib/utils/fetch-json.ts`, `apps/web/src/features/jobs/components/jobs-state-panel.tsx` | `apps/web/tests/components/jobs-page.test.tsx`, `apps/web/tests/unit/fetch-json.test.ts`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/job-discovery.md`, `docs/frontend/flows/discovery-flow.md`, `docs/api/errors.md` |
| test flow discovery/search                      | ✅     | `apps/web/src/app/jobs/page.tsx`, `apps/web/src/app/jobs/[id]/page.tsx`, `apps/web/src/features/jobs/search-params.ts` | `apps/web/tests/unit/jobs-search-params.test.ts`, `apps/web/tests/components/jobs-page.test.tsx`, `apps/web/tests/components/job-detail-page.test.tsx`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/phases/implementation-roadmap.md`, `docs/frontend/traceability/frontend-backend-traceability.md` |

### Iteration 1.2 - Auth & Preferences UX

| Item                                            | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| ----------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| flow register/login/logout frontend             | ✅     | `apps/web/src/app/auth/login/page.tsx`, `apps/web/src/app/auth/register/page.tsx`, `apps/web/src/features/auth/components/login-form.tsx`, `apps/web/src/features/auth/components/register-form.tsx`, `apps/web/src/lib/auth/session-cookie.ts`, `apps/web/src/features/profile/components/account-page-client.tsx` | `apps/web/tests/components/register-form.test.tsx`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/auth-session.md`, `docs/api/auth.md`, `docs/frontend/phases/implementation-roadmap.md` |
| proteksi route berbasis auth state              | ✅     | `apps/web/src/proxy.ts`, `apps/web/src/app/account/page.tsx`, `apps/web/src/app/account/preferences/page.tsx`, `apps/web/src/lib/auth/redirect-path.ts`, `apps/web/src/lib/auth/server-session.ts` | `apps/web/tests/unit/auth-middleware.test.ts`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/auth-session.md`, `docs/frontend/architecture/app-structure.md`, `docs/frontend/traceability/frontend-backend-traceability.md` |
| single-flight refresh untuk multi-request `401` | ✅     | `apps/web/src/services/session-api-client.ts`, `apps/web/src/lib/auth/session-cookie.ts`, `apps/web/src/services/auth.ts` | `apps/web/tests/unit/session-api-client.test.ts`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/auth-session.md`, `docs/frontend/architecture/data-state-flow.md`, `docs/api/auth.md` |
| halaman profile + update preferences            | ✅     | `apps/web/src/app/account/page.tsx`, `apps/web/src/app/account/preferences/page.tsx`, `apps/web/src/features/profile/components/profile-summary.tsx`, `apps/web/src/features/billing/components/subscription-badge.tsx`, `apps/web/src/features/preferences/components/preferences-form.tsx`, `apps/web/src/features/preferences/components/notification-entitlement-banner.tsx`, `apps/web/src/features/preferences/components/account-preferences-client.tsx`, `apps/web/src/services/preferences.ts`, `apps/web/src/services/billing.ts` | `apps/web/tests/components/account-preferences-client.test.tsx`, `apps/web/tests/unit/session-api-client.test.ts`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/profile-account.md`, `docs/frontend/features/preferences-notifications.md`, `docs/api/preferences.md`, `docs/api/billing.md` |
| test flow auth/preferences                      | ✅     | `apps/web/src/app/auth/login/page.tsx`, `apps/web/src/app/account/preferences/page.tsx`, `apps/web/src/services/session-api-client.ts` | `apps/web/tests/components/register-form.test.tsx`, `apps/web/tests/components/account-preferences-client.test.tsx`, `apps/web/tests/unit/auth-middleware.test.ts`, `apps/web/tests/unit/session-api-client.test.ts`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/phases/implementation-roadmap.md`, `docs/frontend/phases/implementation-checklist.md`, `docs/frontend/traceability/frontend-backend-traceability.md` |

### Iteration 1.3 - Premium Activation UX

| Item                                                                  | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --------------------------------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| halaman paket premium + inisiasi checkout                             | ✅     | `apps/web/src/app/pricing/page.tsx`, `apps/web/src/features/billing/components/upgrade-cta.tsx`, `apps/web/src/services/billing.ts`, `apps/web/src/features/billing/checkout-session-cache.ts` | `apps/web/tests/components/upgrade-cta.test.tsx`, `apps/web/tests/unit/billing-service.test.ts`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/premium-upgrade.md`, `docs/frontend/flows/upgrade-billing-flow.md`, `docs/api/billing.md` |
| UI status subscription aktif/non-aktif                                | ✅     | `apps/web/src/features/billing/components/subscription-status-card.tsx`, `apps/web/src/app/account/subscription/page.tsx`, `apps/web/src/app/pricing/page.tsx`, `apps/web/src/features/billing/components/billing-history-list.tsx` | `apps/web/tests/components/account-subscription-page.test.tsx`, `apps/web/tests/components/billing-success-page.test.tsx`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/premium-upgrade.md`, `docs/frontend/features/profile-account.md`, `docs/api/billing.md` |
| verifikasi callback berbasis `billing/status` (bukan asumsi redirect) | ✅     | `apps/web/src/app/billing/success/page.tsx`, `apps/web/src/services/billing.ts`, `apps/web/src/app/pricing/page.tsx` | `apps/web/tests/components/billing-success-page.test.tsx`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/flows/upgrade-billing-flow.md`, `docs/frontend/features/premium-upgrade.md`, `docs/api/billing.md`, `docs/api/webhooks.md` |
| error/retry handling untuk checkout flow                              | ✅     | `apps/web/src/features/billing/components/upgrade-cta.tsx`, `apps/web/src/app/pricing/page.tsx`, `apps/web/src/app/billing/success/page.tsx` | `apps/web/tests/components/upgrade-cta.test.tsx`, `apps/web/tests/components/billing-success-page.test.tsx`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/features/premium-upgrade.md`, `docs/frontend/flows/upgrade-billing-flow.md`, `docs/api/errors.md` |
| test flow checkout initiation                                         | ✅     | `apps/web/src/app/pricing/page.tsx`, `apps/web/src/app/account/subscription/page.tsx`, `apps/web/src/app/billing/success/page.tsx`, `apps/web/src/features/billing/components/upgrade-cta.tsx` | `apps/web/tests/components/upgrade-cta.test.tsx`, `apps/web/tests/components/billing-success-page.test.tsx`, `apps/web/tests/components/account-subscription-page.test.tsx`, `apps/web/tests/unit/billing-service.test.ts`, `pnpm --filter web test` | `act pull_request -W .github/workflows/ci-web.yml -j lint-type-test-build` (job succeeded) | `docs/frontend/phases/implementation-roadmap.md`, `docs/frontend/phases/implementation-checklist.md`, `docs/frontend/traceability/frontend-backend-traceability.md` |

## Phase 2 - Growth

| Item                                      | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| ----------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| saved searches management UI              | ✅     | `apps/web/src/app/account/saved-searches/page.tsx`, `apps/web/src/features/growth/components/account-saved-searches-client.tsx`, `apps/web/src/services/growth.ts`, `apps/web/src/services/session-api-client.ts` | `apps/web/tests/components/account-saved-searches-client.test.tsx`, `apps/web/tests/unit/growth-service.test.ts`, `pnpm --filter web test` | `pnpm --filter web lint && pnpm --filter web test && pnpm --filter web build` | `docs/frontend/features/growth-engagement.md`, `docs/frontend/flows/growth-engagement-flow.md`, `docs/api/growth.md` |
| notification center UI                    | ✅     | `apps/web/src/app/account/notifications/page.tsx`, `apps/web/src/features/growth/components/account-notifications-client.tsx`, `apps/web/src/services/growth.ts`, `apps/web/src/services/session-api-client.ts` | `apps/web/tests/components/account-notifications-client.test.tsx`, `apps/web/tests/unit/growth-service.test.ts`, `pnpm --filter web test` | `pnpm --filter web lint && pnpm --filter web test && pnpm --filter web build` | `docs/frontend/features/growth-engagement.md`, `docs/frontend/flows/growth-engagement-flow.md`, `docs/api/growth.md` |
| pengaturan frekuensi/digest notifikasi    | ✅     | `apps/web/src/features/preferences/components/notification-digest-control.tsx`, `apps/web/src/features/preferences/components/account-preferences-client.tsx`, `apps/web/src/app/account/preferences/page.tsx`, `apps/web/src/services/preferences.ts`, `apps/web/src/services/session-api-client.ts` | `apps/web/tests/components/notification-digest-control.test.tsx`, `apps/web/tests/components/account-preferences-client.test.tsx`, `pnpm --filter web test` | `pnpm --filter web lint && pnpm --filter web test && pnpm --filter web build` | `docs/frontend/features/preferences-notifications.md`, `docs/frontend/features/growth-engagement.md`, `docs/api/preferences.md` |
| observability web vitals halaman kritikal | ✅     | `apps/web/src/features/observability/components/web-vitals-observer.tsx`, `apps/web/src/lib/observability/web-vitals.ts`, `apps/web/src/app/api/web-vitals/route.ts`, `apps/web/src/app/layout.tsx` | `apps/web/tests/unit/web-vitals.test.ts`, `pnpm --filter web test` | `pnpm --filter web lint && pnpm --filter web test && pnpm --filter web build` | `docs/frontend/architecture/accessibility-performance.md`, `docs/standards/nextjs-coding-standards.md` |
| e2e flow fitur growth prioritas           | ✅     | `apps/web/tests/e2e/growth-engagement-flow.test.tsx`, `apps/web/src/features/growth/components/account-saved-searches-client.tsx`, `apps/web/src/features/growth/components/account-notifications-client.tsx`, `apps/web/src/features/preferences/components/notification-digest-control.tsx` | `apps/web/tests/e2e/growth-engagement-flow.test.tsx`, `pnpm --filter web test` | `pnpm --filter web lint && pnpm --filter web test && pnpm --filter web build` | `docs/frontend/features/growth-engagement.md`, `docs/frontend/flows/growth-engagement-flow.md` |
| refinement visual pass ala Paper lintas halaman | ✅     | `apps/web/src/components/layout/app-shell.tsx`, `apps/web/src/components/ui/page-header.tsx`, `apps/web/src/app/page.tsx`, `apps/web/src/app/globals.css`, `apps/web/src/features/billing/components/upgrade-cta.tsx`, `apps/web/src/features/billing/components/subscription-status-card.tsx`, `apps/web/src/features/billing/components/subscription-badge.tsx` | `apps/web/tests/components/home-page.test.tsx`, `apps/web/tests/components/app-shell.test.tsx`, `apps/web/tests/components/account-subscription-page.test.tsx`, `pnpm --filter web test` | `pnpm --filter web lint && pnpm --filter web test && pnpm --filter web build` | `DESIGN_REFERENCE.md`, `docs/frontend/phases/implementation-roadmap.md`, `docs/frontend/phases/implementation-checklist.md` |

## Phase 5 - Phase 4 Backend Follow-up (Coupon Checkout UX)

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --- | --- | --- | --- | --- | --- |
| coupon-enabled checkout UX (`coupon_code` + amount breakdown) | 🟡 | `apps/web/src/features/billing/components/upgrade-cta.tsx`, `apps/web/src/features/billing/checkout-session-cache.ts`, `apps/web/src/services/billing.ts` | `apps/web/tests/components/upgrade-cta.test.tsx`, `apps/web/tests/unit/billing-service.test.ts` | local gate: `pnpm --filter web lint && pnpm --filter web test && pnpm --filter web build` | `docs/frontend/features/premium-upgrade.md`, `docs/frontend/flows/upgrade-billing-flow.md`, `docs/frontend/phases/implementation-roadmap.md`, `docs/api/billing.md` |

## Phase 6 - AI Experience & Premium Value Layer

| Item | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --- | --- | --- | --- | --- | --- |
| AI search assistant UX (query rewrite/refinement) | 📝 | - | - | - | `docs/frontend/phases/implementation-roadmap.md`, `docs/features/optional-features.md` |
| AI job-fit summary UX pada halaman detail/account | 📝 | - | - | - | `docs/frontend/phases/implementation-roadmap.md`, `docs/features/optional-features.md` |
| AI cover letter composer + interview prep workspace | 📝 | - | - | - | `docs/frontend/phases/implementation-roadmap.md`, `docs/features/optional-features.md` |
| quota meter + tier badge (free vs premium AI capability) | 📝 | - | - | - | `docs/frontend/phases/implementation-roadmap.md`, `docs/features/optional-features.md`, `docs/api/billing.md` |
| fallback UX saat AI provider unavailable/quota exhausted | 📝 | - | - | - | `docs/frontend/phases/implementation-roadmap.md`, `docs/standards/security-observability-standards.md`, `docs/api/errors.md` |
| traceability endpoint AI FE-BE | 📝 | - | - | - | `docs/frontend/traceability/frontend-backend-traceability.md`, `docs/frontend/phases/implementation-roadmap.md` |

## 5) Cara Pakai

1. Ubah status item mengikuti progres (`⬜` → `🟡` → `✅`).
2. Saat status `✅`, isi seluruh kolom evidence.
3. Validasi gate mengacu ke `docs/standards/ci-quality-gates.md` dan `docs/standards/testing-strategy.md`.
4. Update checklist ini pada PR yang sama dengan perubahan implementasi frontend.
