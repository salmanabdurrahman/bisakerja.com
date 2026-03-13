# Monorepo Project Layout (Golang + Next.js)

Dokumen ini memberi contoh struktur folder lengkap untuk satu repo yang berisi backend Go dan frontend Next.js, dengan acuan utama `golang-standards/project-layout`.

## 1) Tujuan Desain

- Modular dan scalable untuk tim kecil hingga menengah.
- Memisahkan concern aplikasi (`apps`) dan shared package (`packages`).
- Menjaga backend tetap idiomatik Go (`cmd`, `internal`, `pkg`, `migrations`, `test`).
- Menjaga frontend tetap idiomatik Next.js (`app`, `components`, `features`, `lib`).

## 2) Struktur Folder yang Direkomendasikan

```text
bisakerja/
├── apps/
│   ├── api/                                  # Go backend app (module terpisah)
│   │   ├── cmd/                              # golang-standards: executable entrypoints
│   │   │   ├── api/                          # HTTP API server
│   │   │   │   └── main.go
│   │   │   ├── scraper/                      # scraper worker
│   │   │   │   └── main.go
│   │   │   ├── notifier/                     # matcher/notifier worker
│   │   │   │   └── main.go
│   │   │   ├── billing-worker/               # retry/reconciliation worker
│   │   │   │   └── main.go
│   │   │   └── migrate/                      # migration runner
│   │   │       └── main.go
│   │   ├── internal/                         # golang-standards: private app code
│   │   │   ├── app/                          # usecase/application layer
│   │   │   │   ├── auth/
│   │   │   │   ├── jobs/
│   │   │   │   ├── preferences/
│   │   │   │   ├── billing/
│   │   │   │   ├── notifications/
│   │   │   │   └── admin/
│   │   │   ├── domain/                       # entities, value objects, contracts
│   │   │   │   ├── user/
│   │   │   │   ├── job/
│   │   │   │   ├── transaction/
│   │   │   │   └── notification/
│   │   │   ├── adapter/                      # adapter/infrastructure implementations
│   │   │   │   ├── http/
│   │   │   │   │   ├── handler/
│   │   │   │   │   ├── middleware/
│   │   │   │   │   └── router/
│   │   │   │   ├── persistence/
│   │   │   │   │   ├── postgres/
│   │   │   │   │   └── redis/
│   │   │   │   ├── queue/
│   │   │   │   ├── mail/
│   │   │   │   └── mayar/
│   │   │   ├── platform/                     # bootstrap/config/wiring
│   │   │   │   ├── config/
│   │   │   │   ├── logger/
│   │   │   │   ├── observability/
│   │   │   │   └── server/
│   │   │   └── jobsched/                     # cron/scheduler orchestration
│   │   ├── pkg/                              # golang-standards: reusable public libs
│   │   │   ├── response/
│   │   │   ├── pagination/
│   │   │   ├── validator/
│   │   │   ├── authjwt/
│   │   │   └── errcode/
│   │   ├── api/                              # OpenAPI/contract docs backend
│   │   │   └── openapi.yaml
│   │   ├── configs/                          # golang-standards: config templates
│   │   │   ├── config.example.yaml
│   │   │   └── sources.example.yaml
│   │   ├── init/                             # golang-standards: process manager init
│   │   │   └── systemd/
│   │   ├── migrations/
│   │   │   ├── 000001_init.up.sql
│   │   │   └── 000001_init.down.sql
│   │   ├── scripts/
│   │   │   ├── seed.sh
│   │   │   └── dev.sh
│   │   ├── build/
│   │   │   ├── package/
│   │   │   └── ci/
│   │   ├── deployments/
│   │   │   ├── docker/
│   │   │   └── k8s/
│   │   ├── test/                             # golang-standards: integration/e2e tests
│   │   │   ├── integration/
│   │   │   └── e2e/
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── Makefile
│   │
│   └── web/                                  # Next.js app
│       ├── app/                              # App Router pages/layouts
│       ├── components/                       # shared UI components
│       ├── features/                         # domain-based UI modules
│       │   ├── auth/
│       │   ├── jobs/
│       │   ├── billing/
│       │   ├── notifications/
│       │   └── admin/
│       ├── services/                         # API clients (Bisakerja API)
│       ├── lib/                              # utility/helper
│       ├── hooks/
│       ├── store/                            # state management
│       ├── styles/
│       ├── public/
│       ├── tests/
│       │   ├── unit/
│       │   └── e2e/
│       ├── next.config.ts
│       ├── package.json
│       └── tsconfig.json
│
├── packages/                                 # shared packages lint/type/ui
│   ├── ui/                                   # optional shared UI kit
│   ├── types/                                # shared TypeScript types
│   ├── eslint-config/
│   └── tsconfig/
│
├── contracts/
│   ├── openapi/                              # generated/curated API specs
│   └── events/                               # webhook/event schemas
│
├── deployments/                              # environment-level deploy config
│   ├── docker-compose/
│   ├── staging/
│   └── production/
│
├── scripts/                                  # workspace scripts
│   ├── setup-dev.sh
│   ├── lint-all.sh
│   └── test-all.sh
│
├── tools/                                    # dev tools configs
│   ├── golangci-lint/
│   ├── sqlc/
│   └── pre-commit/
│
├── docs/
│   ├── prd/
│   ├── api/
│   ├── features/
│   ├── architecture/
│   ├── frontend/
│   ├── flows/
│   ├── phases/
│   ├── standards/
│   └── README.md
│
├── .github/
│   └── workflows/
│       ├── ci-api.yml
│       ├── ci-web.yml
│       └── release.yml
│
├── .editorconfig
├── .gitignore
├── package.json                              # workspace manager (pnpm/npm/yarn)
├── pnpm-workspace.yaml                       # jika menggunakan pnpm
└── README.md
```

