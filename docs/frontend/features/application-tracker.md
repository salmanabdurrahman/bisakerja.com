# Application Tracker Frontend (Phase 6)

Dokumen ini mendeskripsikan implementasi frontend untuk fitur Application Tracker & Bookmark pada Bisakerja.

## Tujuan

Memberikan user pengalaman yang terintegrasi untuk menyimpan lowongan favorit langsung dari halaman detail job dan melacak status pipeline lamaran mereka dari dashboard akun.

## Status Implementasi

- ✅ Phase 6 frontend complete (33 test files pass, `pnpm --filter web typecheck` zero errors).
- Backend dependency: Phase 6 backend complete — semua 7 endpoint aktif.

## Modul & File

```
apps/web/src/
  services/
    tracker.ts                                        — tipe Bookmark, TrackedApplication, ApplicationStatus + service functions
  services/session-api-client.ts                     — 7 metode tracker ditambahkan (modified)
  features/tracker/components/
    bookmark-button.tsx                              — tombol bookmark di job detail (toggle)
    account-tracker-client.tsx                       — client component dashboard tracker (/account/tracker)
  app/account/tracker/page.tsx                       — server component route /account/tracker (protected)
  features/profile/components/
    account-dashboard-nav.tsx                        — link "Application tracker" ditambahkan (modified)
  app/jobs/[id]/page.tsx                             — BookmarkButton diintegrasikan (modified)

apps/web/tests/
  unit/tracker-service.test.ts                       — unit test service tracker
  components/account-tracker-client.test.tsx         — component test dashboard tracker
  components/job-detail-page.test.tsx                — modified: verifikasi BookmarkButton render
```

## Types

```typescript
// apps/web/src/services/tracker.ts
export type ApplicationStatus =
  | "applied"
  | "interview"
  | "offer"
  | "rejected"
  | "withdrawn";

export interface Bookmark {
  id: string;
  job_id: string;
  created_at: string;
}

export interface TrackedApplication {
  id: string;
  job_id: string;
  status: ApplicationStatus;
  notes: string;
  created_at: string;
  updated_at: string;
}
```

## Fitur Per Komponen

### BookmarkButton (`/jobs/[id]`)

- Menampilkan ikon bookmark di halaman detail lowongan.
- State: bookmarked vs not bookmarked, ditentukan dari `GET /bookmarks` saat server render.
- Toggle: `POST /bookmarks` untuk tambah, `DELETE /bookmarks/{job_id}` untuk hapus.
- Hanya tampil untuk user yang login; guest tidak melihat tombol ini.

### AccountTrackerClient (`/account/tracker`)

- Dua tab: **Bookmarks** dan **Applications**.
- **Bookmarks tab**: daftar lowongan yang di-bookmark user; tombol hapus per item.
- **Applications tab**:
  - Daftar tracked applications dengan pipeline status badge per item.
  - Dropdown update status: `applied` → `interview` → `offer` → `rejected` / `withdrawn`.
  - Tombol hapus tracked application.
  - Free tier: banner notice jika mendekati limit 5 active applications; error `TRACKER_LIMIT_EXCEEDED` ditangani dengan pesan yang actionable (upgrade prompt).
  - Premium tier: unlimited, tanpa banner limit.

## API Dependencies

| Method | Path | Digunakan di |
|---|---|---|
| `POST /api/v1/bookmarks` | Tambah bookmark | BookmarkButton |
| `DELETE /api/v1/bookmarks/{job_id}` | Hapus bookmark | BookmarkButton + AccountTrackerClient |
| `GET /api/v1/bookmarks` | List bookmarks | job detail page (SSR) + AccountTrackerClient |
| `POST /api/v1/applications` | Mulai tracking | AccountTrackerClient |
| `PATCH /api/v1/applications/{id}/status` | Update status pipeline | AccountTrackerClient |
| `DELETE /api/v1/applications/{id}` | Hapus tracked app | AccountTrackerClient |
| `GET /api/v1/applications` | List tracked apps | AccountTrackerClient |

## UX States

Setiap section pada AccountTrackerClient mendefinisikan state:

| State | Tampilan |
|---|---|
| `loading` | skeleton placeholder |
| `empty` | ilustrasi + CTA (mis. "Browse jobs") |
| `error` | error message + retry button |
| `success` | list items |
| `limit_exceeded` | banner upgrade + disable tombol "Track Application" baru |

## Tier Differentiation

| Capability | Free | Premium |
|---|---|---|
| Bookmark jobs | Unlimited | Unlimited |
| Tracked applications | Maks 5 active | Unlimited |
| Status update pipeline | ✅ | ✅ |
| Notes per application | ✅ | ✅ |

## Navigasi

Route `/account/tracker` terdaftar di sidebar account melalui `account-dashboard-nav.tsx` dengan label "Application tracker".

## Referensi

- Backend: [`docs/features/application-tracker.md`](../../features/application-tracker.md)
- API: [`docs/api/tracker.md`](../../api/tracker.md)
- Traceability: [`docs/frontend/traceability/frontend-backend-traceability.md`](../traceability/frontend-backend-traceability.md)
