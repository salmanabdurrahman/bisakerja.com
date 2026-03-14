# Frontend Implementation Roadmap (Foundation → MVP → Growth)

Checklist audit implementasi frontend:

- [`implementation-checklist.md`](./implementation-checklist.md)

## Snapshot Saat Ini

Implementasi **Phase 0 frontend** sudah dimulai:

- baseline `apps/web` (Next.js App Router) sudah terbentuk,
- struktur domain awal (`features`, `components`, `services`, `lib`) sudah tersedia,
- baseline lint/typecheck/test/build lulus pada validasi lokal,
- workflow `ci-web` tervalidasi lewat run lokal `act`,
- **Iteration 1.1 frontend** sudah diimplementasikan (`/jobs` list/filter/pagination + `/jobs/[id]` detail + loading/error/empty/429 states + discovery tests).

Status saat ini: **Phase 0 frontend baseline complete + Phase 1 Iteration 1.1 frontend complete**, siap lanjut ke Iteration 1.2 frontend.

## Prinsip Eksekusi

- Setiap phase harus punya artefak terverifikasi (kode, test, CI, docs).
- Gate yang berlaku mengikuti `docs/standards/ci-quality-gates.md` (G1-G8).
- Checklist phase frontend wajib diupdate di PR yang sama dengan implementasi.

## Quality Gates Lintas Phase (Wajib)

| Gate                         | Referensi                                                                                   | Evidence Minimum                         |
| ---------------------------- | ------------------------------------------------------------------------------------------- | ---------------------------------------- |
| G1 - Lint                    | [CI Quality Gates](../../standards/ci-quality-gates.md)                                     | hasil `eslint` + job CI                  |
| G2 - Type safety             | [CI Quality Gates](../../standards/ci-quality-gates.md)                                     | hasil `tsc --noEmit` + job CI            |
| G3 - Unit/Component test     | [Testing Strategy](../../standards/testing-strategy.md)                                     | path test + job CI                       |
| G4 - E2E test (jika trigger) | [Testing Strategy](../../standards/testing-strategy.md)                                     | suite journey + job CI                   |
| G5 - Coverage                | [Testing Strategy](../../standards/testing-strategy.md)                                     | angka coverage dan perbandingan baseline |
| G6 - Security                | [Security and Observability Standards](../../standards/security-observability-standards.md) | hasil secret/dependency/security scan    |
| G7 - Documentation           | [Engineering Standards Hub](../../standards/README.md)                                      | docs update + markdown link check pass   |
| G8 - Review readiness        | [Code Review Checklist](../../standards/code-review-checklist.md)                           | approval + required checks hijau         |

## Milestone Map (Roadmap ↔ Standards)

| Phase                  | Outcome Utama                                               | Required Gates                                | Acceptance Evidence                                       |
| ---------------------- | ----------------------------------------------------------- | --------------------------------------------- | --------------------------------------------------------- |
| Phase 0 - Foundation   | Baseline Next.js + CI frontend siap digunakan               | G1, G2, G3, G6, G7, G8 (+ G4/G5 jika trigger) | `apps/web` bootstrap, build pass, checklist phase lengkap |
| Phase 1 - MVP Delivery | Discovery + auth + preferences + checkout initiation usable | G1-G8                                         | journey MVP pass, regression test tersedia                |
| Phase 2 - Growth       | Fitur retensi/engagement aktif tanpa regress core           | G1-G8 (scope perubahan)                       | e2e growth pass + metrik adopsi tersedia                  |

## Phase 0 - Foundation

### Objective

Membangun baseline frontend Next.js yang aman, typed, dan siap dikembangkan.

### Scope

- bootstrap `apps/web` (Next.js App Router) dengan struktur domain-first,
- setup layer `features/`, `components/`, `services/`, `lib/`,
- definisi data fetching server/client sesuai boundary App Router,
- baseline lint, type-check, unit/component test, build check,
- UX states baseline (loading/error/empty) + accessibility minimum.

### Exit Criteria (Measurable)

- `apps/web` dapat dijalankan dan build pass,
- struktur modul mengikuti standar `nextjs-coding-standards.md`,
- gate required untuk Phase 0 pass,
- item mandatory Phase 0 pada checklist frontend berstatus `✅` dengan evidence lengkap.

## Phase 1 - MVP Delivery

### Objective

Mengirim value inti frontend: discovery lowongan, identitas user, dan aktivasi premium flow.

### Iteration 1.1 - Discovery & Search UX

- jobs listing dengan filter/sort/pagination,
- detail lowongan dengan metadata penting,
- loading/error/empty state eksplisit.

**API dependency wajib siap**

- `GET /api/v1/jobs`
- `GET /api/v1/jobs/:id`

**Output implementasi minimum**

- route `/jobs` dan `/jobs/[id]`,
- komponen search/filter/list/detail,
- test URL-driven filtering + detail `404`.

### Iteration 1.2 - Auth & Preferences UX

- register/login/logout/refresh session flow,
- halaman profil dan update preferences,
- proteksi route berbasis auth state.

**API dependency wajib siap**

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/auth/me`
- `GET /api/v1/preferences`
- `PUT /api/v1/preferences`
- `GET /api/v1/billing/status`

**Output implementasi minimum**

- auth routes + guard protected area,
- session refresh single-flight,
- halaman `/account` dan `/account/preferences`,
- test submit preferences + recovery saat `401`.

### Iteration 1.3 - Premium Activation UX

- halaman paket premium + inisiasi checkout,
- status subscription user di UI,
- fallback error/retry handling untuk checkout flow.

**API dependency wajib siap**

- `POST /api/v1/billing/checkout-session`
- `GET /api/v1/billing/status`
- `GET /api/v1/billing/transactions`
- `POST /api/v1/webhook/mayar` (dipakai backend async, diverifikasi via status)

**Output implementasi minimum**

- route `/pricing` dan `/account/subscription`,
- callback verify screen (mis. `/billing/success`),
- state handling `pending_payment`/`premium_active`/failed transaction,
- test checkout initiation + post-payment verify.

### Exit Criteria (Measurable)

- journey MVP frontend (auth → search → preferences → checkout start) lulus E2E,
- regression test komponen/flow kritikal tersedia,
- seluruh gate required Phase 1 pass di CI.

## Dependency Risk Notes (MVP)

- Pastikan `GET`/`PUT /api/v1/preferences` konsisten (field + default user baru) agar bootstrap form tidak regress.
- Entitlement premium wajib selalu mengikuti `billing/status.subscription_state` bila terjadi mismatch dengan `auth/me.is_premium`.

## Phase 2 - Growth

### Objective

Meningkatkan retensi, engagement, dan value premium melalui fitur growth.

### Scope

- saved searches + alert management,
- notification center di aplikasi,
- pengaturan frekuensi/digest notifikasi,
- optimasi performa halaman kritikal + monitoring web vitals.

### Exit Criteria (Measurable)

- fitur growth prioritas aktif tanpa regression pada flow core MVP,
- metrik adopsi fitur growth tersedia,
- e2e journey growth prioritas pass dan stabil,
- perubahan arsitektur signifikan didokumentasikan (ADR bila diperlukan).
