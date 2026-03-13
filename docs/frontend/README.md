# Frontend Documentation Hub

Dokumen ini adalah pintu masuk utama untuk spesifikasi frontend Bisakerja (Next.js user-facing app).

## Peta Dokumen Frontend

- [Feature Specs](./features/README.md)
- [Architecture](./architecture/README.md)
- [User Flows](./flows/README.md)
- [Implementation Phases](./phases/README.md)
- [Frontend-Backend Traceability Matrix](./traceability/frontend-backend-traceability.md)

## Dokumen PRD Terkait

- [Frontend PRD](../prd/bisakerja-frontend-prd.md)
- [Backend/API PRD](../prd/bisakerja-api-prd.md)

## Cara Membaca

- Mulai dari Frontend PRD untuk memahami konteks produk dan scope MVP.
- Lanjutkan ke spesifikasi fitur, arsitektur, dan flow sesuai domain yang dikerjakan.
- Gunakan traceability matrix saat sinkronisasi kontrak frontend-backend dan acceptance evidence.

## Guardrails Implementasi Cepat

- Gunakan path runtime frontend dengan prefix `/api/v1/*`; backend docs menuliskan resource path inti tanpa prefix.
- Jadikan `billing/status.data.subscription_state` sebagai source of truth untuk entitlement premium di UI.
- Perlakukan `auth/me.is_premium` sebagai data pendukung, bukan penentu akhir entitlement jika konflik.
- Semua journey wajib mendefinisikan state `loading`, `empty`, `error`, dan `success` (jika ada mutation).
