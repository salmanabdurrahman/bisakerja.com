# Frontend Architecture (Next.js User App)

Dokumen ini menjadi acuan arsitektur untuk aplikasi user-facing Next.js (`apps/web`) dengan App Router.

## Scope

- Struktur App Router dan batas modul.
- Kebijakan penggunaan Server vs Client Components.
- Data fetching, caching, dan state management.
- Design system token/component serta UI states wajib.
- Baseline accessibility dan performance yang terukur.

## Referensi Standar Wajib

- [Next.js Coding Standards](../../standards/nextjs-coding-standards.md)
- [CI Quality Gates](../../standards/ci-quality-gates.md)
- [Testing Strategy](../../standards/testing-strategy.md)

## Peta Dokumen

- [App Structure & Module Boundaries](./app-structure.md)
- [Data & State Flow](./data-state-flow.md)
- [Design System & UI States](./design-system-ui-states.md)
- [Accessibility & Performance Baseline](./accessibility-performance.md)

## Prinsip Arsitektur Inti

1. **Server-first rendering**: default ke Server Component untuk data-heavy route.
2. **Domain-first modules**: logic bisnis per domain di `features/*`.
3. **Explicit boundaries**: `app`, `features`, `components`, `services`, `store`, `lib` punya tanggung jawab jelas.
4. **Accessible by default**: semua fitur wajib keyboard-friendly dan semantic.
5. **Performance budgets**: semua halaman penting mengikuti target Web Vitals dan budget bundle.

## Definition of Done untuk Perubahan Arsitektur Frontend

Perubahan dianggap selesai jika:

- Dokumen architecture terkait di folder ini sudah diperbarui.
- Tetap selaras dengan standar di `docs/standards`.
- Quality gates frontend tetap terpenuhi: lint, type-check, unit/component test, dan build check.
