# Product Requirements Document

# Bisakerja Backend & Scraper Engine

Version: **1.6 (Production Readiness Audit)**
Owner: **Bisakerja Engineering Team**
Status: **Approved for Implementation**

## 1. Ringkasan Produk

Bisakerja adalah platform agregasi lowongan kerja yang mengumpulkan data dari berbagai portal, menyediakan pencarian cepat, dan mengirim notifikasi relevan ke user premium.

Backend mencakup:

- Job aggregation (scraper)
- REST API publik + protected
- Subscription billing API + payment webhook
- Smart notification engine
- Admin operational endpoints

## 2. Tujuan Produk

- Mempercepat job discovery user dibanding cek portal satu per satu.
- Menyediakan API stabil untuk web dan channel lain.
- Menyediakan alert lowongan baru yang relevan dan tepat waktu.

## 3. Scope MVP

### In Scope

- Scraper HTTP untuk source utama.
- Search jobs + pagination + filter.
- Auth JWT (register/login/refresh/me).
- User preferences untuk matching.
- Notification email untuk premium user.
- Integrasi billing dan webhook payment Mayar.
- Endpoint admin trigger scraper + stats.

### Out of Scope

- Headless scraping untuk portal complex SPA.
- Semantic AI matching.
- WhatsApp channel real-time sebagai kanal utama.

## 4. Business Model

- **Free**: Search dan lihat detail lowongan.
- **Pro (Rp49.000/bulan)**: akses data terbaru + smart notification.

## 5. Arsitektur Ringkas

- API Service (Golang)
- Scraper Worker
- Matcher/Notifier Worker
- PostgreSQL (source of truth)
- Redis (cache + queue + rate limiting)

Detail lengkap: [`../architecture/README.md`](../architecture/README.md)

## 6. Konvensi API & Data Canonical

### 6.1 Endpoint Convention

- Base URL: `/api/v1`
- Field path pada docs endpoint ditulis tanpa prefix base URL (contoh: `/jobs`, `/billing/status`).

### 6.2 Enum Canonical

- `subscription_state`: `free`, `pending_payment`, `premium_active`, `premium_expired`.
- `transactions.status`: `pending`, `reminder`, `success`, `failed`.
- `webhook_deliveries.processing_status`: `processed`, `ignored_duplicate`, `rejected`.

### 6.3 Ownership Boundary

- Endpoint user-scoped wajib memakai `JWT.sub` sebagai `user_id` sumber kebenaran.
- Akses lintas user hanya melalui endpoint `/admin/*` dengan role `admin`.

## 7. Daftar Dokumen Detail

PRD ini sengaja dipadatkan. Detail dipisahkan agar lebih maintainable.

### 7.1 Feature Specs

- Katalog fitur: [`../features/README.md`](../features/README.md)
- Job Aggregation: [`../features/job-aggregation.md`](../features/job-aggregation.md)
- Source Scraping Playbook: [`../features/source-scraping-playbook.md`](../features/source-scraping-playbook.md)
- Job Search: [`../features/job-search.md`](../features/job-search.md)
- Smart Notification: [`../features/smart-notification.md`](../features/smart-notification.md)
- Subscription & Billing: [`../features/subscription-billing.md`](../features/subscription-billing.md)
- Admin Operations: [`../features/admin-operations.md`](../features/admin-operations.md)
- Optional Features: [`../features/optional-features.md`](../features/optional-features.md)

### 7.2 API Specs

- API hub: [`../api/README.md`](../api/README.md)
- Auth: [`../api/auth.md`](../api/auth.md)
- Jobs: [`../api/jobs.md`](../api/jobs.md)
- Preferences: [`../api/preferences.md`](../api/preferences.md)
- Billing: [`../api/billing.md`](../api/billing.md)
- Admin: [`../api/admin.md`](../api/admin.md)
- Webhooks: [`../api/webhooks.md`](../api/webhooks.md)
- Error Codes: [`../api/errors.md`](../api/errors.md)
- Mayar Endpoint Mapping: [`../api/mayar-headless.md`](../api/mayar-headless.md)

### 7.3 Architecture Specs

