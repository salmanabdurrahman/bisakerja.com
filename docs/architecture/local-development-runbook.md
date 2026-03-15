# Local Development Runbook (API, Workers, Frontend)

Panduan ini adalah langkah end-to-end untuk menjalankan Bisakerja secara lokal: setup `.env`, migrasi database, API, scraper, notifier, billing worker, dan frontend.

## 1) Prasyarat

- PostgreSQL 16+
- Go 1.25+
- Node.js 20+ dan `pnpm` (workspace root)

## 2) Bootstrap Workspace

Jalankan dari root monorepo:

```bash
pnpm install
cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env.local
```

> Runtime backend otomatis mencoba load env dari `.env` lalu `apps/api/.env`. Nilai yang sudah diexport di shell tidak akan dioverride.

## 3) Konfigurasi Environment

## 3.1 Backend (`apps/api/.env`)

### Wajib untuk local minimum

| Variable | Contoh | Kegunaan |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/bisakerja?sslmode=disable` | koneksi PostgreSQL |
| `AUTH_JWT_SECRET` | `change-this-jwt-secret` | signing access/refresh token |
| `BILLING_WEBHOOK_TOKEN` | `change-this-webhook-token` | autentikasi inbound webhook Mayar |
| `BILLING_REDIRECT_ALLOWLIST` | `app.bisakerja.com,localhost:3000,127.0.0.1:3000,[::1]:3000` | host redirect checkout yang diizinkan |

### Wajib untuk flow checkout Mayar

| Variable | Contoh | Kegunaan |
|---|---|---|
| `MAYAR_API_KEY` | `...` | kredensial API Mayar |
| `MAYAR_BASE_URL` | `https://api.mayar.id/hl/v1` (prod) / `https://api.mayar.club/hl/v1` (sandbox) | endpoint base Mayar |

### Wajib jika fitur AI dipakai

| Variable | Contoh | Kegunaan |
|---|---|---|
| `AI_PROVIDER_API_KEY` | `...` | API key provider OpenAI-compatible |
| `AI_PROVIDER_BASE_URL` | `https://api.openai.com/v1` | base URL provider (bisa custom) |
| `AI_PROVIDER_MODEL_DEFAULT` | `gpt-4.1-mini` | model default |

### Opsional tapi direkomendasikan

- `SCRAPER_KEYWORDS`, `SCRAPER_PAGE_SIZE`, `SCRAPER_MAX_PAGES`
- `JOBSTREET_BEARER_TOKEN`, `JOBSTREET_COOKIE`, `JOBSTREET_EC_SESSION_ID`, `JOBSTREET_EC_VISITOR_ID`, `GLINTS_COOKIE`
- tuning pool DB: `DATABASE_MAX_OPEN_CONNS`, `DATABASE_MIN_OPEN_CONNS`, `DATABASE_MAX_CONN_LIFETIME`, `DATABASE_MAX_CONN_IDLE_TIME`

## 3.2 Frontend (`apps/web/.env.local`)

| Variable | Default | Kegunaan |
|---|---|---|
| `NEXT_PUBLIC_API_BASE_URL` | `/api/v1` | base path API di browser (same-origin) |
| `API_ORIGIN` | `http://localhost:8080` | origin upstream API untuk rewrite/SSR |

## 4) Migrasi Database

Pastikan PostgreSQL sudah jalan dan database tujuan ada, lalu:

```bash
make -C apps/api check-migrations
make -C apps/api migrate-up
```

Rollback satu arah (opsional):

```bash
make -C apps/api migrate-down
```

## 5) Menjalankan Service

Gunakan terminal terpisah untuk tiap proses.

### Terminal A — API

```bash
make -C apps/api run-api
```

### Terminal B — Scraper Worker

```bash
make -C apps/api run-scraper
```

### Terminal C — Notifier Worker

```bash
make -C apps/api run-notifier
```

### Terminal D — Billing Reconciliation Worker

```bash
make -C apps/api run-billing-worker
```

### Terminal E — Frontend

```bash
pnpm --filter web dev
```

## 6) Urutan Startup yang Direkomendasikan

1. PostgreSQL
2. `check-migrations` + `migrate-up`
3. API
4. Frontend
5. Scraper + notifier + billing worker

## 7) Smoke Check Cepat

Backend health:

```bash
curl -sS http://localhost:8080/healthz
```

API via frontend rewrite:

```bash
curl -sS http://localhost:3000/api/v1/healthz
```

## 8) Troubleshooting Checkout (`Invalid checkout request`)

Jika UI menampilkan error ini, cek response API (`errors[0].code`):

- `INVALID_REDIRECT_URL`
  - Pastikan host redirect terdaftar di `BILLING_REDIRECT_ALLOWLIST`.
  - Untuk local dev, `http` hanya diizinkan untuk `localhost`, `127.0.0.1`, atau `::1` (dan tetap harus ada di allowlist).
  - Contoh valid local redirect: `http://localhost:3000/billing/success`.
- `INVALID_PLAN_CODE`
  - Pastikan request memakai `plan_code=pro_monthly`.
- `MAYAR_RATE_LIMITED` / `SERVICE_UNAVAILABLE`
  - Retry beberapa saat; cek kredensial/key Mayar dan konektivitas.

## 9) Quality Gate Lokal

Backend:

```bash
make -C apps/api lint test build check-migrations
```

Frontend:

```bash
pnpm --filter web lint
pnpm --filter web build
```

Full monorepo:

```bash
pnpm lint
pnpm test
pnpm build
```

## 10) Referensi Integrasi Mayar

- Dokumentasi resmi: `https://docs.mayar.id/api-reference`
- Mapping internal Bisakerja: [`../api/mayar-headless.md`](../api/mayar-headless.md)
