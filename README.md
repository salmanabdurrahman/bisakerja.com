# Bisakerja Monorepo

Monorepo implementasi Bisakerja untuk backend Go (`apps/api`) dan frontend Next.js (`apps/web`).

## Quick Start

1. Install workspace dependencies:

   ```bash
   pnpm install
   ```

2. Jalankan API:

   ```bash
   make -C apps/api run-api
   ```

3. Jalankan frontend:

   ```bash
   pnpm --filter web dev
   ```

## Quality Commands

- Lint: `pnpm lint`
- Type/static check: `pnpm typecheck`
- Test: `pnpm test`
- Build: `pnpm build`

Dokumentasi implementasi ada di [`docs/`](./docs/README.md).
