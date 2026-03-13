# ADR Guidelines (Architecture Decision Records)

ADR dipakai untuk mencatat keputusan teknis penting agar jejak reasoning tidak hilang.

## 1) Kapan Wajib Buat ADR

- Mengubah arsitektur utama atau boundary module.
- Menambah dependency inti yang berdampak luas.
- Mengubah contract data/API secara signifikan.
- Memilih trade-off teknis yang tidak trivial.

## 2) Lokasi dan Penamaan

- Simpan di `docs/adr/` (direkomendasikan saat implementasi dimulai).
- Format nama: `YYYYMMDD-short-title.md`.

Contoh:
- `20260313-adopt-mayar-webhook-idempotency.md`

## 3) Template ADR

Gunakan template berikut:

```md
# ADR: <Title>

## Status
Proposed | Accepted | Deprecated | Superseded

## Context
Masalah/constraint yang melatarbelakangi keputusan.

## Decision
Keputusan yang diambil.

## Alternatives Considered
Opsi lain dan alasannya tidak dipilih.

## Consequences
Dampak positif/negatif, risiko, dan mitigasi.

## Rollout Plan
Langkah penerapan dan validasi.
```

## 4) Lifecycle

1. Author membuat ADR saat proposal perubahan.
2. Reviewer utama memberi feedback.
3. Setelah disetujui, status jadi `Accepted`.
4. Jika diganti, tandai `Superseded` dan referensikan ADR pengganti.
