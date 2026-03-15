# Implementation Roadmap (Phase & Iteration)

Checklist status implementasi per phase:

- [`implementation-checklist.md`](./implementation-checklist.md)

## Snapshot Saat Ini

Implementasi backend saat ini:

- baseline `apps/api` tersedia (cmd, internal, pkg, migrations, healthcheck, worker skeleton),
- baseline `apps/web` tersedia (Next.js App Router + domain-first skeleton),
- validasi lokal lint/typecheck/test/build sudah dijalankan,
- workflow `ci-api` dan `ci-web` tervalidasi lewat run lokal `act`,
- **Iteration 1.1 backend** sudah diimplementasikan (source adapters + token provider + dedup + Jobs API list/detail + scrape telemetry + regression tests).
- **Iteration 1.2 backend** sudah diimplementasikan (auth register/login/refresh/me + JWT middleware/role guard dasar + preferences GET/PUT + auth integration tests).
- **Iteration 1.3 backend** sudah diimplementasikan (matcher worker rule-based + email notifier + notification status persistence + notification flow tests).
- **Iteration 2.1 backend** sudah diimplementasikan (checkout orchestration endpoint + adapter Mayar customer/invoice + pending transaction persistence + idempotency replay + rate-limit protection).
- **Iteration 2.2 backend** sudah diimplementasikan (webhook receiver + token validation + idempotency delivery key + transaction status update + premium activation + duplicate handling tests).
- **Iteration 2.3 backend** sudah diimplementasikan (`GET /billing/status` + `GET /billing/transactions`, reconciliation ke Mayar `GET /invoice/{id}`, retry-aware recovery path, dan anomaly warning pada billing worker).
- **Phase 3 backend** sudah diimplementasikan (saved searches, company watchlist, notification center, dan preference digest control).
- **Phase 4 backend increment 1** sudah mulai diimplementasikan (coupon-enabled checkout: validasi `coupon_code` ke Mayar + invoice amount diskon + metadata amount response), diikuti hardening `salary normalization` untuk mapping scraper (`min-only`/`max-only`/`exact`/`range`) termasuk comparator (`<= ...`) dan shorthand bulanan (`Rp 8 – Rp 12 per month`), hardening validasi `redirect_url` checkout untuk local development (HTTP localhost/loopback only), retry guard yang mereuse checkout pending agar tidak mudah terkena `429`, serta parser kompatibilitas payload Mayar (object/array + link/expiry variants).
- **Phase 5 (AI Career Intelligence & Value Layer)** sudah berjalan pada increment 1-3 backend (`POST /ai/search-assistant`, `POST /ai/job-fit-summary`, `POST /ai/cover-letter-draft`, `GET /ai/usage`) dan increment 1 frontend (`/account/ai-tools` + AI usage meter + assistant/job-fit/cover-letter UI).

Status saat ini: **Phase 0 complete + Phase 1 backend (Iteration 1.1-1.3) complete + Phase 2 backend (Iteration 2.1-2.3) complete + Phase 3 backend complete + Phase 4 backend in progress (increment 1 + salary normalization + checkout redirect/rate-limit hardening) + Phase 5 in progress (increment 1-3 backend + frontend increment 1) + Phase 6 (Application Tracker) backend complete + Phase 6 frontend complete**.

## Rencana Lanjutan (Document-First, One-by-One)

Sebelum implementasi feature lanjutan, roadmap eksekusi dikunci terlebih dahulu agar setiap perubahan berjalan bertahap dan tervalidasi.

