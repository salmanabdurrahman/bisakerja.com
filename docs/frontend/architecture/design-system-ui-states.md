# Design System Tokens, Components, and UI States

Dokumen ini menetapkan baseline design system untuk konsistensi UI user-facing app.

## 1) Design System Principles

- Token-driven styling (bukan hardcoded value per komponen).
- Semantic naming (berbasis intent, bukan warna mentah).
- Komponen reusable diprioritaskan sebelum membuat variasi baru.
- Semua komponen interaktif harus aksesibel dan memiliki state lengkap.

## 2) Token Architecture

Gunakan 2 layer token:

1. **Core tokens**: raw values (warna, spacing, radius, shadow, font-size).
2. **Semantic tokens**: dipakai langsung komponen (`surface`, `text-muted`, `border-danger`, dst).

### Token Categories (minimum)

| Kategori | Contoh token | Tujuan |
|---|---|---|
| Color - surface/text | `surface-default`, `surface-elevated`, `text-primary`, `text-secondary` | Konsistensi tema dan kontras |
| Color - feedback | `feedback-success`, `feedback-warning`, `feedback-danger`, `feedback-info` | State message/status |
| Spacing | `space-1..space-8` (4px scale) | Ritme layout |
| Radius | `radius-sm`, `radius-md`, `radius-lg` | Konsistensi sudut komponen |
| Typography | `font-size-xs..xl`, `font-weight-regular/medium/semibold` | Hierarki konten |
| Elevation | `shadow-sm/md/lg` | Penanda layer visual |
| Motion | `duration-fast/base/slow`, `easing-standard` | Transisi konsisten |
| Focus | `focus-ring-color`, `focus-ring-width` | Aksesibilitas keyboard |

## 3) Component Layering

### A. Primitives (`components/*`)

Contoh: `Button`, `Input`, `Select`, `Checkbox`, `Badge`, `Card`, `Dialog`.

### B. Composite Components

Contoh: `SearchBar`, `FilterPanel`, `JobCard`, `PricingCard`, `EmptyState`, `ErrorState`.

### C. Feature Sections (`features/*`)

Komposisi domain-specific yang menggunakan primitive/composite, tidak membuat token baru tanpa review.

## 4) UI State Requirements

Setiap komponen async/interactive minimal punya state berikut:

| State | Wajib ada | Catatan |
|---|---|---|
| Idle/Default | ✅ | Tampilan normal |
| Hover/Focus/Active | ✅ | Wajib keyboard-visible focus |
| Loading | ✅ | Gunakan skeleton/spinner + label jelas |
| Empty | ✅ | Ada CTA lanjutan (mis. "Cari lowongan") |
| Error | ✅ | Pesan jelas + aksi retry bila memungkinkan |
| Success | ✅ (untuk mutation) | Feedback non-intrusif, bisa auto-dismiss |
| Disabled | ✅ (jika relevan) | Tetap kontras dan jelaskan alasan bila perlu |

## 5) Page-Level State Composition

Untuk halaman data-heavy (jobs, dashboard, billing):

- **Initial loading**: skeleton layout utama (hindari layout shift besar).
- **Partial loading**: hanya section yang berubah, bukan seluruh halaman.
- **Empty state**: jelaskan kondisi + langkah berikutnya.
- **Error state**: tampilkan fallback aman dan opsi retry/contact support.

## 6) Content and Feedback Consistency

- Gunakan tone pesan yang konsisten dan actionable.
- Validation error ditempatkan dekat field terkait.
- Gunakan `aria-live` untuk feedback dinamis (success/error toast).

## 7) Testing Requirements for UI States

Selaras dengan [Testing Strategy](../../standards/testing-strategy.md):

- Component test wajib mencakup state penting: loading, empty, error, success (jika mutation).
- E2E memverifikasi alur inti dan state transisi utama.
- Perubahan komponen shared harus menghindari breaking visual behavior lintas fitur.
