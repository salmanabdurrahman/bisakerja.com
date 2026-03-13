# CI Quality Gates

Dokumen ini menetapkan gate CI yang bersifat **enforceable** untuk merge/release.

## 1) Gate Catalogue (Merge-Blocking)

| Gate | Check Wajib | Kriteria Lulus | Berlaku Untuk |
|---|---|---|---|
| G1 - Lint | Backend: `golangci-lint run ./...`<br>Frontend: `npm run lint` | Exit code `0`, tanpa lint error | Semua perubahan kode |
| G2 - Type/Static Analysis | Backend: `go vet ./...`<br>Frontend: `tsc --noEmit` | Exit code `0`, tanpa type/static analysis error | Semua perubahan kode |
| G3 - Unit/Component Test | Suite unit/component sesuai scope | Semua test wajib pass, tanpa test gagal | Semua perubahan kode |
| G4 - Integration/E2E Test | Integration/E2E sesuai trigger | Semua suite wajib pass | Perubahan flow lintas service, repository, billing/webhook, auth, dan journey UI kritikal |
| G5 - Coverage Gate | Hitung coverage pada scope terdampak | Backend domain/usecase ≥ 70%, frontend critical modules ≥ 60%, dan coverage total tidak turun > 2 poin tanpa exception | Perubahan kode production |
| G6 - Security Gate | Dependency scan + secret scan + static security scan | 0 secret findings, 0 high/critical vuln yang belum dimitigasi | Semua perubahan kode/dependency |
| G7 - Documentation Gate | Update docs terkait behavior/contract/arsitektur + link check | Dokumen terkait terupdate di PR yang sama dan markdown link check pass | Perubahan behavior, contract, arsitektur, runbook |
| G8 - Review & Merge Gate | Approval + branch protection + rollback plan (jika high-risk) | Low-risk: ≥1 approval, high-risk: ≥2 approval, required checks hijau | Semua PR |

## 2) Trigger Matrix (Gate per Jenis Perubahan)

| Jenis Perubahan | Gate Minimum |
|---|---|
| Dokumentasi saja (`docs/**`) | G7 |
| Backend logic non-IO | G1, G2, G3, G5, G6, G7, G8 |
| Backend repository/DB/billing/webhook/auth | G1-G8 (semua) |
| Frontend UI non-critical copy/layout | G1, G2, G3, G5, G7, G8 |
| Frontend auth/search/checkout/journey utama | G1-G8 (semua) |
| Penambahan dependency | Gate default perubahan + G6 wajib |

## 3) Exception Policy

Exception hanya boleh dipakai untuk G4/G5 dan harus memiliki seluruh evidence berikut:

1. alasan teknis terukur (contoh: suite sedang flaky, issue ID tercantum),
2. mitigasi sementara,
3. tanggal kedaluwarsa exception (maksimal 14 hari kalender),
4. approver yang menyetujui exception.

Tanpa empat item di atas, PR tetap blocked.

## 4) Branch Protection Baseline

- Required status checks wajib mencakup gate yang aktif pada PR.
- Branch harus up-to-date dengan target branch sebelum merge.
- Merge langsung ke protected branch tanpa PR tidak diperbolehkan.

## 5) Release Gate (Pre-Deploy)

| Item | Kriteria Lulus |
|---|---|
| Release notes/changelog | Tersedia dan menjelaskan dampak user/system |
| Database migration (jika ada) | `up`/`down` tervalidasi dan rollback step terdokumentasi |
| Observability readiness | Dashboard/alert untuk flow terdampak sudah aktif |
| Rollback readiness | Langkah rollback teruji di staging/simulasi |

Release tidak boleh dilakukan jika salah satu item di atas belum memenuhi kriteria.
