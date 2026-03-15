# Profile & Account

## Objective

Menyediakan halaman akun yang ringkas untuk menampilkan identitas user, status langganan, dan aksi account dasar yang relevan di MVP, dengan navigasi dashboard yang konsisten di seluruh area account.

## Scope MVP

- Menampilkan profil user (`name`, `email`, `role`) dari session backend.
- Menampilkan status langganan canonical (`free`, `pending_payment`, `premium_active`, `premium_expired`).
- Menampilkan informasi masa aktif premium jika tersedia.
- Menyediakan navigasi dashboard bersama di seluruh route account: `/account`, `/account/preferences`, `/account/saved-searches`, `/account/notifications`, `/account/subscription`, `/account/ai-tools`.
- Logout client-side yang membersihkan session frontend.

## Out of Scope

- Edit profil, ubah email, ubah password.
- Hapus akun dan manajemen keamanan lanjutan.
- Riwayat aktivitas akun detail.

## UI Behavior Rules

- Halaman account hanya bisa diakses user dengan session valid.
- Data profil diambil dari `GET /api/v1/auth/me`.
- State langganan diambil dari `GET /api/v1/billing/status`; jika berbeda dengan indikator profil, frontend memprioritaskan `subscription_state` dari billing.
- Semua halaman area account menampilkan navigation yang sama untuk:
  - account overview,
  - preferences,
  - saved searches,
  - notification center,
  - subscription,
  - AI tools,
  - logout.
- Status badge harus menggunakan istilah canonical tanpa sinonim tambahan:
  - `free`
  - `pending_payment`
  - `premium_active`
  - `premium_expired`
- Prioritas render badge:
  1. `billing/status.data.subscription_state` (utama),
  2. fallback sementara `auth/me.is_premium` saat billing belum terambil,
  3. tampilkan `status_unavailable` bila billing gagal sementara.
- Aksi logout harus:
  1. membersihkan token/session lokal,
  2. mereset cache data user,
  3. redirect ke halaman login/public.

## UI State Transitions

| Current State | Trigger | Next State | Catatan UI |
|---|---|---|---|
| `account_bootstrap` | Halaman account diakses dengan session valid | `account_loading` | Fetch paralel `auth/me` + `billing/status` |
| `account_loading` | `auth/me` sukses + `billing/status` sukses | `account_ready` | Render profil + badge canonical |
| `account_loading` | `auth/me` sukses + `billing/status` gagal `5xx`/timeout | `account_partial_error` | Profil tetap tampil, panel langganan retryable |
| `account_ready` | Token mendekati expired | `session_refreshing` | Jalankan refresh single-flight |
| `session_refreshing` | Refresh sukses | `account_ready` | Lanjutkan tanpa kehilangan konteks halaman |
| `session_refreshing` | Refresh gagal (`401`) | `session_expired_redirect` | Clear session, simpan intent route, redirect login |

## Edge Cases

- **`auth/me` sukses, `billing/status` gagal sementara**: tampilkan profil lebih dulu, tampilkan fallback badge "status unavailable" + retry.
- **Session timeout saat berada di halaman account**: tampilkan notifikasi session berakhir lalu redirect login.
- **Data premium kedaluwarsa**: transisi badge ke `premium_expired` tanpa membutuhkan reload manual.
- **Konflik data** (`auth/me.is_premium=true` tapi `subscription_state=premium_expired`): tampilkan `premium_expired` dan log event observability.
- **`billing/transactions` lambat**: section histori boleh deferred, tidak boleh memblokir render profil utama.

## API Dependencies

| Endpoint | Tujuan di Frontend | Field minimum yang dikonsumsi | Referensi |
|---|---|---|---|
| `GET /api/v1/auth/me` | Menampilkan data profil user aktif. | `id`, `email`, `name`, `role`, `is_premium`, `premium_expired_at` | [auth.md](../../api/auth.md) |
| `POST /api/v1/auth/refresh` | Menjaga session tetap valid saat user aktif di halaman account. | `access_token`, `expires_in` | [auth.md](../../api/auth.md) |
| `GET /api/v1/billing/status` | Menampilkan state subscription canonical dan info premium expiry. | `subscription_state`, `is_premium`, `premium_expired_at`, `last_transaction_status` | [billing.md](../../api/billing.md) |
| `GET /api/v1/billing/transactions` | Menampilkan histori transaksi ringkas di section account (opsional MVP ringan). | `data[].status`, `data[].amount`, `data[].created_at` | [billing.md](../../api/billing.md) |

## Loading / Error / Empty States

- **Loading**
  - Initial profile load: tampilkan profile skeleton.
  - Refresh status langganan: tampilkan inline loading pada badge/status block.
- **Error**
  - `401 UNAUTHORIZED`: clear session + redirect login.
  - `500`/`503` pada billing: tampilkan pesan non-blocking pada panel langganan.
  - Network issue: tampilkan tombol retry data profile.
- **Empty**
  - Histori transaksi kosong: tampilkan "Belum ada transaksi".
  - Field profil opsional kosong (mis. name) tampilkan fallback aman, bukan layout kosong.

## Acceptance Criteria

- Halaman account menampilkan profil user login secara konsisten.
- Badge status langganan selalu mengikuti `billing/status.subscription_state` jika tersedia.
- Dashboard navigation shared tersedia konsisten di semua halaman account yang relevan.
- Logout bekerja bersih tanpa menyisakan data session sensitif di client.
- Error parsial (mis. billing gagal) tidak membuat halaman account unusable.
- Saat refresh token gagal, user selalu diarahkan ulang ke login tanpa state account yang corrupt.

## Output Implementasi Minimum

- Route: `/account` (protected).
- Komponen: `ProfileSummary`, `SubscriptionBadge`, shared `AccountDashboardShell`/navigation, `TransactionsSummary` (opsional MVP ringan).
- Service calls typed: `getMe`, `refreshToken`, `getBillingStatus`, `getBillingTransactions`.
- Test minimum:
  - component/integration untuk `account_ready` dan `account_partial_error`,
  - component test shared account navigation + logout,
  - integration test refresh gagal -> redirect login.

## Related Specs

- [Auth & Session](./auth-session.md)
- [Premium Upgrade](./premium-upgrade.md)
- [Preferences & Notifications](./preferences-notifications.md)
