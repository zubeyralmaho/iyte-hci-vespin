# API Reference

The full HTTP contract is defined in
[`backend/api/openapi.yaml`](../backend/api/openapi.yaml). This document
provides a high-level summary and conventions guide.

## Base URL

| Environment | URL |
|-------------|-----|
| Local development | `http://localhost:8080` |
| Production | `https://<configured-domain>` |

## Conventions

- **Content type:** `application/json` for all request and response bodies.
- **Field naming:** `camelCase` in JSON; `snake_case` in the database.
- **IDs:** UUIDs as strings (format: `uuid`).
- **Timestamps:** RFC 3339 strings in UTC (e.g., `2026-05-18T14:30:00Z`).
- **Partial updates:** Use `PATCH`, never `PUT`.
- **Health check:** `GET /healthz` — no auth required.

## Authentication

All endpoints (except `/healthz`, `/auth/guest`, `/auth/register` without a
token, and `/auth/login`) require a JWT bearer token in the `Authorization`
header:

```
Authorization: Bearer <token>
```

See [Authentication](./authentication.md) for full details on auth flows.

## Error Format

All errors follow a consistent envelope:

```json
{
  "error": {
    "code": "snake_case_error_code",
    "message": "Human-readable description"
  }
}
```

Error codes are defined in the OpenAPI spec. Do not invent new codes without
adding them to the spec first.

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `unauthorized` | 401 | Missing or invalid JWT token |
| `forbidden` | 403 | Insufficient role or not resource owner |
| `not_found` | 404 | Resource does not exist |
| `validation_failed` | 400 | Request body failed validation |
| `email_taken` | 409 | Email already registered |
| `already_registered` | 409 | Account is already registered |
| `internal_error` | 500 | Unexpected server error |

## Endpoint Groups

### Auth (`/auth/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/auth/guest` | Create anonymous guest account | None |
| POST | `/auth/register` | Register new account or convert guest | None or Guest JWT |
| POST | `/auth/login` | Login with email/password | None |

### Users (`/users/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/users/me` | Get current user profile | Any |
| GET | `/users/me/preferences` | Get user preferences | Registered |
| PATCH | `/users/me/preferences` | Update user preferences | Registered |

### Devices (`/devices/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/devices` | List user's paired devices | Any |
| POST | `/devices` | Pair a new device | Any |
| GET | `/devices/{id}` | Get device details | Any (owner) |
| PATCH | `/devices/{id}` | Update device | Any (owner) |
| DELETE | `/devices/{id}` | Remove paired device | Any (owner) |

### EQ Profiles (`/eq-profiles/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/eq-profiles` | List system presets + user's custom profiles | Any |
| POST | `/eq-profiles` | Create a custom EQ profile | Any |
| GET | `/eq-profiles/{id}` | Get profile details | Any (owner or system) |
| PATCH | `/eq-profiles/{id}` | Update custom profile | Any (owner) |
| DELETE | `/eq-profiles/{id}` | Delete custom profile | Any (owner) |
| POST | `/eq-profiles/{id}/fork` | Fork a system preset into a custom profile | Any |

### Party Sessions (`/party-sessions/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/party-sessions` | List user's party sessions | Any |
| POST | `/party-sessions` | Create a party session | Any |
| GET | `/party-sessions/{id}` | Get session details | Any (owner) |
| PATCH | `/party-sessions/{id}` | Update session (status changes) | Any (owner) |
| DELETE | `/party-sessions/{id}` | Delete a party session | Any (owner) |
| POST | `/party-sessions/{id}/devices` | Add device to session | Any (owner) |
| DELETE | `/party-sessions/{id}/devices/{deviceId}` | Remove device from session | Any (owner) |

### Firmware (`/firmware/*`)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/firmware/check` | Check firmware version | Registered |

## Pagination

List endpoints that support pagination use query parameters:

- `limit` — maximum results to return (default varies by endpoint)
- `offset` — number of results to skip

## OpenAPI Spec as Source of Truth

The spec at `backend/api/openapi.yaml` is the canonical reference. When
behavior changes, update the spec first, then implement. The frontend API
client is auto-generated from this spec using Orval.
