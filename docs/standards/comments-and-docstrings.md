# Comments and Docstrings Standards

Tujuan utama komentar adalah memberi konteks, bukan mengulang kode.

## 1) Prinsip Umum

- Tulis komentar hanya jika menambah pemahaman.
- Hindari komentar yang menjelaskan hal yang sudah obvious.
- Komentar harus singkat, spesifik, dan up-to-date.

## 2) Go Doc Comment Standards

### Wajib untuk simbol exported

- Package comment (jika package penting/domain utama).
- Type, interface, const, var, function yang diekspor.

### Format

- Gunakan kalimat dimulai dengan nama simbol.
- Contoh:

```go
// CreateCheckoutSession creates a pending payment session in Mayar
// and stores the initial transaction record locally.
func CreateCheckoutSession(...) {...}
```

## 3) TypeScript/Next.js Doc Comment Standards

Gunakan TSDoc/JSDoc untuk fungsi exported dan util kritikal.

```ts
/**
 * Builds a normalized job search query from UI filters.
 * Throws when required fields are invalid.
 */
export function buildSearchQuery(input: SearchFilterInput): SearchQuery
```

## 4) Kapan Wajib Menambah Komentar

- Ada constraint bisnis non-obvious.
- Ada workaround teknis sementara.
- Ada behavior edge case yang mudah salah tafsir.
- Ada alasan trade-off arsitektur.

## 5) Kapan Jangan Menambah Komentar

- Hanya menjelaskan assignment/loop/if sederhana.
- Mengulang nama function tanpa konteks tambahan.

## 6) Good vs Bad Examples

### Bad

```go
// Increment i by one
i++
```

### Good

```go
// We intentionally retry only on 429/5xx because duplicate invoice creation
// can happen if we retry on all error classes.
```

## 7) Sinkronisasi Dokumentasi

- Saat behavior berubah, komentar/docstring terkait wajib diupdate di commit yang sama.
- Jika perubahan berdampak lintas modul, update juga dokumen pada `docs/`.
