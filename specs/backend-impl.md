# Feature: Backend HTTP Implementation

## Summary

The Vespin backend skeleton (config, db pool, auth primitives, httpx, error
mapping, sqlc queries for devices/eq-profiles/party-sessions/preferences, all 7
migrations) is in place, but every domain handler is an empty `RegisterRoutes`
stub. None of the 24 endpoints in `backend/api/openapi.yaml` is wired. This
spec covers implementing all endpoints in one pass: auth (guest/register/login
with guest-to-registered conversion), users + preferences, devices (with
auto-assigned default EQ + seeded simulation fields), EQ profiles (CRUD +
copy-on-write fork), party sessions (with transactional create + status
transitions + device add/remove), and firmware check. It also adds the missing
`users.sql` queries, a `RequireRegistered` middleware, a router restructure so
`/users/*` sits behind auth and `/auth/*` stays public, and the four
table-driven tests prescribed by `backend/CLAUDE.md` plus a unit test for the
party-session status-transition function.

## Requirements

1. Implement every path in `backend/api/openapi.yaml` exactly as written.
   Status codes, error codes, response shapes, and field names must match.
2. Add `backend/internal/db/queries/users.sql` with the queries needed by the
   auth and users domains. Regenerate sqlc.
3. Create `auth.Handler` (in the existing `internal/auth` package) owning
   `/auth/guest`, `/auth/register`, `/auth/login`. Mount it as a public group
   in `server.NewRouter`.
4. Add `auth.RequireRegistered` middleware. Apply it to
   `GET/PATCH /users/me/preferences` and `POST /firmware/check`.
5. Restructure `server.NewRouter` so `/users/*`, `/devices/*`, `/eq-profiles/*`,
   `/party-sessions/*`, `/firmware/*` all live under `AuthMW`. Keep `/healthz`
   and `/auth/*` outside the auth group.
6. POST /devices seeds simulation fields server-side: `firmware_version =
   "1.0.0"`, `battery_level = rand[20..100]`, `is_connected = rand bool`. The
   device's `active_eq_profile_id` is auto-assigned to the default system
   preset (the `is_default=true` row in `eq_profiles`) which is loaded once
   at server startup and passed to `devices.NewHandler`.
7. POST /eq-profiles/{id}/fork creates a new `is_system=false` profile owned by
   the current user, with the supplied name + bands. The source preset must
   have `is_system=true`; otherwise return `400 not_a_system_preset`. No
   comparison between source and submitted bands.
8. POST /party-sessions creates a session and inserts junction rows in a single
   transaction after verifying every supplied `deviceIds[]` belongs to the
   caller. Any unknown/foreign device → `400 invalid_device_reference`.
9. PATCH /party-sessions/{id} enforces legal status transitions
   (`active↔paused`, both → `ended`, `ended` terminal) inside a single
   transaction: SELECT the current row via `GetPartySessionByIDAndOwner`,
   check the transition with the pure `legalTransition` function, then run
   `UpdatePartySession`. No row-level lock — the read is a plain SELECT.
   Concurrent transitions to the same target are benign (last write wins;
   `status=ended` causes the existing sqlc query to set `ended_at` via
   `COALESCE`, preserving the first ended_at if both writers commit). HCI
   scope does not justify adding a `FOR UPDATE` variant.
10. PATCH /devices/{id}: when the patch **sets** `activeEqProfileId` to a
    non-null UUID, call `GetAccessibleEQProfile(profileID, userID)` first; if
    the row does not exist (system OR owned by caller), return
    `400 invalid_eq_profile_reference`. When the patch **sets** it to `null`,
    clear directly without lookup. When the field is **omitted**, leave it
    untouched. See "Optional PATCH fields" below for how the three states are
    distinguished at decode time.
11. POST /auth/register is mounted as a public route. The handler parses the
    `Authorization` header itself:
    - Header absent → fresh registration.
    - Header present with valid guest JWT → convert that guest row.
    - Header present with valid registered JWT → `409 already_registered`.
    - Header present but invalid/expired → `401 unauthorized`.
12. Guest-to-registered conversion happens via `UPDATE users SET role,
    email, password_hash, display_name, converted_at` on the existing row. The
    user's ID never changes. Devices/EQ profiles/party sessions remain attached
    because they reference `user_id` unchanged. A new JWT with
    `role=registered` is issued and returned.
13. On `409 email_taken` during conversion: the guest row is **not** modified.
    The caller's existing guest JWT remains valid and they can retry with a
    different email.
14. Every list endpoint filters by the current user's ID (devices, custom EQ
    profiles, party sessions). System EQ presets are accessible to all
    authenticated users. No endpoint trusts a user ID supplied in the request.