| Milestone | Fokus | Status | Catatan Eksekusi |
| --------- | ----- | ------ | ---------------- |
| M0 | Dokumentasi rencana perubahan menyeluruh | ✅ Complete | Rencana detail sudah ditetapkan sebelum coding lanjutan |
| M1 | Migrasi fondasi persistence ke PostgreSQL | ✅ Complete | Adapter PostgreSQL + wiring API/worker + migration cutover + queue persistence sudah aktif |
| M2 | Pass komentar/docstrings sesuai standar | ✅ Complete | Doc comments Go untuk simbol exported + TSDoc pada service/lib frontend sudah ditambahkan |
| M3 | Migrasi copy UI + API user-facing ke English | ✅ Complete | Seluruh copy user-facing frontend + pesan notifier backend sudah di-English-kan tanpa mengubah kontrak |
| M4 | Redesign frontend ala SaaS + hardening growth | ✅ Complete | UI frontend sudah direfresh berbasis design tokens + observability web vitals + e2e growth coverage + refinement visual pass ala Paper sudah ditutup |
| M5 | Eksekusi Phase 4 backend | 🟡 In Progress | Increment 1 aktif: coupon-enabled checkout pada billing + hardening salary normalization mapping scraper (comparator + shorthand monthly range) + redirect URL/rate-limit hardening + kompatibilitas parser payload Mayar |
| M6 | Phase 5 AI value layer (backend + frontend) | 🟡 In Progress | Increment 1-3 backend aktif + frontend increment 1 aktif di `/account/ai-tools`; interview prep masih planned |
| M7 | Phase 6 Application Tracker (backend + frontend) | ✅ Complete | Phase 6 backend complete (38 tests pass) + Phase 6 frontend complete (33 test files pass, typecheck clean) |

Aturan eksekusi:

1. Jalankan satu milestone pada satu waktu (tidak lompat milestone).
2. Tiap milestone wajib lulus gate kualitas (G1-G8) sebelum lanjut.
3. Jika ada mismatch kontrak FE-BE, perbaiki kontrak/dokumen pada milestone aktif sebelum maju.

## Prinsip Eksekusi

- Milestone phase hanya boleh ditutup jika acceptance criteria phase terpenuhi.
- Required gates mengikuti `docs/standards/ci-quality-gates.md` (G1-G8).
- Evidence harus konsisten dengan model di `docs/standards/README.md`.

## Quality Gates Lintas Phase (Wajib)

| Gate  | Ekspektasi Minimum                                   |
| ----- | ---------------------------------------------------- |
| G1-G2 | Lint + type/static analysis pass                     |
| G3-G4 | Test wajib pass sesuai trigger matrix                |
| G5    | Coverage threshold terpenuhi atau exception resmi    |
| G6    | Security scan tanpa blocker                          |
| G7    | Docs/checklist update + markdown link check pass     |
| G8    | Approval sesuai risk + rollback plan untuk high-risk |

## Definisi Phase Done

Phase dinyatakan selesai jika seluruh kondisi berikut terpenuhi:

1. Semua item **mandatory** phase pada checklist berstatus `✅ Implemented`.
2. Semua gate required untuk phase tersebut pass.
3. Tidak ada blocker open untuk reliability/security pada scope phase.

## Milestone Map (Roadmap ↔ Standards)

| Phase                         | Outcome Utama                                             | Required Gates                                | Acceptance Evidence                                            |
| ----------------------------- | --------------------------------------------------------- | --------------------------------------------- | -------------------------------------------------------------- |
| Phase 0 - Foundation          | Fondasi service, schema, observability dasar siap dipakai | G1, G2, G3, G6, G7, G8 (+ G4/G5 jika trigger) | healthcheck pass, migrasi tervalidasi, checklist phase lengkap |
| Phase 1 - MVP Core            | Search + auth + notification MVP usable end-to-end        | G1-G8                                         | test journey MVP pass, regression suite tersedia               |
| Phase 2 - Billing Hardening   | Billing flow stabil, idempotent, dapat diaudit            | G1-G8                                         | webhook idempotency + reconciliation evidence + alert aktif    |
| Phase 3 - Growth (opsional)   | Fitur retensi/engagement bertambah tanpa regress core     | G1-G8 (untuk scope perubahan)                 | e2e growth + metrik adopsi tersedia                            |
| Phase 4 - Advanced (opsional) | Monetisasi lanjutan + intelligence stabil                 | G1-G8 + ADR wajib untuk keputusan besar       | contract stabil, fallback tervalidasi, ADR tersedia            |
| Phase 5 - AI Value Layer (opsional) | Asisten karier AI + personalisasi premium/free yang terukur | G1-G8 + AI safety/usage policy + ADR arsitektur | kontrak AI gateway stabil, quota/guardrail tervalidasi, UX AI lulus uji utama |

## Phase 0 - Foundation & Baseline

### Objective

