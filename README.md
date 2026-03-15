# Bisakerja

<p align="left">
  <img src="https://img.shields.io/badge/Go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Next.js-black?style=for-the-badge&logo=next.js&logoColor=white" alt="Next.js">
  <img src="https://img.shields.io/badge/PostgreSQL-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Tailwind_CSS-%2338B2AC.svg?style=for-the-badge&logo=tailwind-css&logoColor=white" alt="Tailwind CSS">
  <img src="https://img.shields.io/badge/pnpm-%234a4a4a.svg?style=for-the-badge&logo=pnpm&logoColor=f69220" alt="PNPM">
</p>

**Bisakerja** adalah platform cerdas agregasi lowongan pekerjaan yang memungkinkan pengguna untuk mencari, melacak, dan menganalisis kecocokan profil mereka terhadap berbagai lowongan dari beragam sumber portal kerja terkemuka. Dibangun dalam arsitektur Monorepo, platform ini memisahkan secara modular antara Backend (Golang) dan Frontend (Next.js).

## Fitur Utama

- **Job Aggregation & Search**: Agregasi otomatis (scraping) lowongan pekerjaan dari berbagai sumber dengan filter pencarian canggih (kata kunci, lokasi, remote/onsite, gaji).
- **AI-Powered Tools**: Asisten cerdas untuk mencari pekerjaan (_Search Assistant_), merangkum kecocokan profil pengguna dengan deskripsi lowongan (_Job-Fit Summary_), hingga membuat draf _Cover Letter_.
- **Application Tracker**: Kanvas interaktif bagi pengguna untuk mem-bookmark lowongan serta melacak status setiap lamaran pekerjaan (Applied, Interview, Offer, Rejected, dsb.).
- **Subscription & Billing**: Sistem langganan (Free vs Premium) dengan kuota API berbasis limit harian dan integrasi mulus menggunakan payment gateway **Midtrans**.
- **Smart Notifications**: Preferensi notifikasi pengguna berbasis kata kunci spesifik terkait lowongan pekerjaan terbaru yang masuk.

## Teknologi yang Digunakan

Projek ini disusun menggunakan struktur monorepo (`pnpm workspaces`), terdiri dari:

### Frontend (`apps/web`)