15. Every PATCH endpoint accepts an empty body (`{}`) and returns `200` with
    the current state unchanged (`UserUpdateRequest`, `DeviceUpdateRequest`,
    `EQProfileUpdateRequest`, `PartySessionUpdateRequest`,
    `PreferencesUpdateRequest` all have no required fields). PATCH request
    DTOs use the `Optional[T]` wrapper (see "Optional PATCH fields") for any
    field whose three states (omitted / present-with-null / present-with-value)
    are semantically distinct.
16. Validation caps:
    - Names (device, eq-profile, party-session): `min=1,max=100`, trimmed.
    - `displayName`: `omitempty,max=100`.
    - `email`: `required,email,max=254`.
    - `password`: `required,min=8,max=72` (bcrypt input limit).
    - `language`: `min=2,max=8` (ISO 639-1 base + optional region; defaults `en`).
    - `bands.*`: `min=-12,max=12` (already in `eqprofiles.Bands`).
17. Membership writes on a session with `status='ended'` are rejected.
    `POST /party-sessions/{id}/devices` and
    `DELETE /party-sessions/{id}/devices/{deviceId}` return
    `400 invalid_status_transition` when the session is ended. Reads
    (`GET /party-sessions/{id}`, `GET /party-sessions`) still work.
    Rationale: `ended` is terminal — both for status PATCHes and for the
    session's membership graph.
18. The four CLAUDE-mandated table-driven tests plus a pure unit test for the
    party-session status-transition function. No handler-level integration
    tests for trivial CRUD.
19. Random source for simulated device fields is `math/rand/v2` (Go 1.22+).
    No seeding; per-call independent. Cryptographic randomness is not needed
    because the values are not security-relevant.

## Optional PATCH fields

Standard `encoding/json` collapses three JSON inputs into the same Go zero
value: `{}` (field omitted), `{"x": null}`, and `{"x": ""}` for strings all
deserialize a plain `string` field to `""`. PATCH endpoints in this API need
to distinguish all three states — most critically:

- `DeviceUpdateRequest.activeEqProfileId` — omit = keep, `null` = clear,
  uuid = set (and validate via `GetAccessibleEQProfile`).
- `UserUpdateRequest.displayName` — omit = keep, value = set. Empty string is
  treated as "clear" (allowed per the Edge Cases table).
- `RegisterRequest.displayName` — omit = preserve existing (only relevant for
  the convert flow), value = set.

A shared generic wrapper lives in `internal/httpx/optional.go`:

```go
package httpx

type Optional[T any] struct {
    Set   bool // true if the JSON key was present
    Null  bool // true if the value was JSON null
    Value T    // populated only when Set && !Null
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
    o.Set = true
    if string(data) == "null" {
        o.Null = true
        return nil
    }
    return json.Unmarshal(data, &o.Value)
}
```

DTOs use it for fields with tri-state semantics:

```go
// In package devices:
type UpdateRequest struct {
    Name              httpx.Optional[string]    `json:"name"              validate:"omitempty"`
    ActiveEqProfileID httpx.Optional[uuid.UUID] `json:"activeEqProfileId" validate:"omitempty"`
}
```

Handlers branch on `.Set` and `.Null`:

```go
if req.ActiveEqProfileID.Set {
    if req.ActiveEqProfileID.Null {
        // clear: pass pgtype.UUID{Valid: false}
    } else {
        // validate via GetAccessibleEQProfile, then set
    }
}
// else: omitted — keep current value (read from current row)
```

For PATCH endpoints, the handler reads the current row first, applies the
`Set` fields (replacing the existing value with the new one or with NULL),
and writes the merged result via the existing sqlc UPDATE query. Validator
tags like `min=1,max=100` only fire when the field was Set; an omitted
Optional is treated as zero-value and skipped by `validate.Struct`.

Plain `string`/`int` fields (no nullable semantics) stay as bare Go fields —
no need to wrap `theme`, `language`, `notificationsEnabled`, `bands`, etc.

## Data Model Changes

No new migrations. No schema changes. The existing schema covers everything.

New file: `backend/internal/db/queries/users.sql`. Run `sqlc generate` from
`backend/` after writing it.

```sql
-- name: CreateGuestUser :one
INSERT INTO users (role) VALUES ('guest') RETURNING *;

-- name: CreateRegisteredUser :one
INSERT INTO users (role, email, password_hash, display_name)
VALUES ('registered', $1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ConvertGuestToRegistered :one
UPDATE users
SET role = 'registered',
    email = $2,
    password_hash = $3,
    display_name = COALESCE($4, display_name),
    converted_at = now()
WHERE id = $1
  AND role = 'guest'
RETURNING *;

-- name: UpdateUserDisplayName :one
UPDATE users
SET display_name = $2
WHERE id = $1
RETURNING *;

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1) AS exists;
```

