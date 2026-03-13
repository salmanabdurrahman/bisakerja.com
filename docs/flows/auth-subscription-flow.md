# Auth & Subscription Flow

## 1) Register/Login/Refresh

```mermaid
sequenceDiagram
  participant U as User
  participant FE as Frontend
  participant API as API
  participant DB as PostgreSQL

  U->>FE: Input email/password
  FE->>API: POST /auth/register atau /auth/login
  API->>DB: validate/create user
  DB-->>API: user record
  API-->>FE: access_token + refresh_token
  FE-->>U: logged in

  FE->>API: POST /auth/refresh (when access expired)
  API-->>FE: new access_token
```

## 2) Upgrade to Premium

```mermaid
sequenceDiagram
  participant U as User
  participant FE as Frontend
  participant API as API
  participant M as Mayar
  participant DB as PostgreSQL

  U->>FE: Click Upgrade
  FE->>API: POST /billing/checkout-session + Idempotency-Key
  API->>M: Create customer/invoice (retry 429/5xx max 3x)
  M-->>API: checkout_url + transactionId
  API->>DB: Save transaction(status=pending)
  API-->>FE: checkout_url + subscription_state=pending_payment
  FE->>M: Redirect checkout
  U->>M: Complete payment

  M->>API: POST /webhook/mayar (payment.received/payment.reminder)
  API->>API: Validate token + payload
  API->>DB: Insert webhook_deliveries(idempotency key)
  alt first event
    API->>DB: Update transactions status
    API->>DB: Update users premium (only if status=success)
  else duplicate event
    API->>API: Skip side effect
  end
  API-->>M: 200 OK

  FE->>API: GET /billing/status
  API-->>FE: subscription_state canonical
```

## Failure Path

| Kondisi | Respons | Dampak |
|---|---|---|
| Webhook token invalid | `401` | Event ditolak, tidak ada perubahan data |
| Webhook duplicate | `200` idempotent | Tidak ada update ganda |
| Mayar `429/5xx` saat create invoice | retry internal, lalu `503` jika gagal | User diminta retry |
| User tidak ditemukan dari webhook email | `422` | Event dicatat untuk rekonsiliasi manual |
