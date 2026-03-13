# Testing Strategy

Strategi testing Bisakerja bersifat risk-based dan menjadi gate merge/release (G3, G4, G5).

## 1) Test Pyramid

1. **Unit tests**: mayoritas, cepat, fokus logic murni/usecase.
2. **Integration tests**: validasi interaksi DB/Redis/API eksternal.
3. **E2E tests**: validasi journey user kritikal lintas boundary.

## 2) Mandatory Test Matrix (Per Jenis Perubahan)

| Jenis Perubahan | Wajib Dijalankan | Catatan Lulus |
|---|---|---|
| Refactor internal tanpa perubahan behavior | Unit test pada package/modul terdampak | Semua test pass |
| Endpoint/handler/usecase backend berubah | Unit + integration endpoint/repository terkait | Semua test pass, termasuk skenario error utama |
| Perubahan billing/webhook/migration | Unit + integration + idempotency test + failure-path test | Semua suite pass sebelum merge |
| Perubahan UI logic kritikal (auth/search/checkout/profile) | Unit/component + E2E journey terkait | Journey utama pass end-to-end |
| Bug produksi | Test reproduksi bug (gagal sebelum fix) + test pass setelah fix | Regression test wajib ada di PR yang sama |

## 3) Backend (Go) Minimum

- Unit test fokus pada domain/usecase logic.
- Integration test wajib untuk repository query penting, migration compatibility, dan webhook idempotency.
- Penamaan test mengikuti pola `Test<Function>_<Condition>_<ExpectedResult>`.

## 4) Frontend (Next.js) Minimum

- Unit/component test wajib untuk render logic, interaction, validation, loading/error/empty states.
- E2E minimal untuk journey:
  - login/register,
  - search jobs,
  - set preferences,
  - initiate checkout premium.

## 5) Coverage Gate (G5)

- Backend domain/usecase pada scope terdampak: **line coverage ≥ 70%**.
- Frontend critical modules pada scope terdampak: **line coverage ≥ 60%**.
- Coverage total tidak boleh turun > 2 poin dibanding target branch tanpa exception resmi.
- Exception coverage harus mengikuti policy di `ci-quality-gates.md` (alasan + expiry + approver).

## 6) Test Data & Determinism

- Fixture harus deterministik dan versioned.
- Data test tidak boleh mengandung secret production-like.
- Integration/E2E wajib membersihkan state setelah test selesai.

## 7) Flaky Test Policy

- Test yang flaky harus diberi label `flaky` dan issue perbaikan pada hari yang sama.
- Flaky test tidak boleh dibiarkan > 14 hari tanpa perbaikan atau keputusan deprecate.
- Jika flaky test dibypass sementara, harus tercatat sebagai exception aktif.

## 8) Evidence Model di PR

Untuk memenuhi G3/G4/G5, PR wajib menyertakan:

1. path file test yang ditambah/diubah,
2. command/suite yang dijalankan,
3. nama workflow/job CI yang pass,
4. angka coverage (bila gate coverage berlaku).
