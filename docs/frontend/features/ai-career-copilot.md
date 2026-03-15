# AI Career Copilot

## Objective

Menambahkan tool AI yang langsung bisa dipakai user dari account area untuk mempercepat discovery, evaluasi job fit, dan drafting cover letter.

## Scope (Phase 6 Frontend Increment 1)

- Route protected: `/account/ai-tools`.
- AI search assistant form (prompt + optional context).
- AI usage cards per feature:
  - `search_assistant`
  - `job_fit_summary`
  - `cover_letter_draft`
- AI job-fit summary form (premium path).
- AI cover letter draft form (premium path).
- Messaging premium upsell jika subscription belum `premium_active`.

## UX Improvements (Phase 6 Increment 2)

### 1) Search Assistant — "Open in Jobs" Button

- Di samping hasil generate search assistant, tersedia tombol **"Open in Jobs"** yang membuka halaman `/jobs?q={suggested_query}` di tab baru.
- URL dibangun menggunakan `buildSearchSubmitHref({ q: suggested_query })` dari `@/features/jobs/search-params`.
- Tombol hanya tampil setelah ada hasil generate (`assistantResult` tersedia).

### 2) Multi-Kata Full-Text Search (`GET /jobs`)

- Backend search mendukung query multi-kata (contoh: "backend intern").
- Query dipecah per spasi; setiap kata menjadi kondisi `AND` pada `title`, `company`, dan `description` (LIKE case-insensitive).
- Indeks `idx_jobs_title_lower` dan `idx_jobs_description_lower` tersedia via migrasi `000006_jobs_search_index`.

### 3) Job Title Autocomplete (Job Fit Summary & Cover Letter Draft)

- Input "Job ID" digantikan dengan field **"Job title"** bertipe autocomplete.
- Saat user mengetik ≥ 2 karakter, frontend memanggil `GET /api/v1/jobs/titles?q={query}` (debounce 300ms).
- Hasil tampil sebagai dropdown; saat user memilih salah satu judul:
  - `jobFitTitleQuery` / `coverLetterTitleQuery` diset ke judul yang dipilih.
  - `jobFitJobID` / `coverLetterJobID` diset ke `id` dari hasil `listJobs({ q: selectedTitle, limit: 10 })` pertama.
- `job_id` yang terkirim ke API backend tetap mengikuti kontrak yang ada; hanya UX input-nya yang berubah.
- Validasi: jika user tidak memilih dari dropdown, pesan error "Please select a job from the search results." muncul.

### 4) Cover Letter Draft — Copy to Clipboard

- Tombol **"Copy"** ditambahkan di samping tombol "Generate cover letter".
- Menggunakan `navigator.clipboard.writeText(draft)` (Web Clipboard API).
- State feedback: tombol berubah menjadi **"Copied!"** selama 2 detik, lalu kembali ke "Copy".
- Tombol disabled jika belum ada draft atau sedang generating.
- `copied` state di-reset setiap kali draft baru berhasil di-generate.

## Tier & Entitlement Rules

- Entitlement premium tetap mengikuti `GET /api/v1/billing/status -> subscription_state`.
- `search_assistant` dapat dipakai free dan premium dengan quota harian sesuai tier.
- `job_fit_summary` dan `cover_letter_draft` adalah premium-only.
- Jika backend mengembalikan `403 FORBIDDEN`, UI menampilkan state locked premium tanpa crash.

## API Dependencies

| Endpoint | Tujuan UI | Field minimum |
|---|---|---|
| `GET /api/v1/ai/usage?feature=...` | menampilkan quota meter per fitur | `feature`, `tier`, `daily_quota`, `used`, `remaining`, `reset_at` |
| `POST /api/v1/ai/search-assistant` | generate query refinement + filter suggestion | `suggested_query`, `suggested_filters`, `summary`, `quota_remaining` |
| `POST /api/v1/ai/job-fit-summary` | generate insight kecocokan user vs job | `fit_score`, `verdict`, `strengths`, `gaps`, `next_actions` |
| `POST /api/v1/ai/cover-letter-draft` | generate draft cover letter + key points | `tone`, `draft`, `key_points`, `quota_remaining` |
| `GET /api/v1/billing/status` | source of truth entitlement badge/upsell | `subscription_state` |
| `GET /api/v1/jobs/titles?q={query}` | autocomplete input judul lowongan (job-fit & cover letter) | `data.titles[]` |
| `GET /api/v1/jobs?q={title}&limit=10` | resolve job_id dari title yang dipilih | `data[0].id` |
## UI State Rules

- Usage cards:
  - `usage_loading` saat bootstrap,
  - `usage_ready` saat semua usage terambil,
  - `usage_error` saat API gagal.
- Search assistant:
  - `idle -> generating -> ready/error`.
- Job-fit summary:
  - `idle -> generating -> ready/locked/error`.
- Cover letter draft:
  - `idle -> generating -> ready/locked/error`.

## Error Handling

- `401`:
  - clear browser session,
  - redirect login dengan intent `/account/ai-tools`.
- `403 FORBIDDEN`:
  - tampilkan pesan premium-only dan CTA upgrade.
- `429 AI_QUOTA_EXCEEDED`:
  - tampilkan pesan quota exhausted + tunggu reset.
- `503`/provider unavailable:
  - tampilkan retry-friendly message.
- `404 NOT_FOUND` pada `job_id`:
  - tampilkan pesan validasi Job ID yang jelas.

## Acceptance Criteria

- User login bisa membuka `/account/ai-tools`.
- Usage quota per fitur tampil dan dapat di-refresh.
- Search assistant dapat menghasilkan suggestion dan memperbarui quota card.
- Tombol "Open in Jobs" membuka `/jobs?q={suggested_query}` di tab baru setelah hasil generate tersedia.
- Query multi-kata pada halaman Jobs menghasilkan hasil yang relevan (contoh: "backend intern" menemukan lowongan dengan kedua kata tersebut di title/description).
- Input job title pada job-fit summary dan cover letter draft menampilkan dropdown autocomplete dari `GET /jobs/titles`.
- Memilih judul dari dropdown mengisi `job_id` secara otomatis; input manual job_id tidak digunakan.
- Job-fit dan cover letter menampilkan hasil untuk user premium.
- Tombol "Copy" menyalin isi draft cover letter ke clipboard dan menampilkan feedback "Copied!" selama 2 detik.
- User non-premium mendapat pesan locked/premium-only untuk fitur premium.

## Related Specs

- [Profile & Account](./profile-account.md)
- [Premium Upgrade](./premium-upgrade.md)
- [AI API](../../api/ai.md)
- [Frontend-Backend Traceability](../traceability/frontend-backend-traceability.md)
