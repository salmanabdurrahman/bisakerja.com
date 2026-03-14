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
- Job-fit dan cover letter menampilkan hasil untuk user premium.
- User non-premium mendapat pesan locked/premium-only untuk fitur premium.

## Related Specs

- [Profile & Account](./profile-account.md)
- [Premium Upgrade](./premium-upgrade.md)
- [AI API](../../api/ai.md)
- [Frontend-Backend Traceability](../traceability/frontend-backend-traceability.md)