- **Framework**: [Next.js 15+](https://nextjs.org/) (App Router)
- **Styling**: [Tailwind CSS](https://tailwindcss.com/)
- **UI Components**: Radix UI / Shadcn UI patterns
- **Testing**: [Vitest](https://vitest.dev/) & React Testing Library
- **State Management**: React Hooks standar dipadukan dengan Server/Client Components.

### Backend (`apps/api`)

- **Language**: [Go (Golang)](https://go.dev/) 1.23+
- **Database**: PostgreSQL (relasional utama)
- **Migrations**: `golang-migrate` / standard Go migration tooling
- **Background Workers**: Internal scheduler / worker Go untuk scraping dan sinkronisasi status billing
- **Integrations**: Midtrans Snap API (Payments), OpenAI-Compatible API (AI Tools).

## Struktur Folder Monorepo

```text
bisakerja/
├── apps/
│   ├── api/                 # Backend service (Go)
│   │   ├── cmd/api/         # Entrypoint HTTP server
│   │   ├── internal/        # Domain logic, handlers, services, repositories
│   │   ├── migrations/      # File migrasi SQL
│   │   └── Makefile         # Backend automation commands
│   └── web/                 # Frontend application (Next.js)
│       ├── src/app/         # Next.js App Router (pages & layouts)
│       ├── src/features/    # Domain-specific UI modules (auth, ai, tracker)
│       ├── src/components/  # Shared/reusable UI components
│       └── tests/           # Vitest unit & component tests
├── docs/                    # Dokumentasi lengkap (PRD, Arsitektur, Flow, API)
├── package.json             # Root monorepo workspace configuration
├── pnpm-workspace.yaml      # PNPM workspace definition
└── AGENTS.md                # System instructions/playbook untuk AI development
```

## Prasyarat (Prerequisites)

Sebelum mulai menjalankan di komputer lokal, pastikan Anda telah menginstal:

1. [**Node.js**](https://nodejs.org/en/) (v18+ atau v20 LTS direkomendasikan)
2. [**pnpm**](https://pnpm.io/) (Package manager, `npm install -g pnpm`)
3. [**Go**](https://go.dev/dl/) (versi 1.23+)
4. [**PostgreSQL**](https://www.postgresql.org/download/) (Akses ke database server yang aktif)
5. **GNU Make** (Untuk menjalankan perintah Makefile di backend)

## Setup Environment (`.env`)

Terdapat dua file `.env` yang perlu diatur: satu untuk Frontend dan satu untuk Backend.

### 1. Duplikasi `.env.example`

Buka terminal di root projek dan jalankan:

```bash
cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env.local
```

### 2. Konfigurasi Backend (`apps/api/.env`)

Buka file `apps/api/.env` dan atur utamanya pada bagian _Database_ dan _Kunci Eksternal_:

```env
# Sesuaikan URL koneksi PostgreSQL Anda
DATABASE_URL=postgres://postgres:password@localhost:5432/bisakerja?sslmode=disable

# Ubah JWT Secret
AUTH_JWT_SECRET=rahasia-super-aman-anda

# Midtrans (Dapatkan di Dashboard Midtrans Sandbox)
MIDTRANS_SERVER_KEY=SB-Mid-server-xxxx
MIDTRANS_CLIENT_KEY=SB-Mid-client-xxxx

# AI (Opsional untuk fitur AI)
AI_PROVIDER_API_KEY=sk-xxxxx
```

### 3. Konfigurasi Frontend (`apps/web/.env.local`)

Buka `apps/web/.env.local` dan pastikan Midtrans Client Key sejajar:

```env
# Sesuai dengan konfigurasi Next.js Rewrites (local dev)
API_ORIGIN=http://localhost:8080

# Kunci publik Midtrans (Sama dengan Backend MIDTRANS_CLIENT_KEY)
NEXT_PUBLIC_MIDTRANS_CLIENT_KEY=SB-Mid-client-xxxx
NEXT_PUBLIC_MIDTRANS_ENV=sandbox
```

## Cara Menjalankan Secara Lokal (Quick Start)

### 1. Install Dependencies

Di root direktori projek, jalankan:

```bash
pnpm install
```

### 2. Jalankan Migrasi Database Backend

Arahkan ke folder backend dan eksekusi Makefile untuk menjalankan migrasi skema tabel ke PostgreSQL:

```bash
make -C apps/api check-migrations
make -C apps/api migrate-up
```

### 3. Jalankan Service Backend

Untuk menjalankan server API:

```bash
make -C apps/api run-api
```

_Catatan: API secara default akan berjalan di `http://localhost:8080`._

### 4. Jalankan Service Frontend

Buka tab terminal baru (biarkan backend tetap berjalan), lalu jalankan:

```bash
pnpm --filter web dev
```

_Web client akan dapat diakses melalui `http://localhost:3000`._

## Quality Commands (Linting & Testing)

Kami menjaga kualitas kode menggunakan seperangkat skrip dari _root workspace_:

- **Merapikan Kode (Linting)**:
  ```bash
  pnpm lint
  ```
- **Pengecekan Tipe Data (Typecheck/Go Vet)**:
  ```bash
  pnpm typecheck
  ```
- **Menjalankan Unit/Integration Tests**:
  ```bash
  pnpm test
  ```
- **Melakukan Build Simulasi Production**:
  ```bash
  pnpm build
  ```

## Deployment (Opsional)

### Deployment Frontend (Next.js)

Frontend siap di-deploy secara mudah melalui [**Vercel**](https://vercel.com/):

1. Import repositori GitHub ini ke Vercel.
2. Atur **Root Directory** ke `apps/web`.
3. Masukkan _Environment Variables_ yang ada di `.env.local`.
4. Deploy.

### Deployment Backend (Go)

1. **Build Binary**: `cd apps/api && go build -o main ./cmd/api`
2. Backend dapat dijalankan sebagai _binary file_ mandiri atau dikontainerisasi menggunakan **Docker**.
3. Pastikan environment variabel tersuntikkan (contoh: di VM/VPS, Railway, Fly.io, atau AWS) beserta akses `DATABASE_URL` ke production PostgreSQL.
4. Gunakan Nginx atau Caddy sebagai reverse proxy jika diperlukan.

## Panduan Dokumentasi (Documentation)

Platform Bisakerja terdokumentasi dengan sangat terstruktur. Seluruh pedoman pengembangan, PRD (_Product Requirements Document_), dan standar arsitektur dapat ditemukan dalam direktori `docs/`.

Bagi **pengembang yang baru bergabung**, **WAJIB** membaca:

1. `docs/README.md` - Indeks dan peta jalan seluruh dokumentasi.
2. `docs/architecture/local-development-runbook.md` - Panduan lengkap _troubleshooting_ lingkungan lokal.
3. `docs/standards/` - Aturan format kode (_coding standards_) untuk Next.js dan Go.

_Dibuat dengan cinta untuk kemudahan mencari kerja._
