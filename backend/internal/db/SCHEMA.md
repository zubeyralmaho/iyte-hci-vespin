# Backend Database Schema

This document records the locked database model and the migration order. The
OpenAPI contract remains the API source of truth; this file is the database
shape behind that contract.

## Conventions

- PostgreSQL column names use `snake_case`.
- JSON request and response fields use `camelCase`.
- Primary keys are UUIDs generated with `gen_random_uuid()`.
- Timestamps use `timestamptz`.
- Raw SQL migrations live in `internal/db/migrations`.
- sqlc queries live in `internal/db/queries`.
- Cross-owner rules are enforced in service logic, not SQL constraints, when
  the rule depends on the authenticated user.

## Implemented Migrations

### 000001_create_users

`users`

- `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
- `role text NOT NULL CHECK (role IN ('guest', 'registered'))`
- `email text`
- `password_hash text`
- `display_name text`
- `created_at timestamptz NOT NULL DEFAULT now()`
- `converted_at timestamptz`

Rules:

- Guests must have `email`, `password_hash`, and `converted_at` set to `NULL`.
- Registered users must have `email` and `password_hash`.
- `users_email_unique_idx` is a partial unique index on `email` where email is
  not null.

### 000002_create_user_preferences

`user_preferences`

- `user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE`
- `theme text NOT NULL DEFAULT 'system' CHECK (theme IN ('light', 'dark', 'system'))`
- `language text NOT NULL DEFAULT 'en' CHECK (language ~ '^[a-z]{2}$')`
- `notifications_enabled boolean NOT NULL DEFAULT true`
- `created_at timestamptz NOT NULL DEFAULT now()`
- `updated_at timestamptz NOT NULL DEFAULT now()`

Notes:

- Preferences are registered-only at the API/service layer.
- The GET handler should return defaults when no row exists yet.
- The PATCH handler should upsert.

### 000003_create_eq_profiles

`eq_profiles`

- `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
- `owner_user_id uuid REFERENCES users(id) ON DELETE CASCADE`
- `name text NOT NULL CHECK (char_length(trim(name)) BETWEEN 1 AND 100)`
- `is_system boolean NOT NULL DEFAULT false`
- `bands jsonb NOT NULL`
- `created_at timestamptz NOT NULL DEFAULT now()`
- `updated_at timestamptz NOT NULL DEFAULT now()`

Rules:

- System presets have `is_system = true` and `owner_user_id IS NULL`.
- Custom profiles have `is_system = false` and `owner_user_id IS NOT NULL`.
- System preset names are unique case-insensitively.
- Custom profiles are indexed by `(owner_user_id, created_at DESC)`.

### 000004_seed_system_eq_profiles

System presets are seeded as:

- `Flat`
- `Bass Boost`
- `Rock`
- `Jazz`
- `Classical`
- `R&B`

### 000005_add_default_eq_profile

Adds `eq_profiles.is_default boolean NOT NULL DEFAULT false`.

Rules:

- Only system presets can be default.
- At most one row can have `is_default = true`.
- The migration marks `Flat` as the default system preset.

### 000006_create_devices

`devices`

- `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
- `user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `name text NOT NULL CHECK (char_length(trim(name)) BETWEEN 1 AND 100)`
- `device_type text NOT NULL CHECK (device_type IN ('vespin_classic', 'vespin_mini', 'vespin_pro'))`
- `firmware_version text NOT NULL`
- `battery_level integer NOT NULL CHECK (battery_level BETWEEN 0 AND 100)`
- `is_connected boolean NOT NULL`
- `active_eq_profile_id uuid REFERENCES eq_profiles(id) ON DELETE SET NULL`
- `paired_at timestamptz NOT NULL DEFAULT now()`
- `created_at timestamptz NOT NULL DEFAULT now()`
- `updated_at timestamptz NOT NULL DEFAULT now()`

Rules:

- Device type values are locked to `vespin_classic`, `vespin_mini`, and
  `vespin_pro`.
- `active_eq_profile_id` must reference either a system profile or a custom
  profile owned by the same user. Enforce this in service logic before update.

### 000007_create_party_sessions

Creates `party_sessions` and `party_session_devices` together.

`party_sessions`

- `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
- `owner_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `name text CHECK (name IS NULL OR char_length(trim(name)) BETWEEN 1 AND 100)`
- `status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'ended'))`
- `started_at timestamptz NOT NULL DEFAULT now()`
- `ended_at timestamptz`
- `created_at timestamptz NOT NULL DEFAULT now()`
- `updated_at timestamptz NOT NULL DEFAULT now()`

Rules:

- Valid status transitions are `active -> paused`, `active -> ended`,
  `paused -> active`, and `paused -> ended`.
- Ended sessions should have `ended_at` set; active and paused sessions should
  not. Enforce state transitions in service logic.

### party_session_devices

`party_session_devices`

- `party_session_id uuid NOT NULL REFERENCES party_sessions(id) ON DELETE CASCADE`
- `device_id uuid NOT NULL REFERENCES devices(id) ON DELETE CASCADE`
- `joined_at timestamptz NOT NULL DEFAULT now()`
- `PRIMARY KEY (party_session_id, device_id)`

Rules:

- A device can only be added to sessions owned by the same user. Enforce this
  in service logic before insert.
- Adding a duplicate device returns `device_already_in_session`.

## EQ Bands

EQ settings are stored as one logical JSONB value:

```json
{
  "subBass": 0,
  "bass": 0,
  "mid": 0,
  "treble": 0,
  "presence": 0
}
```

Rules:

- Shape is a five-band object.
- Values are integers from `-12` to `12`.
- Validation belongs in the Go application layer via a typed struct.
- We do not split bands into columns because the app never queries across
  bands, and the five-band shape should remain evolvable.

## No Database Table

Firmware checks are simulated from application constants. No firmware table is
needed for this project scope.

## Migration and sqlc Order

Finish one domain at a time:

1. Migration up/down files.
2. sqlc query file.
3. `sqlc generate`.
4. `go build ./...`.
5. `go test ./...`.
6. Docker Compose migration smoke test.
