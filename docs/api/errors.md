# API Error Codes

Dokumen ini menetapkan standar error lintas endpoint.

## Error Envelope

```json
{
  "meta": {
    "code": 400,
    "status": "error",
    "message": "Validation error",
    "request_id": "req_01J..."
  },
  "errors": [
    {
      "field": "email",
      "code": "INVALID_EMAIL",
      "message": "email is required"
    }
  ]
}
```

## Standar Field

- `meta.request_id` wajib ada di semua response 4xx/5xx untuk tracing.
- `errors[]` boleh kosong untuk error non-field, tetapi tetap gunakan `code` machine-readable.
- `errors[].field` boleh `null` untuk error non-field.

## HTTP Error Matrix

| HTTP | Error Code | Kapan Digunakan |
|---|---|---|
| 400 | `BAD_REQUEST` | Payload/query invalid. |
| 401 | `UNAUTHORIZED` | Token/signature tidak valid. |
| 403 | `FORBIDDEN` | Auth valid tapi role/ownership tidak sesuai. |
| 404 | `NOT_FOUND` | Resource tidak ditemukan. |
| 409 | `CONFLICT` | Duplikasi data/event atau state conflict. |
| 422 | `UNPROCESSABLE_ENTITY` | Payload valid format tapi gagal aturan bisnis. |
| 429 | `TOO_MANY_REQUESTS` | Rate limit terlampaui. |
| 500 | `INTERNAL_SERVER_ERROR` | Gangguan server internal. |
| 502 | `BAD_GATEWAY` | Upstream dependency memberi respons invalid/error. |
| 503 | `SERVICE_UNAVAILABLE` | Dependency down atau retry exhausted. |

## Domain Error Codes (MVP)

| Endpoint Area | Code | HTTP | Kapan Dipakai |
|---|---|---|---|
| auth | `EMAIL_ALREADY_REGISTERED` | 409 | Register dengan email yang sudah ada. |
| auth | `INVALID_CREDENTIALS` | 401 | Login gagal. |
| jobs | `INVALID_SORT` | 400 | `sort` di luar whitelist. |
| jobs | `INVALID_LIMIT` | 400 | `limit <= 0` atau bukan integer valid. |
| preferences | `INVALID_JOB_TYPE` | 400 | Nilai `job_types` tidak masuk enum. |
| billing | `INVALID_PLAN_CODE` | 400 | `plan_code` tidak didukung. |
| billing | `INVALID_COUPON_CODE` | 400 | `coupon_code` tidak valid/tidak berlaku di Mayar. |
| billing | `INVALID_REDIRECT_URL` | 400 | URL tidak HTTPS/di luar allowlist. |
| billing | `ALREADY_PREMIUM` | 409 | User masih premium aktif. |
| billing | `MAYAR_RATE_LIMITED` | 503 | Mayar `429` setelah retry exhausted. |
| billing | `MAYAR_UPSTREAM_ERROR` | 502 | Mayar memberi error non-rate-limit. |
| ai | `INVALID_AI_PROMPT` | 400 | Prompt AI kosong/terlalu pendek/terlalu panjang. |
| ai | `INVALID_AI_FEATURE` | 400 | Feature AI tidak didukung endpoint usage. |
| ai | `FORBIDDEN` | 403 | Endpoint AI premium-only dipanggil oleh user non-premium aktif. |
| ai | `AI_QUOTA_EXCEEDED` | 429 | Quota AI harian user sudah habis. |
| ai | `AI_PROVIDER_RATE_LIMITED` | 503 | Provider AI mengembalikan rate limit. |
| ai | `AI_PROVIDER_UPSTREAM_ERROR` | 502 | Provider AI memberi respons invalid/non-2xx. |
| ai | `AI_PROVIDER_UNAVAILABLE` | 503 | Provider AI tidak tersedia atau timeout/network failure. |
| webhooks | `INVALID_WEBHOOK_TOKEN` | 401 | Token webhook tidak valid. |
| webhooks | `INVALID_WEBHOOK_PAYLOAD` | 400 | Field wajib webhook tidak lengkap. |
| webhooks | `WEBHOOK_USER_NOT_FOUND` | 422 | Email customer tidak bisa dipetakan ke user lokal. |

## Mapping Error Upstream (Mayar -> Bisakerja)

| Kondisi dari Mayar | Respons Internal Bisakerja |
|---|---|
| `400` bad payload | `502 BAD_GATEWAY` + `MAYAR_UPSTREAM_ERROR` (kecuali validasi kupon) |
| `401` API key invalid | `502 BAD_GATEWAY` + `MAYAR_UPSTREAM_ERROR` + alert konfigurasi |
| `404` resource not found | `502 BAD_GATEWAY` (kecuali endpoint validasi kupon) |
| `429` rate limit | retry policy, lalu `503 SERVICE_UNAVAILABLE` + `MAYAR_RATE_LIMITED` |
| `500` error upstream | retry policy, lalu `503 SERVICE_UNAVAILABLE` |

Catatan khusus validasi kupon:

- untuk call `GET /hl/v1/coupon/validate`, respons `400/404` dari Mayar dipetakan menjadi `400 BAD_REQUEST` dengan code `INVALID_COUPON_CODE`.
