# App Router Structure & Module Boundaries

Dokumen ini mendefinisikan struktur standar `apps/web` dan aturan boundary antar modul.

## 1) App Router Structure (Recommended)

```text
apps/web/
├── app/
│   ├── (marketing)/
│   │   ├── layout.tsx
│   │   └── page.tsx
│   ├── (auth)/
│   │   ├── login/page.tsx
│   │   └── register/page.tsx
│   ├── (app)/
│   │   ├── dashboard/page.tsx
│   │   ├── jobs/page.tsx
│   │   ├── preferences/page.tsx
│   │   └── billing/page.tsx
│   ├── api/
│   │   └── */route.ts
│   ├── layout.tsx
│   ├── loading.tsx
│   ├── error.tsx
│   └── not-found.tsx
├── features/
│   ├── auth/
│   ├── jobs/
│   ├── preferences/
│   ├── billing/
│   └── profile/
├── components/
├── services/
├── store/
├── hooks/
└── lib/
```

## 2) Module Ownership & Responsibilities

| Modul | Tanggung jawab | Boleh mengimpor | Tidak boleh mengimpor |
|---|---|---|---|
| `app/*` | Route composition, metadata, segment layout | `features`, `components`, `services`, `lib` | Logic bisnis domain kompleks |
| `features/*` | UI + use-case per domain | `services`, `components`, `hooks`, `lib`, `store` | Import antar domain tanpa kontrak jelas |
| `components/*` | Reusable UI lintas domain | `lib` | `services` domain-specific, logic bisnis |
| `services/*` | Typed API client, request mapping | `lib` | Komponen UI/React hook presentasional |
| `store/*` | Shared client state minimal | `lib` | Data source utama server state |
| `lib/*` | Utility murni, formatter, constants | - | Ketergantungan ke `app`/`features` |

## 3) Import Direction Rules

Arah dependency yang diperbolehkan:

`app -> features -> (services | components | hooks | store | lib)`

`components -> lib`

`services -> lib`

Tidak diperbolehkan:

- `services -> features`
- `components -> features`
- `lib -> app/features/services`

## 4) Server vs Client Component Usage Policy

Selaras dengan [Next.js Coding Standards](../../standards/nextjs-coding-standards.md):

- **Default: Server Component** untuk page/layout/section yang berfokus data.
- Gunakan **Client Component** hanya saat membutuhkan:
  - event handlers (`onClick`, `onChange`),
  - state browser (`useState`, `useEffect`),
  - browser APIs (LocalStorage, IntersectionObserver, dll).
- Letakkan boundary `"use client"` sedekat mungkin ke leaf component untuk menekan ukuran bundle.
- Jangan impor modul server-only (secrets, private env, DB concern) ke Client Component.
- Route-level files (`page.tsx`, `layout.tsx`) tetap server kecuali ada kebutuhan interaktif yang tidak bisa dipindahkan.

## 5) Route Grouping Policy

- `(marketing)` untuk halaman publik dan SEO-first.
- `(auth)` untuk login/register/reset flow.
- `(app)` untuk area pengguna terautentikasi (dashboard, jobs, profile, billing).
- `app/api/*` hanya untuk kebutuhan endpoint internal/webhook frontend-specific, bukan menggantikan backend utama.

## 6) Boundary Review Checklist

Sebelum merge:

- [ ] Fitur baru ditempatkan di domain `features/*` yang benar.
- [ ] Tidak ada business rule kompleks di `app/*` atau `components/*`.
- [ ] Client boundary minimal dan beralasan jelas.
- [ ] Import direction tidak melanggar aturan di dokumen ini.
