# Accessibility and Performance Baseline

Dokumen ini menetapkan baseline terukur untuk accessibility dan performance pada aplikasi Next.js user-facing.

## 1) Accessibility Baseline (WCAG 2.2 AA)

### Aturan Wajib

- Gunakan semantic HTML (`header`, `main`, `nav`, `button`, `form`, `label`, dst).
- Semua kontrol interaktif dapat diakses keyboard (Tab/Shift+Tab/Enter/Space/Escape).
- Focus indicator harus terlihat jelas di semua komponen interaktif.
- Form field wajib memiliki label yang terbaca screen reader.
- Gambar informatif memiliki `alt` yang bermakna; dekoratif diberi perlakuan yang tepat.

### Target Terukur

- **0 critical accessibility violations** pada halaman kritikal saat pemeriksaan otomatis.
- **100% komponen interaktif kritikal keyboard-operable** (auth, search, checkout, profile).
- **Kontras warna minimal WCAG AA** (normal text 4.5:1, large text 3:1).
- **Tidak ada focus trap** di dialog/modal flow utama.

## 2) Performance Baseline (Core Web Vitals)

### Target Terukur (p75, mobile)

- **LCP <= 2.5s**
- **INP <= 200ms**
- **CLS <= 0.1**
- **TTFB <= 800ms** untuk route dinamis utama

### Budget Tambahan

- JavaScript awal per route dijaga sekecil mungkin; hindari Client Component besar di root layout.
- Asset gambar hero/LCP harus teroptimasi (gunakan `next/image` jika sesuai).
- Hindari refetch dan duplicate fetch yang tidak perlu.

## 3) Engineering Policies to Meet Baseline

Selaras dengan [Next.js Coding Standards](../../standards/nextjs-coding-standards.md):

- Default ke Server Component untuk menekan JS bundle.
- Gunakan dynamic import untuk modul client yang berat.
- Definisikan cache strategy eksplisit (`force-cache`, `no-store`, `revalidate`).
- Wajib menyediakan loading/error/empty states untuk menjaga perceived performance.

## 4) CI & Testing Alignment

Selaras dengan:

- [CI Quality Gates](../../standards/ci-quality-gates.md)
- [Testing Strategy](../../standards/testing-strategy.md)

Gate minimum frontend yang harus tetap hijau:

1. Lint (`eslint`)
2. Type-check (`tsc --noEmit`)
3. Unit/component tests
4. Build check (`next build`)

Untuk perubahan yang memengaruhi aksesibilitas/performance:

- Tambahkan atau update component tests untuk state kritikal.
- Pastikan E2E journey minimal tetap lulus: login/register, search jobs, preferences, checkout premium.
- Jangan merge bila terjadi regresi signifikan tanpa mitigasi/justifikasi.

## 5) Pull Request Acceptance Checklist

- [ ] Perubahan tidak melanggar baseline accessibility.
- [ ] Perubahan tidak mendorong halaman kritikal melewati target Web Vitals.
- [ ] UI states (loading/empty/error/success) tetap lengkap.
- [ ] Dokumen arsitektur terkait sudah diupdate jika ada perubahan behavior.
