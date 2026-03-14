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
| halaman jobs list dengan filter/sort/pagination | ⬜     | -             | -             | -           | -                 |
| halaman job detail                              | ⬜     | -             | -             | -           | -                 |
| loading/error/empty states untuk search flow    | ⬜     | -             | -             | -           | -                 |
| handling `429`/retry state pada search          | ⬜     | -             | -             | -           | -                 |
| test flow discovery/search                      | ⬜     | -             | -             | -           | -                 |

### Iteration 1.2 - Auth & Preferences UX

| Item                                            | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| ----------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| flow register/login/logout frontend             | ⬜     | -             | -             | -           | -                 |
| proteksi route berbasis auth state              | ⬜     | -             | -             | -           | -                 |
| single-flight refresh untuk multi-request `401` | ⬜     | -             | -             | -           | -                 |
| halaman profile + update preferences            | ⬜     | -             | -             | -           | -                 |
| test flow auth/preferences                      | ⬜     | -             | -             | -           | -                 |

### Iteration 1.3 - Premium Activation UX

| Item                                                                  | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| --------------------------------------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| halaman paket premium + inisiasi checkout                             | ⬜     | -             | -             | -           | -                 |
| UI status subscription aktif/non-aktif                                | ⬜     | -             | -             | -           | -                 |
| verifikasi callback berbasis `billing/status` (bukan asumsi redirect) | ⬜     | -             | -             | -           | -                 |
| error/retry handling untuk checkout flow                              | ⬜     | -             | -             | -           | -                 |
| test flow checkout initiation                                         | ⬜     | -             | -             | -           | -                 |

## Phase 2 - Growth

| Item                                      | Status | Code Evidence | Test Evidence | CI Evidence | Docs/ADR Evidence |
| ----------------------------------------- | ------ | ------------- | ------------- | ----------- | ----------------- |
| saved searches management UI              | ⬜     | -             | -             | -           | -                 |
| notification center UI                    | ⬜     | -             | -             | -           | -                 |
| pengaturan frekuensi/digest notifikasi    | ⬜     | -             | -             | -           | -                 |
| observability web vitals halaman kritikal | ⬜     | -             | -             | -           | -                 |
| e2e flow fitur growth prioritas           | ⬜     | -             | -             | -           | -                 |

## 5) Cara Pakai

1. Ubah status item mengikuti progres (`⬜` → `🟡` → `✅`).
2. Saat status `✅`, isi seluruh kolom evidence.
3. Validasi gate mengacu ke `docs/standards/ci-quality-gates.md` dan `docs/standards/testing-strategy.md`.
4. Update checklist ini pada PR yang sama dengan perubahan implementasi frontend.