- System Architecture: [`../architecture/system_architecture.md`](../architecture/system_architecture.md)
- Scraper Source Adapters: [`../architecture/scraper-source-adapters.md`](../architecture/scraper-source-adapters.md)
- Database Schema: [`../architecture/database.md`](../architecture/database.md)
- Redis/Queue/Cache: [`../architecture/redis_queue_cache.md`](../architecture/redis_queue_cache.md)
- Mayar Integration: [`../architecture/mayar-integration.md`](../architecture/mayar-integration.md)
- Monorepo Project Layout: [`../architecture/monorepo-project-layout.md`](../architecture/monorepo-project-layout.md)

### 7.4 Flow Specs

- Flows hub: [`../flows/README.md`](../flows/README.md)
- Auth & Subscription Flow: [`../flows/auth-subscription-flow.md`](../flows/auth-subscription-flow.md)
- Scraping & Matching Flow: [`../flows/scraping-matching-flow.md`](../flows/scraping-matching-flow.md)
- Search Serving Flow: [`../flows/search-serving-flow.md`](../flows/search-serving-flow.md)
- Admin Ops Flow: [`../flows/admin-ops-flow.md`](../flows/admin-ops-flow.md)

### 7.5 Phase & Iteration Plan

- Phases hub: [`../phases/README.md`](../phases/README.md)
- Implementation kickoff guide: [`../phases/implementation-kickoff.md`](../phases/implementation-kickoff.md)
- Implementation roadmap: [`../phases/implementation-roadmap.md`](../phases/implementation-roadmap.md)
- Implementation checklist: [`../phases/implementation-checklist.md`](../phases/implementation-checklist.md)

### 7.6 Engineering Standards

- Standards hub: [`../standards/README.md`](../standards/README.md)
- Go coding standards: [`../standards/go-coding-standards.md`](../standards/go-coding-standards.md)
- Next.js coding standards: [`../standards/nextjs-coding-standards.md`](../standards/nextjs-coding-standards.md)
- Comments/docstrings: [`../standards/comments-and-docstrings.md`](../standards/comments-and-docstrings.md)
- Testing strategy: [`../standards/testing-strategy.md`](../standards/testing-strategy.md)
- CI quality gates: [`../standards/ci-quality-gates.md`](../standards/ci-quality-gates.md)
- Security/observability standards: [`../standards/security-observability-standards.md`](../standards/security-observability-standards.md)
- Code review checklist: [`../standards/code-review-checklist.md`](../standards/code-review-checklist.md)
- ADR guidelines: [`../standards/adr-guidelines.md`](../standards/adr-guidelines.md)

### 7.7 Related Frontend Docs

- Frontend PRD: [`../prd/bisakerja-frontend-prd.md`](../prd/bisakerja-frontend-prd.md)
- Frontend docs hub: [`../frontend/README.md`](../frontend/README.md)
- Frontend-backend traceability matrix: [`../frontend/traceability/frontend-backend-traceability.md`](../frontend/traceability/frontend-backend-traceability.md)

## 8. Non-Functional Requirements (MVP Minimum)

| Area | Target |
|---|---|
| API availability (core endpoints) | >= 99.9% per bulan |
| Jobs API p95 latency | < 300 ms |
| Billing checkout p95 latency | < 2 detik |
| Webhook processing p95 latency | < 500 ms |
| API 5xx rate | <= 0.1% / 5 menit |
| Webhook success ratio | >= 99% |
| Outbound Mayar rate-limit hit ratio | < 1% per jam |

## 9. Acceptance (Implementation Gate)

- Semua endpoint MVP tersedia, terdokumentasi, dan mengikuti envelope/error format canonical.
- Data jobs tidak duplikat untuk kombinasi `source + original_job_id`.
- User premium aktif menerima notifikasi sesuai preference, non-premium tidak.
- Alur payment -> premium berjalan otomatis via webhook idempotent.
- Ownership boundary terimplementasi (`JWT.sub` untuk user-scoped endpoint).
- Audit trail transaksi + webhook dapat ditelusuri via `request_id` dan tabel audit.
- Link antar dokumen backend valid.
