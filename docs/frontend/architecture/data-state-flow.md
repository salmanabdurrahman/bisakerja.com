# Data Fetching, Cache Strategy, and State Flow

Dokumen ini menjelaskan alur data, strategi cache, dan pendekatan state management untuk App Router.

## 1) Data Ownership Principles

1. **Server state adalah source of truth** untuk data dari API.
2. **Client state hanya untuk UI state** (modal, filter draft, tab, toast, dsb).
3. **URL state untuk state yang shareable** (search query, sort, pagination).
4. Hindari duplikasi state yang sama di server + client tanpa kebutuhan nyata.

## 1.1) Canonical Backend-Aligned State Slices

| State slice | Source of truth | Dipakai oleh | Catatan implementasi |
|---|---|---|---|
| `auth_session` | `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh` | protected routes + API interceptor | Wajib single-flight refresh saat `401` paralel |
| `user_profile` | `GET /api/v1/auth/me` | account/profile/navigation | Jangan override dari cache lokal tanpa verifikasi session |
| `subscription_entitlement` | `GET /api/v1/billing/status -> subscription_state` | pricing, preferences gating, profile badge, upsell | Source of truth entitlement premium |
| `billing_transactions` | `GET /api/v1/billing/transactions` | account history | Read-only untuk histori, bukan entitlement |
| `job_search_query` | URL params (`q`, `location`, `salary_min`, `sort`, `page`, `limit`, `source`) | jobs list + shareable URL | Saat filter utama berubah, reset `page=1` |
| `preferences_form` | `GET /api/v1/preferences` + draft client untuk perubahan belum tersimpan | account/preferences | Saat `GET` gagal sementara, gunakan cache/draft lokal + retry |

## 2) Read Flow (Server-first)

1. Request masuk ke route `app/*` (Server Component).
2. Route memanggil `services/*` (typed API client) di server.
3. Data diproses ringan di server (mapping/formatting presentasi).
4. Hasil dirender langsung di server; Client Component menerima props minimum untuk interaksi.

## 3) Write Flow (Mutation)

1. User submit action dari Client Component (form/button interaction).
2. Mutation dikirim melalui server boundary (Server Action atau route handler sesuai kebutuhan).
3. Validasi input dilakukan sebelum request final ke backend.
4. Jika sukses, trigger revalidation (`revalidatePath`/`revalidateTag`) pada route terkait.
5. UI menampilkan state transisi: `submitting -> success` atau `submitting -> error`.
6. Jika gagal `401`, jalankan refresh flow (sekali) lalu retry request; bila refresh gagal, arahkan ke login.

## 4) Cache Policy Matrix

Gunakan cache strategy secara eksplisit (jangan implicit default):

| Jenis data | Contoh | Policy | Revalidate | Catatan |
|---|---|---|---|---|
| Public semi-static content | landing/jobs highlight | `force-cache` | 300 detik | Cocok untuk konten sering dibaca, jarang berubah |
| Reference data | job categories, locations | `force-cache` | 3600 detik | Invalidasi berbasis tag saat data master berubah |
| User-scoped private data | profile, preferences, billing status | `no-store` | - | Wajib fresh karena sensitif per user |
| Search result dengan query dinamis | `/jobs?query=...` | `no-store` | - | Hindari stale data per query/filter |
| Session/UI config ringan | non-sensitive feature config | `force-cache` | 60 detik | Bisa dipadukan dengan fallback aman |

## 5) Revalidation Rules

- Setelah mutation sukses, invalidasi route/tag terkait agar data server state tetap konsisten.
- Jangan lakukan refetch manual berlapis jika revalidation sudah cukup.
- Hindari waterfall fetching: tarik data utama di parent server component.

## 6) State Management Approach

### A. Local Component State (default)

Gunakan untuk state sementara:

- input value,
- open/close modal,
- pagination UI lokal,
- loading state komponen kecil.

### B. URL State

Gunakan untuk state yang harus bisa di-share/bookmark:

- keyword pencarian,
- filter,
- sorting,
- page number.

### C. Shared Client State (`store/*`)

Gunakan terbatas untuk state lintas fitur yang bukan source data backend, misalnya:

- user UI preferences (theme, density, table column visibility),
- ephemeral session UI (dismissed banner, wizard step).

### D. Server State

Tetap di server render layer + cache policy Next.js. Jangan dipindahkan ke global client store sebagai default.

## 7) Error Handling Contract

- Semua error API dipetakan ke kategori UI konsisten: validation, unauthorized, forbidden, not-found, unexpected.
- Halaman/section wajib memiliki state: loading, empty, error.
- Error message ke user harus aman (tanpa bocor detail internal).

### HTTP-to-UI Mapping Minimum

| HTTP | UI category | Default action |
|---|---|---|
| `400` | validation error | tampilkan pesan field/form dan aksi perbaikan input |
| `401` | unauthorized/session expired | clear session jika refresh gagal, redirect login |
| `403` | forbidden | tampilkan halaman/section akses ditolak |
| `404` | not found | tampilkan state not-found + CTA kembali |
| `409` | conflict | sinkron ulang state dari endpoint read model |
| `429` | rate limited | tampilkan cooldown + retry terkontrol |
| `5xx` / `503` | retryable backend error | tampilkan retry + fallback aman tanpa crash |

## 8) Testing Alignment (CI)

Selaras dengan:

- [CI Quality Gates](../../standards/ci-quality-gates.md)
- [Testing Strategy](../../standards/testing-strategy.md)

Minimum untuk perubahan data/state flow:

- Unit test untuk util mapping/formatter.
- Component test untuk loading/error/empty state.
- E2E untuk journey kritikal: login/register, search jobs, set preferences, checkout premium.

## 9) Output Implementasi Minimum per Perubahan Data Flow

- Update typed API contract (request/response mapper) di layer `services/*`.
- Update state diagram/transisi (doc atau test naming) untuk alur yang berubah.
- Tambahkan/ubah test untuk minimal 1 happy path + 1 failure path per mutation kritikal.
