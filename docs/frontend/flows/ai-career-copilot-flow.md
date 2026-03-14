# AI Career Copilot Flow

Dokumen ini memetakan alur frontend untuk penggunaan AI tool pada route `/account/ai-tools`.

## Actors

- User authenticated
- Frontend web app (`apps/web`)
- Backend API (`/api/v1/ai/*`, `/api/v1/billing/status`)

## Entry Point

- User membuka `/account/ai-tools` dari account dashboard.

## Flow A — Bootstrap Entitlement + Usage Cards

1. Server page resolve session token.
2. Frontend membaca `billing/status` untuk entitlement canonical.
3. Client component memanggil:
   - `GET /api/v1/ai/usage?feature=search_assistant`
   - `GET /api/v1/ai/usage?feature=job_fit_summary`
   - `GET /api/v1/ai/usage?feature=cover_letter_draft`
4. UI merender quota cards dan premium/free indicator.

## Flow B — Search Assistant (Free + Premium)

1. User mengirim prompt + context opsional.
2. Frontend memanggil `POST /api/v1/ai/search-assistant`.
3. Jika sukses:
   - render `suggested_query`, `suggested_filters`, `summary`,
   - update usage card `search_assistant`.
4. Jika gagal:
   - `429`: tampilkan quota exhausted message,
   - `503/502`: tampilkan provider unavailable message.

## Flow C — Job Fit Summary (Premium)

1. User mengirim `job_id` (+ focus opsional).
2. Frontend memanggil `POST /api/v1/ai/job-fit-summary`.
3. Jika sukses:
   - render `fit_score`, `verdict`, `strengths`, `gaps`, `next_actions`,
   - update usage card `job_fit_summary`.
4. Jika `403 FORBIDDEN`:
   - render premium lock message + CTA upgrade.

## Flow D — Cover Letter Draft (Premium)

1. User mengirim `job_id`, `tone`, dan `highlights`.
2. Frontend memanggil `POST /api/v1/ai/cover-letter-draft`.
3. Jika sukses:
   - render `draft` + `key_points`,
   - update usage card `cover_letter_draft`.
4. Jika `403 FORBIDDEN`:
   - render premium lock message + CTA upgrade.

## Error & Recovery Rules

- `401`:
  - clear session,
  - redirect ke `/auth/login?redirect=/account/ai-tools`.
- `404` (`job_id` tidak ditemukan):
  - tampilkan message validasi Job ID.
- `429`:
  - tetap tampilkan hasil lama (jika ada) + cooldown message.
- `5xx`:
  - tampilkan retry-friendly message dan biarkan user submit ulang.

## Observability Notes (Frontend)

- Track AI action submission outcomes (success/failure) di client log sesuai policy observability frontend.
- Jangan simpan prompt mentah ke persistent client storage.

## Related Specs

- [AI Career Copilot Feature](../features/ai-career-copilot.md)
- [AI API](../../api/ai.md)
- [Frontend-Backend Traceability](../traceability/frontend-backend-traceability.md)
