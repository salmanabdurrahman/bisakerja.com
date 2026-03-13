# Job Discovery

## Objective

Memberikan pengalaman pencarian lowongan yang cepat, mudah dipahami, dan stabil untuk semua state langganan (`free`, `pending_payment`, `premium_active`, `premium_expired`).

## Scope MVP

- Halaman list lowongan dengan search dan filter dasar.
- Filter URL-driven: `q`, `location`, `salary_min`, `sort`, `page`, `limit`, `source`.
- Pagination dan navigasi ke detail lowongan.
- Detail lowongan dengan informasi inti dan external apply URL.
- Konteks upsell premium non-intrusif untuk user non-premium.

## Out of Scope

- Semantic recommendation/personalized ranking.
- Bookmark dan application tracker.
- Saved searches dan alert rules.

## UI Behavior Rules

- Input filter harus disinkronkan ke query URL agar bisa dibagikan (shareable state).
- Saat filter utama berubah (`q`, `location`, `salary_min`, `source`), `page` harus di-reset ke `1`.
- Request pencarian yang sudah usang (stale) harus dibatalkan/diabaikan saat user mengetik cepat.
- Sorting default mengikuti backend (`-posted_at`) jika user belum memilih manual.
- Klik item lowongan membuka halaman detail dengan opsi kembali ke hasil pencarian terakhir.
- Banner/CTA upsell hanya bergantung pada `billing/status.subscription_state`, bukan asumsi lokal.

## UI State Transitions

| Current State | Trigger | Next State | Catatan UI |
|---|---|---|---|
| `search_idle` | User submit/ubah filter | `search_loading` | Update query URL lebih dulu, reset `page=1` bila filter utama berubah |
| `search_loading` | `GET /api/v1/jobs` sukses + data ada | `search_results` | Render list + pagination dari `meta.pagination` |
| `search_loading` | `GET /api/v1/jobs` sukses + data kosong | `search_empty` | Tampilkan empty message + CTA reset filter |
| `search_loading` | `GET /api/v1/jobs` gagal `400` | `search_error_validation` | Tampilkan pesan validasi query |
| `search_loading` | `GET /api/v1/jobs` gagal `429` | `search_error_rate_limited` | Tampilkan cooldown hint + retry |
| `search_loading` | `GET /api/v1/jobs` gagal `5xx`/timeout | `search_error_retry` | Tampilkan retry action |
| `search_results` | User klik job card | `detail_loading` | Navigasi ke `/jobs/:id` dengan mempertahankan query asal |
| `detail_loading` | `GET /api/v1/jobs/:id` sukses | `detail_ready` | Render detail + tombol apply |
| `detail_loading` | `GET /api/v1/jobs/:id` gagal `404` | `detail_not_found` | Tampilkan CTA kembali ke list |

## Edge Cases

- **Query kosong**: tetap valid, tampilkan daftar lowongan terbaru.
- **Page di luar total**: tampilkan empty state dengan CTA kembali ke halaman 1.
- **`salary_min` invalid dari URL**: sanitasi query; jika backend tetap menolak, tampilkan error state.
- **Detail lowongan tidak ditemukan (`404`)**: tampilkan halaman not-found dengan navigasi balik ke list.
- **API rate limited (`429`)**: tampilkan retry dengan jeda, jangan spam auto-retry.
- **Stale response race**: response request lama tidak boleh menimpa hasil filter terbaru.

## API Dependencies

| Endpoint | Tujuan di Frontend | Field minimum yang dikonsumsi | Referensi |
|---|---|---|---|
| `GET /api/v1/jobs` | Ambil list/search jobs dengan query filter dan pagination. | `data[]`, `meta.pagination.page`, `meta.pagination.total_pages`, `meta.pagination.total_records` | [jobs.md](../../api/jobs.md) |
| `GET /api/v1/jobs/:id` | Ambil detail lowongan spesifik. | `data.id`, `data.title`, `data.company`, `data.location`, `data.description`, `data.url` | [jobs.md](../../api/jobs.md) |
| `GET /api/v1/billing/status` | Menentukan tampilan upsell premium berdasarkan state. | `data.subscription_state` | [billing.md](../../api/billing.md) |

## Loading / Error / Empty States

- **Loading**
  - Initial list load: tampilkan card skeleton.
  - Pergantian filter/pagination: tampilkan loading indicator non-blocking.
  - Detail page: tampilkan detail skeleton.
- **Error**
  - `400 BAD_REQUEST`: tampilkan pesan "filter tidak valid" + reset filter.
  - `500`/`503`: tampilkan full-state error + tombol retry.
  - Network timeout: tampilkan retry dan pertahankan filter terakhir.
- **Empty**
  - Search tanpa hasil: tampilkan pesan relevan + CTA clear filter.
  - Sumber lowongan kosong sementara: tampilkan pesan fallback tanpa crash layout.

## Acceptance Criteria

- Filter dan pagination konsisten dengan data API serta URL.
- User dapat membuka detail lowongan dari list dan kembali ke state pencarian sebelumnya.
- Empty/error/loading state tampil jelas tanpa kehilangan konteks query user.
- User non-premium tetap dapat discovery jobs, dengan upsell premium yang tidak memblokir alur utama.
- `429`/`5xx` tidak menyebabkan UI freeze; user selalu mendapat aksi lanjut (retry/ubah filter).

## Output Implementasi Minimum

- Route: `/jobs`, `/jobs/[id]`.
- Komponen: `JobsSearchBar`, `JobsFilterPanel`, `JobsList`, `JobDetail`, `JobsStatePanel`.
- Service calls typed: `listJobs`, `getJobDetail`, `getBillingStatus`.
- Test minimum:
  - integration test URL-driven filter + reset page,
  - component test `search_empty`, `search_error_validation`, `search_error_rate_limited`,
  - e2e list -> detail -> back-to-results.

## Related Specs

- [Auth & Session](./auth-session.md)
- [Premium Upgrade](./premium-upgrade.md)
- [Preferences & Notifications](./preferences-notifications.md)
