# Auth & Session

## Objective

Menyediakan pengalaman autentikasi yang aman, cepat, dan konsisten untuk register, login, pemulihan session, serta akses halaman terproteksi di aplikasi Next.js.

## Scope MVP

- Form register dan login dengan validasi client-side dasar.
- Session bootstrap saat app load (cek user aktif).
- Auto-refresh `access_token` menggunakan `refresh_token` sebelum kedaluwarsa.
- Route guard untuk halaman yang membutuhkan login.
- Logout client-side (clear session lokal + redirect ke login).

## Out of Scope

- Social login (Google/LinkedIn).
- Multi-factor authentication (MFA).
- Device/session management lintas perangkat.

## UI Behavior Rules

- User yang belum login dan mengakses route terproteksi harus diarahkan ke halaman login.
- Token sensitif tidak boleh disimpan di `localStorage`; gunakan mekanisme session yang selaras kontrak keamanan frontend.
- Setelah login sukses, app harus:
  1. menyimpan token session sesuai strategi keamanan frontend,
  2. mengambil profil user,
  3. redirect ke halaman tujuan awal.
- Setelah register sukses (`201`), user belum dianggap login (karena response register tidak mengembalikan token); arahkan ke login dengan email prefilled jika memungkinkan.
- Jika beberapa request bersamaan menerima `401`, frontend hanya boleh menjalankan **satu** proses refresh token (single-flight refresh).
- Jika refresh gagal (`401`), frontend wajib clear session lalu redirect ke login.
- Tombol submit login/register harus nonaktif selama request berjalan untuk mencegah double submit.

## Session State Model

| State | Sumber data utama | Transisi masuk | Transisi keluar |
|---|---|---|---|
| `anonymous` | tidak ada token valid | app start tanpa session / logout | login submit |
| `auth_submitting` | `POST /api/v1/auth/login` | user submit login | `authenticated` / `auth_error` |
| `authenticated` | token + `GET /api/v1/auth/me` | login sukses / refresh sukses | token expiring -> `session_refreshing` |
| `session_refreshing` | `POST /api/v1/auth/refresh` | interceptor menerima `401` | sukses -> `authenticated`, gagal -> `session_expired` |
| `session_expired` | refresh gagal (`401`) | dari `session_refreshing` | clear session -> redirect login (`anonymous`) |

## Edge Cases

- **Token expired di background tab**: ketika tab aktif lagi, frontend melakukan refresh sebelum request baru.
- **Refresh race condition**: request lain harus menunggu hasil refresh pertama, bukan memanggil refresh ulang.
- **User offline/intermittent network**: tampilkan status koneksi dan opsi retry tanpa menghapus input form.
- **Return URL invalid**: fallback redirect ke halaman default yang aman.
- **Register sukses lalu user menutup halaman**: saat kembali, user tetap harus login (tidak ada auto-auth implicit).
- **Refresh endpoint rate-limited (`429`)**: hentikan retry agresif, tampilkan sesi perlu login ulang bila retry budget habis.

## API Dependencies

| Endpoint | Tujuan di Frontend | Field minimum yang dikonsumsi | Referensi |
|---|---|---|---|
| `POST /api/v1/auth/register` | Registrasi akun baru. | `meta.code`, `data.email`, `data.id` | [auth.md](../../api/auth.md) |
| `POST /api/v1/auth/login` | Login dan mendapatkan token session. | `data.access_token`, `data.refresh_token`, `data.expires_in`, `data.token_type` | [auth.md](../../api/auth.md) |
| `POST /api/v1/auth/refresh` | Refresh `access_token` saat kedaluwarsa/akan kedaluwarsa. | `data.access_token`, `data.expires_in` | [auth.md](../../api/auth.md) |
| `GET /api/v1/auth/me` | Hydrate profil user di session global frontend. | `data.id`, `data.email`, `data.name`, `data.role`, `data.is_premium` | [auth.md](../../api/auth.md) |
| `GET /api/v1/billing/status` | Sinkron status langganan (`free`, `pending_payment`, `premium_active`, `premium_expired`) untuk UI global. | `data.subscription_state`, `data.premium_expired_at`, `data.last_transaction_status` | [billing.md](../../api/billing.md) |

## Loading / Error / Empty States

- **Loading**
  - Session bootstrap: tampilkan skeleton/layout placeholder.
  - Submit login/register: tampilkan inline loading di tombol.
- **Error**
  - `400 BAD_REQUEST`: tampilkan pesan validasi per field.
  - `401 UNAUTHORIZED`: tampilkan pesan kredensial salah atau session berakhir.
  - `429 TOO_MANY_REQUESTS`/`503 SERVICE_UNAVAILABLE`: tampilkan retry action.
- **Empty**
  - Tidak ada empty state data-list khusus untuk fitur auth-session.

## Acceptance Criteria

- User dapat register lalu login dengan alur sukses end-to-end.
- Session tetap valid melalui mekanisme refresh selama `refresh_token` masih berlaku.
- Route terproteksi tidak dapat diakses oleh user tanpa session valid.
- Gagal refresh selalu mengakhiri session secara bersih dan redirect ke login.
- State langganan di session frontend konsisten dengan response backend.
- Register sukses tidak menghasilkan session implicit; login tetap langkah terpisah sesuai kontrak API.

## Output Implementasi Minimum

- Route: `/auth/login`, `/auth/register`.
- Modul: auth service client, session store/provider, auth guard middleware/layout.
- HTTP middleware/interceptor dengan single-flight refresh.
- Test minimum:
  - integration test race `401` memastikan satu refresh call,
  - integration test register sukses -> redirect login,
  - e2e protected route redirect saat session invalid.

## Related Specs

- [Job Discovery](./job-discovery.md)
- [Premium Upgrade](./premium-upgrade.md)
- [Profile & Account](./profile-account.md)
