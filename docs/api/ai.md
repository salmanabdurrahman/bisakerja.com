# AI API (Phase 5 Increment 1-3)

Dokumen ini menjelaskan kontrak backend AI yang sudah diaktifkan pada increment awal Phase 5.

Semua endpoint:

- runtime path berprefix `/api/v1/*`,
- user-scoped dan wajib Bearer token valid,
- ownership wajib mengikuti `JWT.sub`.

## 1) Generate Search Assistant Suggestion

- **Method**: `POST`
- **Path**: `/ai/search-assistant`
- **Auth**: Bearer Token (user)

Endpoint ini membantu user menyusun query dan filter pencarian lowongan yang lebih relevan.

### Request Body

```json
{
  "prompt": "I want remote Golang backend jobs with salary above 15 million",
  "context": {
    "location": "Jakarta",
    "job_types": ["fulltime"],
    "salary_min": 15000000
  }
}
```

### Validation

| Field | Rules |
|---|---|
| `prompt` | wajib, 5..500 karakter |
| `context.location` | opsional, string |
| `context.job_types` | opsional, array string |
| `context.salary_min` | opsional, integer `>= 0` |

### Quota & Tier Rules

- Entitlement tier dihitung dari status premium user saat request diproses.
- Quota harian default:
  - free: `AI_DAILY_QUOTA_FREE` (default `5`)
  - premium: `AI_DAILY_QUOTA_PREMIUM` (default `30`)
- Jika quota habis, endpoint mengembalikan `429` dengan code `AI_QUOTA_EXCEEDED`.

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "AI search assistant generated",
    "request_id": "req_01J..."
  },
  "data": {
    "feature": "search_assistant",
    "prompt": "I want remote Golang backend jobs with salary above 15 million",
    "suggested_query": "golang backend engineer remote",
    "suggested_filters": {
      "locations": ["Jakarta", "Remote"],
      "job_types": ["fulltime"],
      "salary_min": 15000000
    },
    "summary": "Start from remote-friendly backend roles and keep location flexible.",
    "tier": "free",
    "provider": "openai_compatible",
    "model": "gpt-4.1-mini",
    "daily_quota": 5,
    "used_today": 2,
    "quota_remaining": 3,
    "reset_at": "2026-03-20T00:00:00Z"
  }
}
```

### Error

- `400 BAD_REQUEST` (`INVALID_AI_PROMPT`) jika prompt invalid.
- `401 UNAUTHORIZED` (`UNAUTHORIZED`) jika token invalid/tidak ada.
- `429 TOO_MANY_REQUESTS` (`AI_QUOTA_EXCEEDED`) jika quota habis.
- `502 BAD_GATEWAY` (`AI_PROVIDER_UPSTREAM_ERROR`) jika respons provider tidak valid.
- `503 SERVICE_UNAVAILABLE` (`AI_PROVIDER_RATE_LIMITED`, `AI_PROVIDER_UNAVAILABLE`) jika provider rate-limited/down.

## 2) Generate Job Fit Summary (Premium)

- **Method**: `POST`
- **Path**: `/ai/job-fit-summary`
- **Auth**: Bearer Token (user)

Endpoint ini menghasilkan ringkasan kecocokan profil user terhadap lowongan tertentu.

### Request Body

```json
{
  "job_id": "job_01H...",
  "focus": "prioritize architecture depth and ownership signal"
}
```

### Validation

| Field | Rules |
|---|---|
| `job_id` | wajib, string non-empty |
| `focus` | opsional, maksimal 300 karakter |

### Entitlement & Quota Rules

- Endpoint ini **premium-only**.
- Jika user tidak premium aktif, endpoint mengembalikan `403 FORBIDDEN`.
- Quota harian mengikuti `AI_DAILY_QUOTA_PREMIUM` (default `30`).
- Jika quota habis, endpoint mengembalikan `429` dengan code `AI_QUOTA_EXCEEDED`.

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "AI job fit summary generated",
    "request_id": "req_01J..."
  },
  "data": {
    "feature": "job_fit_summary",
    "job_id": "job_01H...",
    "fit_score": 84,
    "verdict": "strong_match",
    "strengths": [
      "Strong backend API delivery experience",
      "Relevant Golang stack alignment"
    ],
    "gaps": [
      "Need deeper distributed tracing examples"
    ],
    "next_actions": [
      "Highlight production impact metrics in CV"
    ],
    "summary": "Profile strongly aligns with the role and backend scope.",
    "tier": "premium",
    "provider": "openai_compatible",
    "model": "gpt-4.1-mini",
    "daily_quota": 30,
    "used_today": 6,
    "quota_remaining": 24,
    "reset_at": "2026-03-20T00:00:00Z"
  }
}
```

### Error

