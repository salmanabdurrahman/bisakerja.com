# Code Review Checklist (Merge-Blocking)

Checklist ini dipakai author dan reviewer sebelum merge. Item dengan label **[BLOCKER]** wajib terpenuhi.

## Cara Pakai

1. Author mengisi evidence di deskripsi PR (code, test, CI, docs/ADR).
2. Reviewer menandai checklist ini saat review.
3. Jika ada item **[BLOCKER]** tidak terpenuhi, PR harus ditolak atau diminta revisi.

> Definisi high-risk change: perubahan schema DB, auth/authorization, billing/webhook, atau dependency inti.

## 1) Correctness

- [ ] **[BLOCKER]** Scope PR terhubung ke issue/requirement yang jelas.
- [ ] **[BLOCKER]** Perubahan behavior memiliki skenario verifikasi yang bisa direproduksi.
- [ ] Edge case utama minimal 1 skenario boundary dan 1 skenario failure sudah diuji.

## 2) Architecture & Design

- [ ] **[BLOCKER]** Boundary module/layer tetap konsisten (tidak ada business logic di layer yang salah).
- [ ] **[BLOCKER]** Dependency baru menyertakan justifikasi + hasil security/license check.
- [ ] Perubahan arsitektur signifikan memiliki ADR baru atau referensi ADR existing.

## 3) Quality

- [ ] **[BLOCKER]** Gate G1 (lint) pass di CI.
- [ ] **[BLOCKER]** Gate G2 (type/static analysis) pass di CI.
- [ ] Tidak ada TODO/FIXME baru tanpa issue ID dan rencana tindak lanjut.

## 4) Testing

- [ ] **[BLOCKER]** Gate G3 pass (unit/component test sesuai scope).
- [ ] **[BLOCKER]** Gate G4 pass jika perubahan termasuk flow berisiko.
- [ ] **[BLOCKER]** Bug fix menyertakan regression test pada PR yang sama.
- [ ] Gate G5 (coverage) pass atau memiliki exception resmi yang belum expired.

## 5) Security & Performance

- [ ] **[BLOCKER]** Gate G6 pass (secret scan + vulnerability scan + static security check).
- [ ] **[BLOCKER]** Validasi input dan authorization tercakup untuk endpoint/aksi protected.
- [ ] Untuk area performa kritikal, ada bukti tidak terjadi degradasi tanpa mitigasi.

## 6) Documentation

- [ ] **[BLOCKER]** Gate G7 pass: dokumen relevan di `docs/` diperbarui pada PR yang sama.
- [ ] Contract API berubah: dokumentasi contract/API ikut diperbarui.
- [ ] Markdown link check untuk file docs yang diubah berstatus pass.

## 7) Merge Readiness

- [ ] **[BLOCKER]** Gate G8 pass (approval + required checks + branch up-to-date).
- [ ] **[BLOCKER]** Untuk high-risk change, rollback strategy ditulis jelas.
- [ ] Reviewer memahami dampak operasional (monitoring, alerting, migration, fallback).
