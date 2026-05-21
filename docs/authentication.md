# Authentication

## Overview

Vespin uses JWT bearer tokens for authentication with two user roles:

- **Guest** — anonymous access, limited features
- **Registered** — full access with email/password credentials

Tokens have approximately 30-day expiry with no refresh token mechanism.

## Auth Flows

### 1. Guest Access

```text
Client                          Server
  │                               │
  │  POST /auth/guest             │
  │  (no body, no auth)           │
  │──────────────────────────────►│
  │                               │  Create user with role='guest'
  │  { token, user }              │  Generate JWT
  │◄──────────────────────────────│
```

A guest is a real row in the `users` table with `role='guest'`, no email, and
no password. Guests can pair devices, manage EQ profiles, and create party
sessions.

### 2. Registration (Fresh Account)

```text
Client                          Server
  │                               │
  │  POST /auth/register          │
  │  { email, password,           │
  │    displayName }              │
  │  (no auth header)             │
  │──────────────────────────────►│
  │                               │  Create user with role='registered'
  │  { token, user }              │  Hash password, generate JWT
  │◄──────────────────────────────│
```

### 3. Guest-to-Registered Conversion

```text
Client                          Server
  │                               │
  │  POST /auth/register          │
  │  { email, password,           │
  │    displayName }              │
  │  Authorization: Bearer <guest>│
  │──────────────────────────────►│
  │                               │  UPDATE existing user row:
  │                               │  - Set email, password_hash
  │                               │  - Set role='registered'
  │                               │  - Set converted_at
  │  { token, user }              │  Generate new JWT with updated role
  │◄──────────────────────────────│
```

This is the same endpoint as fresh registration, but the presence of a valid
guest JWT triggers conversion mode. The user's ID does not change — all
existing devices, EQ profiles, and party sessions remain attached.

### 4. Login

```text
Client                          Server
  │                               │
  │  POST /auth/login             │
  │  { email, password }          │
  │──────────────────────────────►│
  │                               │  Verify credentials
  │  { token, user }              │  Generate JWT
  │◄──────────────────────────────│
```

## JWT Token Structure

Tokens contain:

- `sub` — user UUID
- `role` — `guest` or `registered`
- `exp` — expiration timestamp (~30 days)
- `iat` — issued-at timestamp

## Role-Based Access

| Feature | Guest | Registered |
|---------|-------|------------|
| Pair/manage devices | ✓ | ✓ |
| Browse/create EQ profiles | ✓ | ✓ |
| Fork system presets | ✓ | ✓ |
| Create party sessions | ✓ | ✓ |
| View/update preferences | ✗ | ✓ |
| Check firmware updates | ✗ | ✓ |

## Security Notes

- Passwords are hashed with bcrypt before storage.
- JWT signing uses a server-side secret (`JWT_SECRET` env var).
- Tokens are passed in the `Authorization: Bearer <token>` header.
- There are no refresh tokens — clients re-authenticate when tokens expire.
- OAuth buttons exist in the UI for the HCI deliverable, but only
  email/password is implemented on the backend.