- `400 BAD_REQUEST` (`BAD_REQUEST`) jika `job_id` kosong atau `focus` terlalu panjang.
- `401 UNAUTHORIZED` (`UNAUTHORIZED`) jika token invalid/tidak ada.
- `403 FORBIDDEN` (`FORBIDDEN`) jika user belum premium aktif.
- `404 NOT_FOUND` (`NOT_FOUND`) jika `job_id` tidak ditemukan.
- `429 TOO_MANY_REQUESTS` (`AI_QUOTA_EXCEEDED`) jika quota habis.
- `502 BAD_GATEWAY` (`AI_PROVIDER_UPSTREAM_ERROR`) jika respons provider tidak valid.
- `503 SERVICE_UNAVAILABLE` (`AI_PROVIDER_RATE_LIMITED`, `AI_PROVIDER_UNAVAILABLE`) jika provider rate-limited/down.

## 3) Generate Cover Letter Draft (Premium)

- **Method**: `POST`
- **Path**: `/ai/cover-letter-draft`
- **Auth**: Bearer Token (user)

Endpoint ini menghasilkan draft cover letter berdasarkan lowongan, preferensi user, dan tone yang dipilih.

### Request Body

```json
{
  "job_id": "job_01H...",
  "tone": "professional",
  "highlights": [
    "Built and scaled Golang APIs",
    "Owned backend reliability initiatives"
  ]
}
```

### Validation

| Field | Rules |
|---|---|
| `job_id` | wajib, string non-empty |
| `tone` | opsional, salah satu: `professional`, `confident`, `friendly`, `concise` |
| `highlights` | opsional, array string (maks 5 item) |

### Entitlement & Quota Rules

- Endpoint ini **premium-only**.
- Jika user tidak premium aktif, endpoint mengembalikan `403 FORBIDDEN`.
- Quota harian mengikuti `AI_DAILY_QUOTA_PREMIUM` (default `30`).
- Jika quota habis, endpoint mengembalikan `429` dengan code `AI_QUOTA_EXCEEDED`.

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "AI cover letter draft generated",
    "request_id": "req_01J..."
  },
  "data": {
    "feature": "cover_letter_draft",
    "job_id": "job_01H...",
    "tone": "professional",
    "draft": "Dear Hiring Team, I am excited to apply for the Backend Engineer role...",
    "key_points": [
      "Go backend delivery experience",
      "API scalability and reliability ownership"
    ],
    "summary": "Professional draft focused on backend impact and role alignment.",
    "tier": "premium",
    "provider": "openai_compatible",
    "model": "gpt-4.1-mini",
    "daily_quota": 30,
    "used_today": 7,
    "quota_remaining": 23,
    "reset_at": "2026-03-20T00:00:00Z"
  }
}
```

### Error

- `400 BAD_REQUEST` (`BAD_REQUEST`) jika `job_id` kosong atau `tone` tidak valid.
- `401 UNAUTHORIZED` (`UNAUTHORIZED`) jika token invalid/tidak ada.
- `403 FORBIDDEN` (`FORBIDDEN`) jika user belum premium aktif.
- `404 NOT_FOUND` (`NOT_FOUND`) jika `job_id` tidak ditemukan.
- `429 TOO_MANY_REQUESTS` (`AI_QUOTA_EXCEEDED`) jika quota habis.
- `502 BAD_GATEWAY` (`AI_PROVIDER_UPSTREAM_ERROR`) jika respons provider tidak valid.
- `503 SERVICE_UNAVAILABLE` (`AI_PROVIDER_RATE_LIMITED`, `AI_PROVIDER_UNAVAILABLE`) jika provider rate-limited/down.

## 4) Get AI Usage

- **Method**: `GET`
- **Path**: `/ai/usage`
- **Auth**: Bearer Token (user)

Endpoint read model untuk menampilkan quota harian AI di UI.

### Query Parameters

| Name | Type | Required | Default | Description |
|---|---|---|---|---|
| `feature` | string | no | `search_assistant` | nilai yang didukung: `search_assistant`, `job_fit_summary`, `cover_letter_draft` |

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "AI usage retrieved",
    "request_id": "req_01J..."
  },
  "data": {
    "feature": "search_assistant",
    "tier": "premium",
    "daily_quota": 30,
    "used": 8,
    "remaining": 22,
    "reset_at": "2026-03-20T00:00:00Z"
  }
}
```

### Error

- `400 BAD_REQUEST` (`INVALID_AI_FEATURE`) untuk feature yang tidak didukung.
- `401 UNAUTHORIZED` (`UNAUTHORIZED`) jika token invalid/tidak ada.
- `500 INTERNAL_SERVER_ERROR` untuk kegagalan internal.

## 5) Konfigurasi Runtime

Konfigurasi minimum untuk AI gateway:

- `AI_PROVIDER_BASE_URL` (default `https://api.openai.com/v1`)
- `AI_PROVIDER_API_KEY`
- `AI_PROVIDER_MODEL_DEFAULT` (default `gpt-4.1-mini`)
- `AI_PROVIDER_TIMEOUT` (default `10s`)
- `AI_DAILY_QUOTA_FREE` (default `5`)
- `AI_DAILY_QUOTA_PREMIUM` (default `30`)

Catatan keamanan:

- Backend hanya menyimpan `prompt_hash`, bukan prompt mentah.
- Error response ke client tidak mengekspose detail internal provider.