## 3) Mapping ke `golang-standards/project-layout`

| Rekomendasi Golang Standards | Lokasi di Monorepo Ini |
|---|---|
| `/cmd` | `apps/api/cmd` |
| `/internal` | `apps/api/internal` |
| `/pkg` | `apps/api/pkg` |
| `/api` | `apps/api/api` + `contracts/openapi` |
| `/configs` | `apps/api/configs` |
| `/init` | `apps/api/init` |
| `/scripts` | `apps/api/scripts` + root `scripts` |
| `/build` | `apps/api/build` |
| `/deployments` | `apps/api/deployments` + root `deployments` |
| `/test` | `apps/api/test` |
| `/tools` | root `tools` |

## 4) Aturan Modularitas (Penting)

1. Kode domain backend hanya boleh diakses via `internal/app` contract.
2. `internal/adapter` tidak boleh dipakai langsung oleh layer frontend.
3. `pkg` hanya untuk util reusable yang tidak spesifik bisnis.
4. Frontend (`apps/web/features/*`) mengikuti domain backend agar traceability mudah.
5. Semua contract API disimpan di `contracts/openapi` dan dijadikan sumber kebenaran lintas app.

## 5) Rekomendasi Naming

- Service/package backend: `snake_case` untuk nama folder teknis (`billing_worker` bisa dipertimbangkan), namun konsisten.
- Package Go: singkat, lower-case, tanpa underscore jika memungkinkan.
- Folder Next.js features: domain-first (`features/jobs`, `features/billing`).

## 6) Minimal Workspace Commands

- `pnpm --filter web dev` untuk frontend.
- `make -C apps/api run-api` untuk API.
- `make -C apps/api run-scraper` untuk scraper.
- `make -C apps/api test` untuk backend test.
- `pnpm --filter web test` untuk frontend test.

## 7) Urutan Implementasi Struktur (Suggested)

1. Bentuk skeleton folder `apps/api` sesuai layout Go.
2. Bentuk skeleton folder `apps/web` (Next.js).
3. Aktifkan CI terpisah untuk backend/frontend.
4. Sinkronkan contract API (`contracts/openapi`) sebelum implementasi fitur.
5. Lakukan implementasi per phase pada dokumen roadmap.

## 8) Recommended Config Files (Agar Best Practices Terenforce)

Di root repo:

- `.golangci.yml` (aturan lint backend Go)
- `apps/web/eslint.config.js` atau `.eslintrc.*` (lint frontend)
- `.prettierrc` (formatting konsisten frontend/docs)
- `.editorconfig` (indentation universal)
- `.markdownlint.yml` (opsional konsistensi docs)
- `.husky/` atau `lefthook.yml` (pre-commit hooks)

Di CI:

- workflow backend (`ci-api.yml`) menjalankan lint + test + build + security scan.
- workflow frontend (`ci-web.yml`) menjalankan lint + type-check + test + build.
- workflow docs (opsional) untuk validasi link markdown.

## 9) Enforcement Points

- Standar coding: `docs/standards/go-coding-standards.md` dan `docs/standards/nextjs-coding-standards.md`.
- Standar komentar/docstring: `docs/standards/comments-and-docstrings.md`.
- Quality gates: `docs/standards/ci-quality-gates.md`.
- Audit implementasi phase: `docs/phases/implementation-checklist.md`.
