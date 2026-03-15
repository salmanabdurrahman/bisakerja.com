# Application Tracker & Bookmark (Phase 6)

Dokumen ini mendeskripsikan fitur Application Tracker & Bookmark sebagai Phase 6 backend Bisakerja.

## Tujuan Bisnis

Meningkatkan stickiness produk dengan memberi user cara untuk menyimpan lowongan menarik (bookmark) dan melacak status lamaran mereka secara aktif. Diferensiasi free vs premium mendorong konversi upgrade melalui pembatasan jumlah tracking aktif pada free tier.

## Status Implementasi

- ✅ Phase 6 backend complete (38 tests pass, `go test ./...` clean).
- ✅ Phase 6 frontend complete (33 test files pass, typecheck clean).
- Dokumentasi API lengkap: [`docs/api/tracker.md`](../api/tracker.md).

## Domain Model

### Bookmark

Penanda bahwa user tertarik pada suatu lowongan. Tidak memiliki status; hanya ada atau tidak ada.

```
bookmarks
  id          TEXT PK
  user_id     TEXT NOT NULL  (dari JWT.sub)
  job_id      TEXT NOT NULL
  created_at  TIMESTAMPTZ NOT NULL
  UNIQUE(user_id, job_id)
```

### TrackedApplication

Rekaman bahwa user sudah melamar (atau akan melamar) ke suatu lowongan, dengan status pipeline-nya.

```
tracked_applications
  id          TEXT PK
  user_id     TEXT NOT NULL  (dari JWT.sub)
  job_id      TEXT NOT NULL
  status      TEXT NOT NULL  (applied | interview | offer | rejected | withdrawn)
  notes       TEXT
  created_at  TIMESTAMPTZ NOT NULL
  updated_at  TIMESTAMPTZ NOT NULL
  UNIQUE(user_id, job_id)
```

## Business Rules

### Bookmark

- Satu user hanya boleh bookmark satu job satu kali (UNIQUE constraint).
- Delete bookmark menggunakan `job_id` (bukan internal bookmark ID) agar memudahkan frontend tanpa state ID.
- Tidak ada limit jumlah bookmark untuk free tier.

### Application Tracking

- Status awal saat create selalu `applied`.
- Urutan status yang valid (tidak wajib berurutan): `applied` → `interview` → `offer` / `rejected` / `withdrawn`.
- **Free tier limit**: maksimal 5 **active** tracked applications (status bukan `rejected` atau `withdrawn`). Jika limit tercapai, backend mengembalikan `403 FORBIDDEN` dengan error code `TRACKER_LIMIT_EXCEEDED`.
- **Premium tier**: unlimited tracked applications.
- Satu user hanya boleh melacak satu job satu kali (UNIQUE constraint); duplikasi mengembalikan `409 CONFLICT`.

## API Endpoints

Kontrak lengkap ada di [`docs/api/tracker.md`](../api/tracker.md).

| Method | Path | Deskripsi |
|---|---|---|
| `POST` | `/bookmarks` | Tambah bookmark |
| `DELETE` | `/bookmarks/{job_id}` | Hapus bookmark |
| `GET` | `/bookmarks` | Daftar bookmark user |
| `POST` | `/applications` | Mulai melacak lamaran |
| `PATCH` | `/applications/{id}/status` | Update status pipeline |
| `DELETE` | `/applications/{id}` | Hapus tracked application |
| `GET` | `/applications` | Daftar tracked applications user |

## Layer Implementasi

```
apps/api/
  migrations/
    000007_phase6_application_tracker.up.sql    — tabel bookmarks + tracked_applications
    000007_phase6_application_tracker.down.sql
  internal/
    domain/tracker/tracker.go                  — domain types + FREE_TIER_APPLICATION_LIMIT = 5
    app/tracker/
      service.go                               — business logic bookmark + tracking
      service_test.go
    adapter/persistence/
      memory/tracker_repository.go             — in-memory adapter (test/dev)
      memory/tracker_repository_test.go
      postgres/tracker_repository.go           — PostgreSQL adapter
    adapter/http/handler/
      tracker_handler.go                       — HTTP handler 7 endpoint
      tracker_handler_test.go
    adapter/http/router/router.go              — route registration (modified)
  cmd/api/main.go                              — wiring tracker (modified)
  pkg/errcode/codes.go                         — TrackerLimitExceeded = "TRACKER_LIMIT_EXCEEDED" (modified)
  test/integration/tracker_flow_test.go        — integration test end-to-end
```

## Error Codes

| Code | HTTP | Kapan |
|---|---|---|
| `TRACKER_LIMIT_EXCEEDED` | 403 | User free tier mencapai batas 5 active tracked applications |
| `NOT_FOUND` | 404 | Bookmark atau application tidak ditemukan / bukan milik user |
| `CONFLICT` | 409 | Duplikasi bookmark atau application pada job yang sama |

## Observability

- Semua operasi tracker mengikuti structured logging standar (`request_id`, `user_id`, `job_id`).
- Tidak ada side effect worker/scheduler; operasi bersifat synchronous CRUD.

## Referensi

- API: [`docs/api/tracker.md`](../api/tracker.md)
- Frontend: [`docs/frontend/features/application-tracker.md`](../frontend/features/application-tracker.md)
- Traceability: [`docs/frontend/traceability/frontend-backend-traceability.md`](../frontend/traceability/frontend-backend-traceability.md)
- Migration: `apps/api/migrations/000007_phase6_application_tracker.up.sql`
