# Product Requirements Document

# Bisakerja Frontend (User-Facing Next.js App)

Version: **1.1 (Implementation Kickoff Ready)**
Owner: **Bisakerja Product & Frontend Team**
Status: **Approved for Implementation**

## 1. Ringkasan Produk Frontend

Bisakerja frontend adalah aplikasi web Next.js untuk user akhir dalam menemukan lowongan relevan, mengelola akun, dan mengatur preferensi notifikasi secara cepat dan jelas.

Frontend mencakup:

- Discovery lowongan (browse, search, filter, detail)
- Auth dan manajemen sesi user
- Pengelolaan preference dan status subscription
- Pengalaman antarmuka responsif untuk desktop dan mobile web

## 2. Tujuan UX/Product

- Mempercepat waktu dari landing page ke lowongan relevan.
- Menjaga alur pencarian dan apply intent tetap sederhana dan fokus.
- Mengurangi friksi pada onboarding (register/login + set preference).
- Menampilkan nilai paket premium dan status subscription secara transparan.
- Menjaga konsistensi UI lintas halaman utama dan state (loading/empty/error).

## 3. Scope MVP

### In Scope

- Halaman landing/home untuk entry discovery.
- Halaman daftar lowongan dengan search, filter dasar, sort, dan pagination.
- Halaman detail lowongan.
- Halaman auth (register/login/logout) dan profil dasar user.
- Halaman preference notifikasi dan minat lowongan.
- Halaman pricing/subscription + status transaksi dari sisi frontend.
- State UX dasar: loading, empty state, validation error, unauthorized, retryable error, dan feedback sukses.

### Out of Scope

- Aplikasi native mobile (iOS/Android).
- Personalisasi AI tingkat lanjut pada UI.
- Multi-language/internationalization penuh.
- Offline-first/PWA advanced capabilities.
- Real-time collaboration/chat antar user.

### 3.1 Konvensi Terminologi & Kontrak (Canonical)

Untuk menghindari mismatch implementasi, frontend mengikuti kontrak backend berikut:

1. **Konvensi path API**
   - Runtime path frontend: `/api/v1/*` (contoh: `/api/v1/auth/login`).
   - Dokumen API backend menuliskan resource path inti tanpa prefix (`/auth/login`, `/jobs`, dst).
2. **Envelope API**
   - Success: `meta` + `data`
   - Error: `meta` + `errors` (jika ada field-level issue)
3. **Subscription state (source of truth UI entitlement)**
   - `free`
   - `pending_payment`
   - `premium_active`
   - `premium_expired`
   - Sumber utama: `GET /api/v1/billing/status -> data.subscription_state`.
4. **Transaction/payment status**
   - Internal transaksi (`GET /api/v1/billing/transactions`): `pending`, `reminder`, `success`, `failed`
   - Event gateway/webhook dapat membawa status seperti `paid`/`reminder`.
   - Frontend tidak menentukan entitlement dari event mentah; entitlement tetap dari `subscription_state`.

## 4. Rangkuman Halaman Utama (High Level)

- `/` — Landing page, value proposition, CTA ke pencarian lowongan.
- `/jobs` — Listing lowongan + search/filter/sort/pagination.
- `/jobs/[id-or-slug]` — Detail lowongan dan metadata penting.
- `/auth/login` & `/auth/register` — Alur autentikasi user.
- `/account` — Ringkasan profil dan status akun.
- `/account/preferences` — Pengaturan preference pencarian/notifikasi.
- `/pricing` & `/account/subscription` — Paket premium dan status subscription.

### 4.1 Tanggung Jawab Implementasi per Journey

| Journey | Tanggung jawab frontend | Kontrak backend minimum | Output implementasi minimum |
|---|---|---|---|
| Auth & Session | Form auth, guard route protected, refresh token single-flight, redirect aman | `/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/refresh`, `/api/v1/auth/me` | halaman login/register, session provider, HTTP interceptor, test flow auth |
| Job Discovery | URL-driven search/filter/pagination, detail rendering, state loading/empty/error | `/api/v1/jobs`, `/api/v1/jobs/:id` | halaman jobs list/detail, query-state sync util, test search+detail |
| Preferences | Validasi form, normalisasi input, draft recovery saat 401/network issue | `GET /api/v1/preferences`, `PUT /api/v1/preferences` | halaman preferences, form schema validator, test validasi + submit |
| Premium Upgrade | CTA/gating per state, checkout initiation, verifikasi status pasca redirect | `/api/v1/billing/checkout-session`, `/api/v1/billing/status`, `/api/v1/billing/transactions` | halaman pricing/subscription, checkout handler, test upgrade initiation |
| Profile & Account | Menampilkan profile + subscription badge konsisten, logout bersih | `/api/v1/auth/me`, `/api/v1/billing/status` | halaman account, profile summary section, test state precedence |