## API Endpoints

All paths from `backend/api/openapi.yaml`. Auth column:
`public` = no JWT required, `auth` = any valid JWT (guest or registered),
`registered` = JWT with `role=registered` only.

| Method | Path                                 | Auth         | Notes |
|--------|--------------------------------------|--------------|-------|
| GET    | `/healthz`                            | public       | Already wired. |
| POST   | `/auth/guest`                         | public       | INSERT guest user, return JWT. |
| POST   | `/auth/register`                      | optional     | Dual-mode per Requirement 11. |
| POST   | `/auth/login`                         | public       | Verify bcrypt + return JWT. |
| GET    | `/users/me`                           | auth         | Return user from context. |
| PATCH  | `/users/me`                           | auth         | Update displayName only. |
| GET    | `/users/me/preferences`               | registered   | Return row or defaults if missing. |
| PATCH  | `/users/me/preferences`               | registered   | Upsert with provided fields layered over defaults. |
| GET    | `/devices`                            | auth         | List by owner. |
| POST   | `/devices`                            | auth         | Create with seeded sim fields + default EQ. |
| GET    | `/devices/{id}`                       | auth         | Owner-scoped lookup. |
| PATCH  | `/devices/{id}`                       | auth         | Name + activeEqProfileId only; validate EQ ref. |
| DELETE | `/devices/{id}`                       | auth         | Owner-scoped delete; junction rows cascade. |
| GET    | `/eq-profiles`                        | auth         | `?type=system\|custom` filter; sorted system→custom. |
| POST   | `/eq-profiles`                        | auth         | Create custom profile. |
| GET    | `/eq-profiles/{id}`                   | auth         | Accessible if system or owned. |
| PATCH  | `/eq-profiles/{id}`                   | auth         | System preset → `403 system_preset_immutable`. |
| DELETE | `/eq-profiles/{id}`                   | auth         | System preset → `403 system_preset_immutable`. ON DELETE SET NULL on devices. |
| POST   | `/eq-profiles/{id}/fork`              | auth         | Source must be system; custom owned by caller. |
| GET    | `/party-sessions`                     | auth         | `?status=active\|paused\|ended` filter. |
| POST   | `/party-sessions`                     | auth         | Transactional; validate device ownership. |
| GET    | `/party-sessions/{id}`                | auth         | Owner-scoped lookup. |
| PATCH  | `/party-sessions/{id}`                | auth         | Name + status; enforce legal transitions. |
| DELETE | `/party-sessions/{id}`                | auth         | Owner-scoped delete; junction rows cascade. |
| POST   | `/party-sessions/{id}/devices`        | auth         | Add one device; verify ownership + uniqueness; return full session. |
| DELETE | `/party-sessions/{id}/devices/{deviceId}` | auth     | Remove; return full session. |
| POST   | `/firmware/check`                     | registered   | Hardcoded per-deviceType map. |

Request/response shapes are defined in `openapi.yaml`. Handlers convert sqlc
rows to DTOs via `ToResponse` / `ToListResponse` helpers in each domain's
`dto.go`. No rate limiting for HCI scope.

## Module Structure

