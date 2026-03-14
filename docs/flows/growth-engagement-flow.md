# Growth Engagement Flow

## 1) Saved Search & Watchlist Management

```mermaid
sequenceDiagram
  participant U as User
  participant FE as Frontend
  participant API as API
  participant DB as PostgreSQL

  U->>FE: Simpan query pencarian / follow company
  FE->>API: POST /saved-searches atau POST /watchlist/companies
  API->>DB: Simpan growth preference (user-scoped)
  API-->>FE: 201 Created

  FE->>API: GET /saved-searches + GET /watchlist/companies
  API->>DB: Load data growth preference
  API-->>FE: 200 OK (list)
```

## 2) Notification Center Read Flow

```mermaid
sequenceDiagram
  participant U as User
  participant FE as Frontend
  participant API as API
  participant DB as PostgreSQL

  FE->>API: GET /notifications?page=1&limit=20&unread_only=true
  API->>DB: Load notifications milik JWT.sub
  API-->>FE: 200 OK (paginated)

  U->>FE: Tandai sudah dibaca
  FE->>API: PATCH /notifications/:id/read
  API->>DB: Update read_at by ownership
  API-->>FE: 200 OK
```

## 3) Digest Preference Behavior

- `PUT /preferences/notification` mengontrol `alert_mode` (`instant`, `daily_digest`, `weekly_digest`) dan `digest_hour`.
- Saat `alert_mode` digest aktif, matcher menandai kandidat notifikasi sebagai `deferred_digest` (tidak langsung enqueue email instant).
- Entitlement premium tetap mengikuti canonical source `GET /billing/status.subscription_state`.