## 5. Non-Functional Requirements Frontend (Ringkas)

### Accessibility

- Halaman dan komponen utama mengikuti prinsip WCAG 2.1 AA.
- Navigasi keyboard dan focus state wajib tersedia di flow inti.
- Kontras warna, label form, dan semantic HTML harus konsisten.

### Performance

- Core Web Vitals pada halaman utama ditargetkan berada pada kategori “Good”.
- Gunakan optimasi Next.js (SSR/SSG sesuai konteks, code-splitting, image optimization).
- Hindari bundle berlebih pada halaman kritikal (landing, listing, detail).

### Reliability

- UI tetap graceful saat API error/timeout (fallback message + retry path).
- Error boundaries dan penanganan state async harus konsisten.
- Deploy frontend harus menjaga kompatibilitas terhadap perubahan API versioned.
- Ketika backend mengembalikan `401`, `429`, `5xx`, UI harus memberi aksi lanjutan yang jelas (reauth/retry/wait-and-retry).

### Security (Frontend Side)

- Tidak menyimpan token sensitif di localStorage.
- Sanitasi/rendering konten dinamis untuk meminimalkan risiko XSS.
- Gunakan secure cookie/session handling pattern sesuai kontrak auth backend.
- Lindungi form penting dari misuse (validasi input + UX anti-abuse dasar).

## 6. Daftar Dokumen Detail Frontend (Modular Links)

PRD ini sengaja dipadatkan. Spesifikasi frontend detail dipisahkan ke dokumen modular.

### 6.1 Frontend Product & UX

- Frontend docs hub: [`../frontend/README.md`](../frontend/README.md)
- Feature specs hub: [`../frontend/features/README.md`](../frontend/features/README.md)

### 6.2 Frontend Feature Specs

- Auth & Session: [`../frontend/features/auth-session.md`](../frontend/features/auth-session.md)
- Job Discovery: [`../frontend/features/job-discovery.md`](../frontend/features/job-discovery.md)
- Premium Upgrade: [`../frontend/features/premium-upgrade.md`](../frontend/features/premium-upgrade.md)
- Preferences & Notifications: [`../frontend/features/preferences-notifications.md`](../frontend/features/preferences-notifications.md)
- Profile & Account: [`../frontend/features/profile-account.md`](../frontend/features/profile-account.md)

### 6.3 Frontend Architecture Specs

- Architecture hub: [`../frontend/architecture/README.md`](../frontend/architecture/README.md)
- App Structure & Module Boundaries: [`../frontend/architecture/app-structure.md`](../frontend/architecture/app-structure.md)
- Data & State Flow: [`../frontend/architecture/data-state-flow.md`](../frontend/architecture/data-state-flow.md)
- Design System & UI States: [`../frontend/architecture/design-system-ui-states.md`](../frontend/architecture/design-system-ui-states.md)
- Accessibility & Performance Baseline: [`../frontend/architecture/accessibility-performance.md`](../frontend/architecture/accessibility-performance.md)

### 6.4 Frontend Flow Specs

- Flows hub: [`../frontend/flows/README.md`](../frontend/flows/README.md)
- Discovery Flow: [`../frontend/flows/discovery-flow.md`](../frontend/flows/discovery-flow.md)
- Upgrade Billing Flow: [`../frontend/flows/upgrade-billing-flow.md`](../frontend/flows/upgrade-billing-flow.md)

### 6.5 Frontend Delivery & Traceability

- Phases hub: [`../frontend/phases/README.md`](../frontend/phases/README.md)
- Implementation kickoff guide: [`../phases/implementation-kickoff.md`](../phases/implementation-kickoff.md)
- Implementation roadmap: [`../frontend/phases/implementation-roadmap.md`](../frontend/phases/implementation-roadmap.md)
- Implementation checklist: [`../frontend/phases/implementation-checklist.md`](../frontend/phases/implementation-checklist.md)
- Frontend-backend traceability matrix: [`../frontend/traceability/frontend-backend-traceability.md`](../frontend/traceability/frontend-backend-traceability.md)
