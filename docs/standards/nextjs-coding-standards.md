# Next.js Coding Standards

Standar ini ditujukan untuk `apps/web` (Next.js App Router).

## 1) Struktur Modul Frontend

- Gunakan pendekatan domain-first di `features/*`.
- `components/` hanya untuk komponen reusable lintas domain.
- `services/` untuk API client, bukan untuk state UI.
- `lib/` untuk utility murni.

## 2) Server vs Client Components

- Default ke **Server Component**.
- Gunakan Client Component hanya jika butuh interaktivitas browser (`useState`, event handlers, browser APIs).
- Batasi boundary client agar bundle tetap kecil.

## 3) Data Fetching

- Fetch data utama di server component bila memungkinkan.
- Gunakan route handlers/backend API client yang typed.
- Hindari duplicate fetching pada parent-child tree.
- Tetapkan strategi cache secara eksplisit (`force-cache`, `no-store`, `revalidate`).

## 4) State Management

- State lokal sederhana: React state.
- State lintas fitur/global: gunakan store yang disepakati tim (misal Zustand/Redux) secara konsisten.
- Jangan taruh business rule kompleks di komponen presentasi.

## 5) UI dan Accessibility

- Komponen wajib keyboard-friendly dan aksesibel.
- Gunakan semantic HTML.
- Gambar pakai `next/image` jika sesuai.
- Loading/error/empty states harus eksplisit.
- Typography wajib konsisten lintas halaman:
  - gunakan utility scale yang sama untuk header/body/label (contoh: `bk-heading-*`, `bk-body*`, `bk-label`),
  - hindari hardcoded size acak per halaman jika sudah ada utility setara,
  - utamakan komponen shared (`PageHeader`, `CardTitle`, `StateCard`) sebagai source of truth.

## 6) API Integration

- Semua call ke backend lewat layer `services/`.
- Error handling dari API harus dipetakan ke UI message yang konsisten.
- Jangan hardcode endpoint; gunakan env config.
- Untuk browser, gunakan **path relatif** (`/api/v1/*`) agar request tetap same-origin dan tidak kena CORS lintas domain.
- Gunakan `NEXT_PUBLIC_API_BASE_URL` sebagai **path-only** (contoh: `/api/v1`), bukan full URL origin.
- Gunakan `API_ORIGIN` hanya untuk target upstream di rewrite/proxy Next.js dan kebutuhan SSR.
- Wajib normalisasi env URL/path di helper config (hapus trailing slash, paksa leading slash, dan ekstrak pathname dari URL absolut) agar misconfig tidak langsung memicu CORS.

## 7) Security

- Sanitasi input user sebelum render data dinamis.
- Jangan expose secrets di client bundle.
- Gunakan HTTP-only cookie bila memakai session/token cookie pattern.
- Validasi input tetap wajib di backend.

## 8) Performance

- Split komponen berat dengan dynamic import bila perlu.
- Hindari client state berlebih di root layout.
- Pantau web vitals untuk halaman kritikal.
- Wajib kirim metrik Web Vitals via observer client (`WebVitalsObserver`) ke endpoint same-origin (`/api/web-vitals`) agar bisa dipantau tanpa CORS/regress lintas origin.

## 9) Testing Minimum

- Unit test untuk util dan logic murni.
- Component test untuk UI behavior penting.
- E2E test untuk journey utama (auth, search, checkout, profile).

## 10) Anti-Pattern yang Harus Dihindari

- Fetching langsung dari banyak komponen tanpa koordinasi.
- Menaruh seluruh halaman sebagai Client Component tanpa alasan.
- Styling dan logic bisnis bercampur tidak terkontrol.
