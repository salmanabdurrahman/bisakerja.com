# Growth & Engagement

## Objective

Meningkatkan retensi user dengan fitur growth yang user-scoped: saved searches, notification center, dan kontrol frekuensi notifikasi.

## Scope (Phase 2 Frontend)

- Halaman `/account/saved-searches` untuk create/list/delete saved search.
- Halaman `/account/notifications` untuk list notifikasi + mark as read.
- Kontrol digest di `/account/preferences` untuk `alert_mode` dan `digest_hour`.

## UI Behavior Rules

- Semua aksi growth wajib lewat session client (refresh token single-flight tetap aktif).
- Saved search:
  - `query` wajib 2-200 karakter.
  - `salary_min` opsional tapi harus integer `>= 0`.
- Notification center:
  - default menampilkan halaman pertama (`page=1`, `limit=20`),
  - tersedia filter `unread_only`,
  - `mark as read` hanya muncul jika `read_at` kosong.
- Digest control:
  - `alert_mode=instant` menyembunyikan input `digest_hour`,
  - mode digest menampilkan `digest_hour` opsional (`0..23`),
  - jika `digest_hour` kosong, backend menggunakan default `9`.

## API Dependencies

| Endpoint | Tujuan UI | Field minimum |
|---|---|---|
| `POST /api/v1/saved-searches` | membuat saved search | `id`, `query`, `frequency`, `is_active`, `created_at` |
| `GET /api/v1/saved-searches` | menampilkan saved searches | array saved search user |
| `DELETE /api/v1/saved-searches/:id` | menghapus saved search | `id` |
| `GET /api/v1/notifications` | menampilkan notification center | `data[]`, `meta.pagination` |
| `PATCH /api/v1/notifications/:id/read` | menandai notifikasi dibaca | `id`, `read_at` |
| `PUT /api/v1/preferences/notification` | update mode instant/digest | `alert_mode`, `digest_hour`, `updated_at` |
| `GET /api/v1/preferences` | bootstrap pengaturan digest | `alert_mode`, `digest_hour` |

## Error Handling

- `401`: clear session client, redirect login dengan redirect target halaman saat ini.
- `400`: tampilkan error validasi terarah di form/action terkait.
- `429`: tampilkan cooldown message.
- `5xx`: tampilkan error retry-friendly tanpa crash halaman.

## Acceptance Criteria

- User bisa menambah dan menghapus saved search dari UI account.
- User bisa melihat daftar notifikasi dan menandai item sebagai dibaca.
- User bisa mengubah `alert_mode` dan `digest_hour` dari halaman preferences.
- Lint/test/build frontend tetap pass setelah fitur growth ditambahkan.

## Related Specs

- [Preferences & Notifications](./preferences-notifications.md)
- [Profile & Account](./profile-account.md)
- [Growth Engagement Flow](../flows/growth-engagement-flow.md)
- [Growth API](../../api/growth.md)
