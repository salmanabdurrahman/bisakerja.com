# Engineering Standards Hub

Dokumen ini adalah source of truth standar engineering Bisakerja (backend Go + frontend Next.js).

## Tujuan

- Menetapkan gate kualitas yang **terukur** dan **merge-blocking**.
- Menyamakan model evidence antara standar, roadmap, dan checklist phase.
- Memastikan setiap milestone roadmap punya acceptance criteria yang bisa diverifikasi.

## Dokumen Standar

- [Go Coding Standards](./go-coding-standards.md)
- [Next.js Coding Standards](./nextjs-coding-standards.md)
- [Comments and Docstrings Standards](./comments-and-docstrings.md)
- [Testing Strategy](./testing-strategy.md)
- [CI Quality Gates](./ci-quality-gates.md)
- [Security and Observability Standards](./security-observability-standards.md)
- [Code Review Checklist](./code-review-checklist.md)
- [ADR Guidelines](./adr-guidelines.md)

## Quality Gates Wajib (Definition of Engineering Done)

PR hanya boleh merge jika seluruh gate wajib (G1-G8) pada scope perubahan berstatus **pass**.

| Gate | Aturan Lulus | Evidence Minimum | Referensi |
|---|---|---|---|
| G1 - Lint | Tidak ada lint error | nama command + hasil CI job | [CI Quality Gates](./ci-quality-gates.md) |
| G2 - Type/Static Analysis | Tidak ada type/static analysis error | hasil check di CI | [CI Quality Gates](./ci-quality-gates.md) |
| G3 - Unit/Component Test | Seluruh test wajib pass | path test + CI job | [Testing Strategy](./testing-strategy.md) |
| G4 - Integration/E2E Test | Suite wajib pass untuk flow berisiko | daftar suite + CI job | [Testing Strategy](./testing-strategy.md) |
| G5 - Coverage Gate | Coverage minimal tercapai, tanpa drop tak terjustifikasi | angka coverage + perbandingan baseline | [Testing Strategy](./testing-strategy.md) |
| G6 - Security Gate | Tidak ada finding blocker (secret/high-critical vuln) | hasil scan + ticket mitigasi jika diizinkan | [Security and Observability Standards](./security-observability-standards.md) |
| G7 - Documentation Gate | Dokumen terkait behavior/contract diperbarui + link valid | path dokumen + hasil markdown link check | [CI Quality Gates](./ci-quality-gates.md) |
| G8 - Review & Merge Gate | Approval sesuai risk level + rollback plan (jika high-risk) | link review + catatan rollback | [Code Review Checklist](./code-review-checklist.md) |

## Model Evidence Wajib di PR

Untuk setiap item checklist yang ditandai `✅`, isi evidence berikut:

1. **Code evidence**: path file yang berubah (repo-relative).
2. **Test evidence**: path test + nama suite/command.
3. **CI evidence**: nama workflow/job + status.
4. **Docs/ADR evidence**: path dokumen/ADR yang diperbarui, atau `N/A` jika tidak relevan.

Tanpa evidence minimum di atas, status item harus `🟡 Partial`, bukan `✅ Implemented`.

## Integrasi dengan Roadmap dan Checklist

- Backend/core roadmap: [`../phases/implementation-roadmap.md`](../phases/implementation-roadmap.md)
- Backend/core checklist: [`../phases/implementation-checklist.md`](../phases/implementation-checklist.md)
- Frontend roadmap: [`../frontend/phases/implementation-roadmap.md`](../frontend/phases/implementation-roadmap.md)
- Frontend checklist: [`../frontend/phases/implementation-checklist.md`](../frontend/phases/implementation-checklist.md)

Semua dokumen phase wajib mereferensikan gate dan model evidence yang sama dengan dokumen ini.