```
backend/internal/
├── auth/
│   ├── context.go         (existing)
│   ├── middleware.go      (existing) + add RequireRegistered
│   ├── password.go        (existing)
│   ├── tokens.go          (existing)
│   ├── handler.go         REPLACE stub: /auth/guest, /auth/register, /auth/login
│   ├── dto.go             NEW: GuestResponse, RegisterRequest, LoginRequest, AuthResponse
│   └── service.go         NEW: ConvertGuest (transactional)
│                          NOTE: root CLAUDE.md's example places ConvertGuest in
│                          `package users`, but locating it there would require
│                          `auth.Handler.register` to import `users`, while `users`
│                          already imports `auth` for context accessors — a cycle.
│                          ConvertGuest lives in `auth` because the auth handler is
│                          its only caller. Backend CLAUDE.md's "domains never
│                          import each other" rule is satisfied.
├── users/
│   ├── handler.go         REPLACE stub: /me (GET, PATCH), /me/preferences (GET, PATCH)
│   └── dto.go             NEW: UserResponse, UserUpdateRequest, PreferencesResponse, PreferencesUpdateRequest, ToUserResponse, ToPreferencesResponse
├── devices/
│   ├── handler.go         REPLACE stub: full CRUD; takes defaultEQProfileID
│   ├── dto.go             NEW: CreateRequest, UpdateRequest, Response, ToResponse, ToListResponse
│   └── service.go         NEW: UpdateWithEQValidation (SELECT-then-UPDATE)
├── eqprofiles/
│   ├── bands.go           (existing)
│   ├── handler.go         REPLACE stub: full CRUD + fork
│   ├── dto.go             NEW: CreateRequest, UpdateRequest, ForkRequest, Response, ToResponse, ToListResponse, bandsFromJSON, bandsToJSON
│   └── service.go         NEW: Fork (validates is_system source, inserts custom profile)
├── partysessions/
│   ├── handler.go         REPLACE stub: full CRUD + add/remove device
│   ├── dto.go             NEW: CreateRequest, UpdateRequest, AddDeviceRequest, Response, ToResponse, ToListResponse, deviceIDsFromUUIDs
│   ├── service.go         NEW: Create (tx: verify devices + insert + add devices), UpdateStatus (validates transition in tx)
│   └── transitions.go     NEW: pure func legalTransition(from, to string) bool
├── firmware/
│   ├── handler.go         REPLACE stub: /check. ALSO change NewHandler
│   │                      signature from `NewHandler(cfg config.Config)` to
│   │                      `NewHandler()` — version map is package-local.
│   ├── dto.go             NEW: CheckRequest, CheckResponse
│   └── versions.go        NEW: latestVersion(deviceType string) (string, bool); per-deviceType map
├── httpx/
│   ├── (existing files unchanged)
│   └── optional.go        NEW: Optional[T] generic wrapper for tri-state PATCH fields
├── server/
│   └── router.go          REWRITE: /healthz public, /auth/* public group, everything else under AuthMW
└── db/
    ├── queries/
    │   └── users.sql      NEW
    └── (regenerated)      users.sql.go after sqlc generate
```

Tests:

```
backend/internal/auth/service_test.go         ConvertGuest (testcontainers)
backend/internal/auth/tokens_test.go          JWT sign/verify roundtrip (pure)
backend/internal/auth/password_test.go        Hash/Verify (pure)
backend/internal/eqprofiles/service_test.go   Fork (testcontainers)
backend/internal/partysessions/transitions_test.go  legalTransition (pure)
```

Wiring changes in `cmd/api/main.go`:

```go
defaultEQ, err := queries.GetDefaultEQProfile(ctx)
if err != nil {
    return fmt.Errorf("load default eq profile: %w", err)
}
defaultEQProfileID, _ := uuid.FromBytes(defaultEQ.ID.Bytes[:]) // pgtype.UUID → uuid.UUID

r := server.NewRouter(server.Deps{
    AuthMW:          authMW,
    RequireRegMW:    auth.RequireRegistered,
    AuthHandler:     auth.NewHandler(queries, pool, tokens),
    UserHandler:     users.NewHandler(queries),
    DeviceHandler:   devices.NewHandler(queries, defaultEQProfileID),
    EQHandler:       eqprofiles.NewHandler(queries, pool),
    PartyHandler:    partysessions.NewHandler(queries, pool),
    FirmwareHandler: firmware.NewHandler(),  // cfg no longer needed; versions are in versions.go
})
```

`server.Deps` gains an `AuthHandler` field and a `RequireRegMW` middleware
field. `firmware.NewHandler` no longer takes a Config — the version map is
package-local.

Router shape:

```go
r := chi.NewRouter()
r.Get("/healthz", ...)

r.Route("/auth", deps.AuthHandler.RegisterRoutes)  // public

r.Group(func(r chi.Router) {
    r.Use(deps.AuthMW)
    mount(r, "/users",          deps.UserHandler)      // /me + /me/preferences (preferences inside applies RequireReg)
    mount(r, "/devices",        deps.DeviceHandler)
    mount(r, "/eq-profiles",    deps.EQHandler)
    mount(r, "/party-sessions", deps.PartyHandler)
    mount(r, "/firmware",       deps.FirmwareHandler)  // /check inside applies RequireReg
})
```

`UserHandler.RegisterRoutes`:

```go
r.Get("/me", h.getMe)
r.Patch("/me", h.patchMe)
r.Route("/me/preferences", func(r chi.Router) {
    r.Use(auth.RequireRegistered)
    r.Get("/", h.getPreferences)
    r.Patch("/", h.patchPreferences)
})
```

`FirmwareHandler.RegisterRoutes`:

```go
r.With(auth.RequireRegistered).Post("/check", h.check)
```

## Business Logic

### Guest creation (`POST /auth/guest`)

