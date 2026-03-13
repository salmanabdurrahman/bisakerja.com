# Preferences & Notifications

## Objective

Memungkinkan user mengelola preferensi lowongan dan memahami ketersediaan notifikasi berdasarkan status langganan (`free`, `pending_payment`, `premium_active`, `premium_expired`).

## Scope MVP

- Form preferensi: `keywords`, `locations`, `job_types`, `salary_min`.
- Validasi input sebelum submit.
- Simpan preferensi ke backend.
- Menampilkan status akses notifikasi berdasarkan subscription state.
- Upsell premium jika user belum `premium_active`.

## Out of Scope

- Pengaturan frekuensi notifikasi (instant/digest).
- In-app notification center/histori notifikasi.
- Multi-channel notification settings (email/WhatsApp toggle detail).

## UI Behavior Rules

- `keywords` wajib minimal 1 item (mengikuti validasi backend).
- Input list (`keywords`, `locations`, `job_types`) dinormalisasi: trim, dedupe, lowercase untuk konsistensi.
- Tombol simpan aktif hanya jika form valid dan ada perubahan.
- Bootstrap form:
  - ambil data awal dari `GET /api/v1/preferences`,
  - jika request awal gagal sementara, fallback ke cache/draft lokal terakhir + tombol retry,
  - jika user baru belum punya preferensi, gunakan default kosong dari response backend.
- Gating notifikasi:
  - `premium_active`: tampilkan status notifikasi aktif.
  - `free`, `pending_payment`, `premium_expired`: tampilkan pesan bahwa matching notifikasi hanya untuk premium aktif + CTA upgrade.
- Setelah simpan sukses, tampilkan konfirmasi non-blocking dan timestamp update terakhir.

## Notification Gating Decision Table

| `subscription_state` | Status notifikasi di UI | CTA |
|---|---|---|
| `premium_active` | enabled | Tidak ada CTA upsell |
| `free` | gated | Tampilkan CTA upgrade |
| `pending_payment` | gated (payment in progress) | Tampilkan CTA "Lanjutkan Pembayaran" |
| `premium_expired` | gated | Tampilkan CTA re-upgrade |

## Edge Cases

- **User menghapus semua keyword**: blok submit di client, tampilkan error terarah.
- **`salary_min` negatif/non-numeric**: blok submit, minta input valid.
- **Session expired saat submit** (`401`): simpan draft form lokal sementara lalu arahkan login.
- **Concurrency update** (tab ganda): data terakhir yang berhasil tersimpan dianggap source of truth.
- **`GET /api/v1/preferences` timeout/`5xx`**: render form dari cache/draft lokal + banner retry.
- **`429 TOO_MANY_REQUESTS` saat submit berulang**: tampilkan cooldown + cegah spam submit.

## API Dependencies

| Endpoint | Tujuan di Frontend | Field minimum yang dikonsumsi | Referensi |
|---|---|---|---|
| `GET /api/v1/preferences` | Bootstrap data preferensi user saat halaman dibuka. | `data.keywords`, `data.locations`, `data.job_types`, `data.salary_min`, `data.updated_at` | [preferences.md](../../api/preferences.md) |
| `PUT /api/v1/preferences` | Menyimpan preferensi pencarian/notifikasi user. | `data.keywords`, `data.locations`, `data.job_types`, `data.salary_min`, `data.updated_at` | [preferences.md](../../api/preferences.md) |
| `GET /api/v1/billing/status` | Menentukan state gating notifikasi. | `data.subscription_state` | [billing.md](../../api/billing.md) |
| `GET /api/v1/auth/me` | Validasi session user aktif sebelum submit perubahan preferensi. | `data.id`, `data.email` | [auth.md](../../api/auth.md) |

Catatan readiness kontrak:

- `GET` dan `PUT` preferences wajib memakai struktur field yang konsisten agar bootstrap + submit tidak perlu mapper ganda.
- Fallback cache/draft hanya dipakai saat gangguan sementara, bukan pengganti source of truth backend.

## Loading / Error / Empty States

- **Loading**
  - Saat bootstrap `GET /api/v1/preferences`: tampilkan skeleton form.
  - Saat simpan preferensi: tombol simpan loading + disable.
  - Saat ambil status billing: tampilkan badge/placeholder state.
- **Error**
  - `400 BAD_REQUEST`: tampilkan error field-level sesuai payload invalid.
  - `401 UNAUTHORIZED`: arahkan login ulang dan pulihkan draft lokal.
  - `500`/`503`: tampilkan toast error + retry.
- **Empty**
  - Form awal tanpa data sebelumnya: tampilkan placeholder bantuan pengisian.
  - Tidak ada `locations`/`job_types`: tetap valid selama `keywords` terisi.

## Acceptance Criteria

- User dapat menyimpan preferensi valid tanpa reload penuh halaman.
- Bootstrap preferensi awal menggunakan `GET /api/v1/preferences` dan tetap usable saat fallback retry.
- UI menolak input invalid sebelum request dikirim.
- Status akses notifikasi selalu mengikuti state subscription canonical.
- User non-`premium_active` tetap bisa set preferensi, namun mendapatkan pesan gating yang jelas.
- Saat `401` di submit, draft form dapat dipulihkan setelah user login ulang.

## Output Implementasi Minimum

- Route: `/account/preferences` (protected).
- Komponen: `PreferencesForm`, `NotificationEntitlementBanner`.
- Service calls typed: `getPreferences`, `updatePreferences`, `getBillingStatus`, `getMe`.
- Test minimum:
  - component test validasi field (`keywords`, `salary_min`),
  - integration test submit sukses + `updated_at` ter-render,
  - integration test `401` menyimpan draft dan meminta reauth.

## Related Specs

- [Job Discovery](./job-discovery.md)
- [Premium Upgrade](./premium-upgrade.md)
- [Profile & Account](./profile-account.md)