Membangun fondasi teknis implementasi yang siap dikembangkan dan diuji.

### Scope

- setup service API, worker, scraper, database, redis,
- standard response envelope, error code, auth middleware,
- logging + request/trace id + monitoring dasar.

### Exit Criteria (Measurable)

- service utama dapat dijalankan pada environment development/CI,
- healthcheck API/worker lulus,
- migrasi database `up`/`down` tervalidasi tanpa konflik,
- seluruh item mandatory Phase 0 di checklist berstatus `✅` dengan evidence lengkap.

## Phase 1 - MVP Core Delivery

### Objective

Menyediakan value inti Bisakerja: agregasi lowongan, pencarian, dan notifikasi premium.

### Iteration 1.1 - Aggregation & Search

- source adapter `glints`, `kalibrr`, `jobstreet` aktif,
- token provider untuk source auth-required (JobStreet) terpasang,
- deduplikasi (`UNIQUE(source, original_job_id)`),
- endpoint jobs list/detail + pagination + filter,
- telemetry scrape run per source tersedia untuk troubleshooting cepat.

### Iteration 1.2 - Identity & Preferences

- register/login/refresh/me,
- endpoint read/update preferences (`GET/PUT /preferences`),
- validasi premium gating untuk notifikasi.

### Iteration 1.3 - Notification MVP

- worker matching rule-based,
- email notifier,
- pencatatan status `notifications`.

### Exit Criteria (Measurable)

- journey MVP (auth → search → preference → notification trigger) lulus pada test wajib,
- regression test untuk auth/jobs/notification tersedia,
- seluruh gate required untuk Phase 1 pass di CI,
- checklist item mandatory Phase 1 berstatus `✅` dengan evidence lengkap.

## Phase 2 - Billing Hardening (Mayar)

### Objective

Membuat monetisasi stabil, idempotent, dan dapat diaudit operasional.

### Iteration 2.1 - Checkout Orchestration

- endpoint internal `POST /billing/checkout-session`,
- sinkronisasi customer + create invoice Mayar,
- simpan transaksi status `pending`.

### Iteration 2.2 - Webhook Reliability

- endpoint `POST /webhook/mayar`,
- idempotency key berbasis event transaksi,
- audit table `webhook_deliveries`.

### Iteration 2.3 - Reconciliation & Recovery

- endpoint read model billing internal (`GET /billing/status`, `GET /billing/transactions`),
- rekonsiliasi status transaksi via endpoint Mayar `GET /invoice/{id}`,
- retry-aware recovery untuk transient upstream failures + anomaly summary untuk transaksi stale.

### Exit Criteria (Measurable)

- pembayaran sukses mengaktifkan premium otomatis,
- event webhook duplikat tidak menyebabkan update ganda (dibuktikan test idempotency),
- histori transaksi + webhook dapat ditelusuri end-to-end,
- alert minimum billing aktif sesuai standar observability.

## Phase 3 - Growth Features (Optional Set A)

### Objective

Meningkatkan retensi dan engagement user tanpa merusak core experience.

### Scope

- saved searches + alert rules (siap di-hook ke scheduler),
- in-app notification center,
- company watchlist,
- notification frequency/digest control.

### Exit Criteria (Measurable)

- e2e journey fitur growth prioritas pass,
- metrik adopsi fitur growth tersedia,
- tidak ada regression pada flow core MVP berdasarkan suite regression.

## Phase 4 - Advanced Monetization & Intelligence (Optional Set B)

### Objective

Meningkatkan conversion dan relevansi secara bertahap dengan risiko terkontrol.

### Scope

- advanced billing (coupon, installment, QR checkout),
- salary normalization lanjutan,
- semantic recommendation layer,
- B2B team subscription (opsional lanjut).

### Exit Criteria (Measurable)

- kontrak API + schema DB untuk fitur baru stabil,
- observability mencakup flow monetisasi lanjutan,
- fallback/recovery untuk dependency eksternal tervalidasi,
- ADR tersedia untuk keputusan arsitektur signifikan pada phase ini.

## Phase 5 - AI Career Intelligence & Value Layer (Optional Set C)

### Objective

Menambahkan value nyata berbasis AI untuk user free dan premium tanpa mengorbankan stabilitas core flow.

### Scope