1. `queries.CreateGuestUser(ctx)` → returns the new row.
2. `tokens.Sign(userID, "guest")` → JWT.
3. `201` with `AuthResponse{user: ToUserResponse(row), token: jwt}`.

### Login (`POST /auth/login`)

1. Decode + validate `LoginRequest`. Lowercase + trim email.
2. `queries.GetUserByEmail(ctx, email)`. `pgx.ErrNoRows` →
   `ErrInvalidCredentials` (mapped to `401 invalid_credentials`).
   (The query only matches rows with a non-NULL email; the schema's CHECK
   constraint guarantees those rows are `role='registered'`, so no extra
   role check is needed.)
3. `auth.VerifyPassword(row.PasswordHash.String, req.Password)`. Failure →
   `ErrInvalidCredentials`.
4. Sign new JWT with role `registered`.
5. `200` with `AuthResponse`.

### Register or convert (`POST /auth/register`)

1. Decode + validate `RegisterRequest`. Lowercase + trim email.
2. Hash the password once (`auth.HashPassword`) — used by both branches below.
3. Inspect `Authorization` header:
   - **Absent**: fresh register flow.
   - **Present**: `tokens.Verify(...)`.
     - Error → `WriteError(401 unauthorized)`.
     - Role `registered` → `WriteError(409 already_registered)`.
     - Role `guest` → convert flow with that guestID.
4. **Fresh register** (no pre-flight EmailExists; rely on the unique index):
   - `queries.CreateRegisteredUser(email, hash, displayName)`.
   - If the insert returns a Postgres `unique_violation` (SQLSTATE `23505`)
     on `users_email_unique_idx`, return `ErrEmailTaken`. Other errors →
     `internal_error`.
   - Sign JWT for the new user as `registered`.
   - `201` with `AuthResponse`.
5. **Convert** (via `auth.ConvertGuest` service — transactional):
   ```go
   tx, _ := pool.Begin(ctx); defer tx.Rollback(ctx)
   qtx := q.WithTx(tx)
   if exists, _ := qtx.EmailExists(ctx, email); exists {
       return ErrEmailTaken
   }
   row, err := qtx.ConvertGuestToRegistered(ctx, guestID, email, hash, dispName)
   if errors.Is(err, pgx.ErrNoRows) {
       return ErrNotFound // guest row disappeared mid-flight
   }
   if isUniqueViolation(err) {
       return ErrEmailTaken // racy collision; unique index caught it
   }
   if err != nil { return err }
   return tx.Commit(ctx)
   ```
   `EmailExists` inside the tx preempts the common case with a clear error;
   the unique-index check is the backstop for the rare concurrent-conversion
   race. After commit, sign a new JWT for `guestID` with role `registered`
   and return `201`.
   `dispName` here is `pgtype.Text{Valid: req.DisplayName.Set, String:
   req.DisplayName.Value}` so the COALESCE in `ConvertGuestToRegistered`
   preserves the guest's existing display name when the field was omitted.

### Users

- `GET /users/me`: `queries.GetUserByID(ctx, UserIDFromContext)` → `ToUserResponse`.
- `PATCH /users/me`: decode + validate; if `displayName` provided, call
  `queries.UpdateUserDisplayName`. If body empty → re-read and return current
  user.
- `GET /users/me/preferences`: `queries.GetUserPreferences`. On
  `pgx.ErrNoRows` return defaults `{theme:"system", language:"en",
  notificationsEnabled:true}` without inserting.
- `PATCH /users/me/preferences`: read current (or defaults), apply patched
  fields, call `UpsertUserPreferences` with the merged values.

### Devices

- `GET /devices`: `ListDevicesByUser(userID)` → `ToListResponse`.
- `POST /devices`: decode + validate. Generate:
  ```go
  firmwareVersion := "1.0.0"
  batteryLevel := 20 + mathrand.Intn(81) // 20..100 inclusive
  isConnected := mathrand.Intn(2) == 1
  ```
  Call `CreateDevice(userID, name, deviceType, firmwareVersion, batteryLevel,
  isConnected, defaultEQProfileID)`. Return `201`. (Uses `math/rand/v2` for
  Go 1.22+ — its `IntN` is fine; the simulation does not need crypto-grade
  randomness.)
- `GET /devices/{id}`: parse UUID; `GetDeviceByIDAndUser`. `pgx.ErrNoRows` →
  `not_found`.
- `PATCH /devices/{id}`: via `devices.UpdateWithEQValidation` service.
  - If patch has non-null `activeEqProfileId`, call `GetAccessibleEQProfile`;
    not found → `ErrInvalidEQProfileRef` (`400`).
  - Read current row (`GetDeviceByIDAndUser`), apply patched fields (keep
    existing values when omitted), call `UpdateDevice`.
  - Empty body returns current row.
