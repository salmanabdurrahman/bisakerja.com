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

Status saat ini: **Phase 0 complete + Phase 1 backend (Iteration 1.1-1.3) complete + Phase 2 backend Iteration 2.1-2.2 complete**, siap lanjut ke backend Iteration 2.3 (reconciliation & recovery).

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

- endpoint/status page billing internal,
- rekonsiliasi via endpoint Mayar (`invoice`, `transactions`, `webhook/history`),
- playbook retry jika callback gagal.

### Exit Criteria (Measurable)

- pembayaran sukses mengaktifkan premium otomatis,
- event webhook duplikat tidak menyebabkan update ganda (dibuktikan test idempotency),
- histori transaksi + webhook dapat ditelusuri end-to-end,
- alert minimum billing aktif sesuai standar observability.

## Phase 3 - Growth Features (Optional Set A)

### Objective

Meningkatkan retensi dan engagement user tanpa merusak core experience.

### Scope

- saved searches + scheduled alerts,
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