- OpenAI-compatible gateway di backend dengan provider config yang bisa dicustom (`base_url`, `api_key`, `model`, `timeout`) via environment/runtime config.
- AI query assistant untuk membantu user menyusun pencarian lowongan yang lebih relevan.
- AI job fit summary (ringkasan kecocokan profil/preferensi terhadap lowongan) sebagai value premium.
- AI cover letter draft & interview prep pack (template + talking points) dengan quota berbasis tier.
- Usage metering + quota enforcement per user/tier agar biaya AI terkendali.
- Audit-safe observability: structured logs tanpa menyimpan prompt sensitif mentah.

Progress increment 1-3 yang sudah aktif di backend:

- endpoint `POST /api/v1/ai/search-assistant`,
- endpoint `POST /api/v1/ai/job-fit-summary`,
- endpoint `POST /api/v1/ai/cover-letter-draft`,
- endpoint `GET /api/v1/ai/usage`,
- migration `ai_usage_logs` untuk metering + quota harian per tier.

Progress increment 1 frontend yang sudah aktif:

- route protected `/account/ai-tools`,
- AI usage meter untuk `search_assistant`, `job_fit_summary`, `cover_letter_draft`,
- form AI search assistant + job-fit summary + cover letter draft,
- fallback UX untuk premium lock (`FORBIDDEN`), quota exhausted, dan provider unavailable.

### Free vs Premium Value Target

- **Free**:
  - AI search assistant quota harian terbatas.
  - Interview prep tips dasar (template standar).
- **Premium**:
  - quota AI lebih tinggi,
  - job-fit insight lebih mendalam,
  - cover letter draft dengan konteks job+profile,
  - prioritas respons (latency budget lebih ketat).

### Exit Criteria (Measurable)

- endpoint AI backend terdokumentasi + rate/quota policy jelas,
- konfigurasi provider AI bisa dialihkan via `AI_PROVIDER_BASE_URL` tanpa ubah code,
- guardrail keamanan (redaction + abuse/rate limit) aktif dan tervalidasi,
- metrik penggunaan AI per tier tersedia untuk evaluasi conversion premium,
- checklist phase backend + frontend untuk scope AI sudah berisi evidence minimal `📝`/`🟡` sesuai progres.

## Phase 6 - Application Tracker & Bookmark

### Objective

Meningkatkan stickiness produk dengan memberi user cara menyimpan lowongan (bookmark) dan melacak status lamaran secara aktif, dengan pembatasan jumlah tracking pada free tier untuk mendorong konversi upgrade.

### Scope

- Bookmark engine: simpan/hapus/list bookmark lowongan per user.
- Application tracker: lacak status pipeline lamaran (`applied` → `interview` → `offer` / `rejected` / `withdrawn`).
- Free tier limit: maksimal 5 active tracked applications (status bukan `rejected`/`withdrawn`).
- Premium tier: unlimited tracked applications.
- Frontend: BookmarkButton di job detail + dashboard `/account/tracker`.

### Progress Backend

- migration `000007_phase6_application_tracker` (tabel `bookmarks` + `tracked_applications`),
- domain/tracker, app/tracker service + repository memory + repository postgres,
- handler tracker (7 endpoint),
- error code `TRACKER_LIMIT_EXCEEDED`,
- integration test `tracker_flow_test.go`.

### Progress Frontend

- `apps/web/src/services/tracker.ts` (types + service functions),
- `apps/web/src/features/tracker/components/bookmark-button.tsx`,
- `apps/web/src/features/tracker/components/account-tracker-client.tsx`,
- `apps/web/src/app/account/tracker/page.tsx`,
- BookmarkButton diintegrasikan ke `apps/web/src/app/jobs/[id]/page.tsx`,
- nav link "Application tracker" ditambahkan ke `account-dashboard-nav.tsx`.

### Exit Criteria (Measurable)

- 38 Go tests pass (`go test ./...` clean),
- 33 frontend test files pass, `pnpm --filter web typecheck` zero errors,
- endpoint bookmark + application tracking terdokumentasi di `docs/api/tracker.md`,
- feature spec tersedia di `docs/features/application-tracker.md` (backend) + `docs/frontend/features/application-tracker.md` (frontend).