- `DELETE /devices/{id}`: `DeleteDevice`. `execrows == 0` → `not_found`.
  Otherwise `204`.

### EQ profiles

- `GET /eq-profiles?type=`:
  - empty → `ListAccessibleEQProfiles(userID)`.
  - `system` → `ListSystemEQProfiles()`.
  - `custom` → `ListCustomEQProfiles(userID)`.
  - other → `400 validation_failed`.
- `POST /eq-profiles`: validate name + bands; marshal bands to JSON;
  `CreateCustomEQProfile(userID, name, bandsJSON)`.
- `GET /eq-profiles/{id}`: `GetAccessibleEQProfile(id, userID)`. Missing →
  `not_found`.
- `PATCH /eq-profiles/{id}`:
  - Look up via `GetAccessibleEQProfile`.
  - If `IsSystem` → `ErrSystemPresetImmutable` (`403`).
  - If empty body → return current.
  - Else `UpdateCustomEQProfile`.
- `DELETE /eq-profiles/{id}`:
  - Look up via `GetAccessibleEQProfile`.
  - If `IsSystem` → `ErrSystemPresetImmutable` (`403`).
  - Else `DeleteCustomEQProfile`. Devices referencing it get
    `active_eq_profile_id` set to NULL by the FK cascade.
- `POST /eq-profiles/{id}/fork` (via `eqprofiles.Fork` service):
  - `GetAccessibleEQProfile(sourceID, userID)`. Missing → `not_found`.
  - If `IsSystem == false` → `ErrNotASystemPreset` (`400`).
  - Marshal supplied bands; `CreateCustomEQProfile(userID, req.Name,
    bandsJSON)`. No tx — single insert.

### Party sessions

- `GET /party-sessions?status=`: empty → `ListPartySessionsByOwner`; else
  `ListPartySessionsByOwnerAndStatus(status)`; invalid status → `400`.
- `POST /party-sessions` (via `partysessions.Create` service):
  ```go
  tx, _ := pool.Begin(ctx); defer tx.Rollback(ctx)
  qtx := q.WithTx(tx)
  count, _ := qtx.CountDevicesByIDsAndUser(ctx, userID, deviceIDs)
  if int(count) != len(deviceIDs) { return ErrInvalidDeviceRef }
  session, _ := qtx.CreatePartySession(ctx, userID, name)
  _, _ = qtx.AddPartySessionDevices(ctx, session.ID, userID, deviceIDs)
  tx.Commit(ctx)
  ```
  Then re-fetch via `GetPartySessionByIDAndOwner` to return the session with
  `device_ids[]`.
- `GET /party-sessions/{id}`: `GetPartySessionByIDAndOwner`.
- `PATCH /party-sessions/{id}` (via `partysessions.UpdateStatus` service when
  status is patched; pure name-only patch can call UpdatePartySession
  directly):
  ```go
  tx, _ := pool.Begin(ctx); defer tx.Rollback(ctx)
  qtx := q.WithTx(tx)
  current, _ := qtx.GetPartySessionByIDAndOwner(ctx, id, userID)
  if !legalTransition(current.Status, req.Status) { return ErrInvalidStatusTransition }
  updated, _ := qtx.UpdatePartySession(ctx, id, userID, mergedName, req.Status)
  tx.Commit(ctx)
  ```
- `POST /party-sessions/{id}/devices`:
  - Verify session exists and is owned: `GetPartySessionByIDAndOwner`.
  - **If `session.Status == "ended"` → `ErrInvalidStatusTransition` (`400`).**
    Ended sessions are frozen; their membership graph cannot change.
  - Verify device is owned by caller: `GetDeviceByIDAndUser`. Missing →
    `ErrInvalidDeviceRef`.
  - `AddPartySessionDevice(sessionID, deviceID, userID)`. PG `unique_violation`
    on the `(party_session_id, device_id)` PK → `ErrDeviceAlreadyInSession`
    (`409`). Other errors → `internal_error`.
  - Return the updated session.
- `DELETE /party-sessions/{id}/devices/{deviceId}`:
  - Verify session exists and is owned: `GetPartySessionByIDAndOwner`.
  - **If `session.Status == "ended"` → `ErrInvalidStatusTransition` (`400`).**
  - `RemovePartySessionDevice`. Always return the (re-fetched) updated
    session as `200` (the openapi spec defines this as a body-returning
    200, not 204).

### Status transition table

