# Bisakerja Monorepo

Monorepo implementasi Bisakerja untuk backend Go (`apps/api`) dan frontend Next.js (`apps/web`).

## Quick Start

1. Install workspace dependencies:

   ```bash
   pnpm install
   ```

2. Copy env examples:

   ```bash
   cp apps/api/.env.example apps/api/.env
   cp apps/web/.env.example apps/web/.env.local
   ```

3. Jalankan migrasi backend:

   ```bash
   make -C apps/api check-migrations
   make -C apps/api migrate-up
   ```

4. Jalankan API:

   ```bash
   make -C apps/api run-api
   ```

5. Jalankan frontend:

   ```bash
   pnpm --filter web dev
   ```

Panduan lengkap (API + scraper + notifier + billing worker + troubleshooting checkout) ada di [`docs/architecture/local-development-runbook.md`](./docs/architecture/local-development-runbook.md).

## Quality Commands

- Lint: `pnpm lint`
- Type/static check: `pnpm typecheck`
- Test: `pnpm test`
- Build: `pnpm build`

Dokumentasi implementasi ada di [`docs/`](./docs/README.md).
