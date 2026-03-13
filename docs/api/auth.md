# Authentication API

## 1) Register

- **Method**: `POST`
- **Path**: `/auth/register`
- **Auth**: Public

### Request Body

```json
{
  "email": "user@example.com",
  "password": "StrongPassword123!",
  "name": "Budi"
}
```

### Validation

- `email`: wajib, format email valid, unik (case-insensitive).
- `password`: wajib, minimal 8 karakter, minimal 1 huruf besar + 1 angka.
- `name`: wajib, 2-100 karakter.

### Response `201 Created`

```json
{
  "meta": {
    "code": 201,
    "status": "success",
    "message": "User registered",
    "request_id": "req_01J..."
  },
  "data": {
    "id": "2f2d5b4f-4ad8-4379-88df-d03790d1e9df",
    "email": "user@example.com",
    "name": "Budi",
    "role": "user",
    "created_at": "2026-03-13T18:00:00Z"
  }
}
```

### Error

- `409 CONFLICT` (`EMAIL_ALREADY_REGISTERED`) jika email sudah terdaftar.
- `400 BAD_REQUEST` jika payload invalid.

## 2) Login

- **Method**: `POST`
- **Path**: `/auth/login`
- **Auth**: Public

### Request Body

```json
{
  "email": "user@example.com",
  "password": "StrongPassword123!"
}
```

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Login successful",
    "request_id": "req_01J..."
  },
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

### Error

- `401 UNAUTHORIZED` (`INVALID_CREDENTIALS`) untuk kombinasi email/password salah.

## 3) Refresh Token

- **Method**: `POST`
- **Path**: `/auth/refresh`
- **Auth**: Public

### Request Body

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI..."
}
```

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Token refreshed",
    "request_id": "req_01J..."
  },
  "data": {
    "access_token": "new_access_token",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

### Error

- `401 UNAUTHORIZED` jika refresh token invalid/expired.

## 4) Get Current User

- **Method**: `GET`
- **Path**: `/auth/me`
- **Auth**: Bearer Token

### Response `200 OK`

```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Profile retrieved",
    "request_id": "req_01J..."
  },
  "data": {
    "id": "2f2d5b4f-4ad8-4379-88df-d03790d1e9df",
    "email": "user@example.com",
    "name": "Budi",
    "role": "user",
    "is_premium": true,
    "premium_expired_at": "2026-04-13T18:00:00Z",
    "subscription_state": "premium_active"
  }
}
```

### Catatan

- `subscription_state` mengikuti enum canonical (`free`, `pending_payment`, `premium_active`, `premium_expired`).
- Source of truth entitlement tetap `GET /billing/status`; field di `/auth/me` dipertahankan untuk sinkronisasi profil.

### Error

- `401 UNAUTHORIZED` jika token hilang atau invalid.
