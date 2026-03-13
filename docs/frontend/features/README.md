# Frontend Feature Specs (Next.js User App)

Dokumen ini mendefinisikan spesifikasi fitur frontend yang modular untuk aplikasi Next.js user-facing.

## Terminology (Canonical)

Gunakan istilah status berikut secara konsisten di seluruh UI:

| Term | Arti di UI | Sumber kontrak backend |
|---|---|---|
| `free` | User belum memiliki langganan premium aktif. | `GET /api/v1/billing/status -> data.subscription_state` |
| `pending_payment` | User sudah memulai checkout, pembayaran belum terkonfirmasi. | `GET /api/v1/billing/status -> data.subscription_state` |
| `premium_active` | User premium aktif dan mendapat akses fitur premium. | `GET /api/v1/billing/status -> data.subscription_state` |
| `premium_expired` | User pernah premium, tetapi masa aktif sudah berakhir. | `GET /api/v1/billing/status -> data.subscription_state` |

### Transaction Status (Billing)

Gunakan istilah transaksi sesuai endpoint:

| Field | Nilai utama | Catatan implementasi frontend |
|---|---|---|
| `billing/transactions.data[].status` | `pending`, `reminder`, `success`, `failed` | Dipakai untuk histori transaksi di UI account/subscription. |
| `billing/status.data.last_transaction_status` | contoh: `pending`, `reminder`, `success`, `failed` | Informasi status pembayaran terakhir; jangan dijadikan source of truth entitlement. |

### API Path Convention

- Dokumen feature frontend menggunakan path runtime `/api/v1/*`.
- Dokumen API backend dapat menggunakan path inti tanpa prefix (`/auth/*`, `/jobs`, `/billing/*`, `/preferences`).
- Saat implementasi client, gunakan base URL + resource path agar tidak hardcode host.

### State Ownership Hierarchy

Saat ada konflik data antar endpoint:

1. **Entitlement premium UI**: prioritaskan `billing/status.subscription_state`.
2. **Identitas user**: gunakan `auth/me` (`id`, `email`, `name`, `role`).
3. **Flag `is_premium` di `auth/me`**: indikator pendukung/fallback sementara.

## Feature Modules

1. [Auth & Session](./auth-session.md)
2. [Job Discovery](./job-discovery.md)
3. [Premium Upgrade](./premium-upgrade.md)
4. [Preferences & Notifications](./preferences-notifications.md)
5. [Profile & Account](./profile-account.md)

## Dependency Flow (Frontend)

- **Auth & Session** menjadi fondasi token, identitas user, dan guard halaman.
- **Profile & Account** dan **Premium Upgrade** mengonsumsi state langganan dari session/billing.
- **Job Discovery** memanfaatkan state langganan untuk upsell konteks premium.
- **Preferences & Notifications** memanfaatkan state langganan untuk gating fitur notifikasi.

## Backend API References

- [Auth API](../../api/auth.md)
- [Jobs API](../../api/jobs.md)
- [Billing API](../../api/billing.md)
- [User Preferences API](../../api/preferences.md)
- [Error Codes](../../api/errors.md)
- [API Hub](../../api/README.md)
