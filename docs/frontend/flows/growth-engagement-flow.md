# Growth Engagement Flow

Dokumen ini merangkum flow UI untuk fitur growth di account area.

## 1) Saved Searches Flow

1. User membuka `/account/saved-searches`.
2. UI bootstrap list via `GET /api/v1/saved-searches`.
3. User submit form create:
   - validasi query/salary di client,
   - kirim `POST /api/v1/saved-searches`,
   - prepend item baru ke list.
4. User hapus item:
   - klik hapus pada card,
   - kirim `DELETE /api/v1/saved-searches/:id`,
   - item di-remove dari state list.

State utama:

- `saved_search_idle`
- `saved_search_submitting`
- `saved_search_ready`
- `saved_search_error`

## 2) Notification Center Flow

1. User membuka `/account/notifications`.
2. UI bootstrap list via `GET /api/v1/notifications?page=1&limit=20`.
3. User toggle `unread_only`:
   - reload data dengan query baru.
4. User klik `Tandai dibaca`:
   - kirim `PATCH /api/v1/notifications/:id/read`,
   - update `read_at` item secara in-place.

State utama:

- `notif_center_loading`
- `notif_center_ready`
- `notif_center_empty`
- `notif_center_error`

## 3) Notification Digest Preference Flow

1. User membuka `/account/preferences`.
2. UI bootstrap preference + digest setting via `GET /api/v1/preferences`.
3. User ubah `alert_mode` (`instant`/`daily_digest`/`weekly_digest`) dan optional `digest_hour`.
4. User simpan:
   - validasi `digest_hour` client-side (0-23),
   - kirim `PUT /api/v1/preferences/notification`,
   - tampilkan confirmation + `updated_at`.

State utama:

- `digest_pristine`
- `digest_dirty`
- `digest_submitting`
- `digest_saved`
- `digest_error_validation`

## Error & Recovery Rules

- `401`: clear session browser + redirect ke `/auth/login?redirect=<current-path>`.
- `400`: tampilkan pesan validasi terarah pada form/action terkait.
- `429`: tampilkan cooldown message.
- `5xx`: tampilkan retry-friendly error di section terkait.

## Referensi API

- [Growth API](../../api/growth.md)
- [Preferences API](../../api/preferences.md)
- [Auth API](../../api/auth.md)
