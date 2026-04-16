# PLAN-013: JWT Authentication (Backend + Frontend)

| | |
|-|-|
| **Status** | Not Started |
| **Date** | 2026-04-16 |
| **Depends on** | [PLAN-011](plan-011-frontend-scaffold.md), [PLAN-012](plan-012-admin-api.md) |

## Goal

Protect the `/admin/` API with JWT-based authentication.
Simple single-user admin: credentials (username + password hash) stored in env vars — no user DB.
Access token (short-lived) + refresh token (long-lived, httpOnly cookie).

## Token Design

| Token | TTL | Storage | Transport |
|-------|-----|---------|-----------|
| Access | 15 min | JS memory (not localStorage) | `Authorization: Bearer <token>` header |
| Refresh | 7 days | httpOnly + Secure cookie | Cookie on `/admin/auth/refresh` |

Access token in JS memory: survives page navigation, wiped on tab close.
Refresh token in httpOnly cookie: invisible to JS, not vulnerable to XSS.

## Backend

### Endpoints

| Method | Path | Auth required | Description |
|--------|------|--------------|-------------|
| `POST` | `/admin/auth/login` | No | Issue access + refresh tokens |
| `POST` | `/admin/auth/refresh` | No (cookie) | Issue new access token |
| `POST` | `/admin/auth/logout` | No | Clear refresh cookie |

### Login flow

```
POST /admin/auth/login
Body: { "username": "admin", "password": "..." }

200: { "access_token": "<jwt>", "expires_in": 900 }
     + Set-Cookie: refresh_token=<jwt>; HttpOnly; Secure; SameSite=Strict; Path=/admin/auth/refresh
401: { "error": "invalid_credentials" }
```

Password stored as bcrypt hash in env var `ADMIN_PASSWORD_HASH`.
Username in env var `ADMIN_USERNAME` (default `admin`).

### JWT Claims

```json
{
  "sub": "admin",
  "role": "admin",
  "iat": 1713261600,
  "exp": 1713262500
}
```

Signed with HS256, secret from env var `ADMIN_JWT_SECRET` (min 32 bytes, fail-fast on startup).

### Middleware

```go
func JWTAdminMiddleware(secret []byte) func(http.Handler) http.Handler
```

Validates `Authorization: Bearer <token>` on every `/admin/` request except `/admin/auth/*`.
Returns `401` on missing/expired/invalid token.

### Library

`github.com/golang-jwt/jwt/v5` — idiomatic, maintained, no CGO.

### Config additions

```go
type AdminConfig struct {
    Username     string // ADMIN_USERNAME, default "admin"
    PasswordHash string // ADMIN_PASSWORD_HASH (bcrypt)
    JWTSecret    string // ADMIN_JWT_SECRET
}
```

Fail fast at startup if `JWTSecret` is empty or shorter than 32 chars.

## Frontend

### Auth store

```ts
// src/lib/auth.ts
// Access token in module-level variable (JS memory only)
let accessToken: string | null = null

export const setAccessToken = (token: string) => { accessToken = token }
export const getAccessToken = () => accessToken
export const clearAccessToken = () => { accessToken = null }
```

### API client interceptors

`src/api/client.ts` wraps `openapi-fetch` with two interceptors:
1. **Request**: attach `Authorization: Bearer <accessToken>` header
2. **Response**: on `401`, call `POST /admin/auth/refresh` once, retry original request.
   If refresh also fails → redirect to `/login`.

### TanStack Router auth guard

`_authenticated` layout route checks `getAccessToken()` on each navigation.
If null → redirect to `/login` with `redirect` search param.

Login page calls `POST /admin/auth/login`, stores returned `access_token` via `setAccessToken`,
then navigates to `redirect` or `/dashboard`.

### Token refresh on page load

On app mount, call `POST /admin/auth/refresh` (cookie is sent automatically).
If succeeds → store access token, render app.
If fails → show login page.

## Acceptance Criteria

- [ ] `POST /admin/auth/login` with correct credentials returns access token + sets refresh cookie
- [ ] `POST /admin/auth/login` with wrong credentials returns `401`
- [ ] `GET /admin/accounts` without token returns `401`
- [ ] `GET /admin/accounts` with valid access token returns data
- [ ] `POST /admin/auth/refresh` with valid cookie returns new access token
- [ ] `POST /admin/auth/refresh` with expired/missing cookie returns `401`
- [ ] Access token expires after 15 min (test with short TTL in test env)
- [ ] `ADMIN_JWT_SECRET` shorter than 32 chars causes startup failure with clear error
- [ ] Frontend redirects unauthenticated requests to `/login`
- [ ] Frontend silently refreshes token on `401` and retries the original request
- [ ] Logout clears cookie and access token, redirects to `/login`

## Tasks

- [ ] Add `golang-jwt/jwt/v5` and `golang.org/x/crypto` to wallet-service `go.mod`
- [ ] Add `AdminConfig` to `config.go`, fail-fast validation on startup
- [ ] Implement `POST /admin/auth/login` handler (bcrypt verify + JWT issue)
- [ ] Implement `POST /admin/auth/refresh` handler (verify cookie JWT + issue new access token)
- [ ] Implement `POST /admin/auth/logout` handler (clear cookie)
- [ ] Implement `JWTAdminMiddleware` — validate Bearer token, skip `/admin/auth/*`
- [ ] Wire middleware onto `/admin` router group
- [ ] Frontend: auth store (`src/lib/auth.ts`)
- [ ] Frontend: API client with request/response interceptors
- [ ] Frontend: login page wired to `POST /admin/auth/login`
- [ ] Frontend: token refresh on app mount
- [ ] Frontend: auth guard in `_authenticated` layout
- [ ] Frontend: logout button in sidebar nav
- [ ] Update docs