```go
// transitions.go
func legalTransition(from, to string) bool {
    switch from {
    case "active":
        return to == "active" || to == "paused" || to == "ended"
    case "paused":
        return to == "paused" || to == "active" || to == "ended"
    case "ended":
        return to == "ended"
    }
    return false
}
```

Same-value transitions (`active→active`) are no-ops and allowed; the UPDATE
still runs and `updated_at` ticks.

### Firmware check

```go
// versions.go
var latest = map[string]string{
    "vespin_classic": "1.0.2",
    "vespin_mini":    "1.0.2",
    "vespin_pro":     "1.1.0",
}
```

Handler:
1. Decode + validate `CheckRequest` (deviceID UUID, currentVersion string).
2. `GetDeviceByIDAndUser(deviceID, userID)`. Missing → `not_found`.
3. `latestForType := latest[device.DeviceType]`. (Type comes from the row;
   the spec's `currentVersion` body field is only echoed back for comparison.)
4. `updateAvailable := req.CurrentVersion != latestForType`.
5. Return `CheckResponse{latestVersion: latestForType, updateAvailable,
   releaseNotes: ""}`.

## Edge Cases & Error Handling

| Case | Behavior |
|------|----------|
| Empty PATCH body for any update endpoint | `200`, current state unchanged. |
| Path UUID malformed | `400 validation_failed`. Handler calls `uuid.Parse(chi.URLParam(r, "id"))`; on error: `httpx.WriteError(w, httpx.NewValidationError(map[string]string{"id": "must be a valid UUID"}, err))`. |
| Resource owned by another user | `404 not_found`. Never `403` — don't leak existence. |
| `pgx.ErrNoRows` from owner-scoped lookups | `404 not_found` (httpx already maps this). |
| EQ profile delete cascades to devices | FK `ON DELETE SET NULL` handles. Frontend invalidates its cache; backend does nothing extra. |
| Device delete cascades to junction rows | FK `ON DELETE CASCADE` handles. No party-session status changes. |
| Adding a device already in this session | Catch pg `unique_violation` (SQLSTATE `23505`) on the junction insert → `409 device_already_in_session`. |
| Adding a device that's in a different session | Allowed. Same device can be in multiple sessions. |
| Adding or removing a device on an `ended` session | `400 invalid_status_transition` (per Requirement 17). Ended sessions are frozen for both status PATCH and membership writes. |
| PATCH /party-sessions/{id} with both name and status | Service path runs the transition check; if legal, both are written in the single UPDATE. |
| PATCH /party-sessions/{id} with `status` omitted | Pure name update; no transition check needed. |
| Concurrent guest conversion + same email | Second tx hits `users_email_unique_idx` → pg `unique_violation` → translate to `ErrEmailTaken`. |
| `Authorization` header on a public route is malformed | Public routes ignore it (`/auth/guest`, `/auth/login`). `/auth/register` rejects malformed/invalid tokens with `401` per Requirement 11. |
| `displayName` set to empty string in PATCH | Validator passes (`omitempty`). Empty string clears the name. Allowed. |
| `theme`/`language`/`notifications` upsert when row missing | Patch values layered over defaults (`theme=system, language=en, notificationsEnabled=true`), then `UpsertUserPreferences` inserts the row. |
| `firmware/check` for a deviceType not in the map | Defensive: return `internal_error`. The enum is closed; this shouldn't happen unless schema drifts. |

Logging:
- `Info`: server starting, guest created, user registered, guest converted,
  device created, party session created, party session ended,
  custom EQ profile created.
- `Warn`: unexpected pg error codes that fall through to `internal_error`
  (so they're visible without the `Error` level noise of every 5xx response).
- `Error`: any 5xx path. Never `Error` for 4xx.
- Never log `password`, `password_hash`, JWT contents, or full request/response
  bodies.

## Security Considerations

- **IDOR**: every owner-scoped query takes `user_id` and filters by it.
  Handlers always source `userID` from `auth.UserIDFromContext`, never from
  request body or path. Cross-user reads return `404`, not `403`, so existence
  doesn't leak.
- **Role gating**: `auth.RequireRegistered` middleware is the single
  enforcement point for guest-forbidden endpoints. Adding a new
  registered-only route is one `r.Use(auth.RequireRegistered)` line.
- **Password storage**: bcrypt default cost. Plaintext password leaves the
  process only via the hashed column. Never logged. Comparing during login
  uses `bcrypt.CompareHashAndPassword`.
- **JWT**: HS256 with `JWT_SECRET` (envRequired). Tokens carry `sub`
  (uuid), `role`, `iat`, `exp`. Verify enforces `WithExpirationRequired` and
  `WithValidMethods=[HS256]` — already in place. No refresh tokens.
- **Email handling**: trim + lowercase before all email queries
  (`CreateRegisteredUser`, `ConvertGuestToRegistered`, `GetUserByEmail`,
  `EmailExists`). The unique index is case-sensitive; lowercasing at the
  application layer is sufficient and avoids changing the index.
- **Unknown fields in JSON**: `DecodeJSON` already uses
  `DisallowUnknownFields` + 1MB body cap. No changes needed.
- **UUID parsing**: all `{id}` and `{deviceId}` path params go through
  `uuid.Parse`; failure → `validation_failed`.
- **Active EQ profile reference**: extra application-layer check (Requirement
  10) catches the cross-user case that the FK alone cannot.
- **Party session device add**: ownership of both the session and the device
  is verified before insertion — no IDOR through device IDs.

## Testing Plan

All tests live in `*_test.go` alongside the code they exercise. The two DB
tests use `testcontainers-go` to spawn a short-lived Postgres, run all
migrations, and tear down at the end of the package. No mocks of sqlc.

### Guest-to-registered conversion (`internal/auth/service_test.go`)

```
Given a guest user with attached devices, EQ profiles, and a party session,
when ConvertGuest is called with a fresh email + password,
then:
  - the same user_id is returned with role=registered, email set, converted_at non-null
  - device rows, EQ profile rows, and party_session rows still reference that user_id
  - GetUserByEmail returns the converted user

Given a guest user,
when ConvertGuest is called with an email that already belongs to another registered user,
then:
  - the function returns ErrEmailTaken
  - the guest row is unchanged (still role=guest, email=null, converted_at=null)
  - the user can still convert with a different email afterwards

Given a registered user attempting ConvertGuest on their own ID,
when called,
then it returns ErrNotFound (UPDATE matched zero rows because role != 'guest').
```

### EQ profile fork (`internal/eqprofiles/service_test.go`)

```
Given a seeded system preset and an authenticated user,
when Fork is called with a name and modified bands,
then:
  - a new row is created with is_system=false, owner_user_id=user, name+bands as supplied
  - the source preset is unchanged
  - the returned profile id is different from the source

Given a custom profile owned by user A,
when user A calls Fork on it,
then ErrNotASystemPreset is returned and no row is created.

Given a system preset and any authenticated user (no ownership relation),
when Fork is called,
then it succeeds — system presets are not owner-scoped.

Given a system preset that does not exist,
when Fork is called,
then ErrNotFound is returned.
```

### JWT roundtrip (`internal/auth/tokens_test.go`)

```
Given a Tokens with a fixed secret,
when Sign then Verify the resulting token,
then the parsed Claims.UserID and Claims.Role match the inputs.

When verifying a token signed with a different secret,
then Verify returns an error wrapping ErrUnauthorized.

When verifying an expired token,
then Verify returns an error wrapping ErrUnauthorized.

When verifying a token signed with RS256 (or a missing alg),
then Verify rejects it.

When signing with an unknown role,
then Sign returns an error.
```

### Password hash (`internal/auth/password_test.go`)

```
Given a plaintext password,
when HashPassword then VerifyPassword,
then VerifyPassword returns nil.

When VerifyPassword is called with the wrong password,
then it returns an error wrapping ErrInvalidCredentials.

When HashPassword is called twice on the same input,
then the two hashes differ (bcrypt salt) but both verify.
```

### Status transitions (`internal/partysessions/transitions_test.go`)

```
Table:
  active  -> active  : true
  active  -> paused  : true
  active  -> ended   : true
  paused  -> active  : true
  paused  -> paused  : true
  paused  -> ended   : true
  ended   -> active  : false
  ended   -> paused  : false
  ended   -> ended   : true
  unknown -> active  : false
  active  -> unknown : false
```

## Open Questions

1. **firmware/check `currentVersion` body field.** Currently only the
   device's persisted `firmware_version` could be used. The spec passes
   `currentVersion` from the client; we trust it as the comparison input.
   Slightly redundant with `device.firmware_version` but matches the contract.
2. **Login email enumeration via timing.** Email-not-found returns before
   bcrypt runs; email-found-wrong-password runs bcrypt. The difference lets
   an attacker enumerate registered emails by response timing. Out of scope
   for HCI per CLAUDE.md, but if hardened later, the fix is to always run
   bcrypt against a dummy hash on the "no such user" path.
3. **CORS.** Not in scope per CLAUDE.md silence. If the frontend hits the API
   from a non-native context (Expo web preview), add a permissive CORS
   middleware later in a separate PR.
4. **Health/readiness `/readyz`.** Explicitly deferred per `backend/CLAUDE.md`.
   Not in this spec.
