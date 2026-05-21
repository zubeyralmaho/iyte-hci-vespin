# Data Model

This document provides an overview of the database design. For the full schema
with column definitions and constraints, see
[`backend/internal/db/SCHEMA.md`](../backend/internal/db/SCHEMA.md).

## Entity Relationship Diagram

```text
┌──────────────┐       ┌────────────────────┐
│    users     │       │  user_preferences  │
│              │◄──────│  (1:1, optional)   │
│  id (PK)    │       │  user_id (FK, PK)  │
│  role        │       └────────────────────┘
│  email       │
│  ...         │       ┌────────────────────┐
│              │◄──────│     devices        │
└──────┬───────┘       │  id (PK)          │
       │               │  user_id (FK)     │
       │               │  active_eq_...    │──┐
       │               └────────────────────┘  │
       │                                       │
       │               ┌────────────────────┐  │
       └──────────────►│   eq_profiles      │◄─┘
                       │  id (PK)          │
                       │  owner_user_id    │
                       │  is_system        │
                       └────────────────────┘

┌──────────────┐       ┌──────────────────────────┐
│party_sessions│◄──────│  party_session_devices   │
│  id (PK)    │       │  party_session_id (FK)   │
│  owner_...  │       │  device_id (FK)          │
└──────────────┘       └──────────────────────────┘
```

## Tables

### users

The central identity table. Supports two roles:

- **Guest** (`role='guest'`): No email/password, anonymous access.
- **Registered** (`role='registered'`): Has email/password, full feature access.

Guest-to-registered conversion is an in-place `UPDATE` — the user ID never
changes, so all associated data (devices, profiles, sessions) remains attached.

### user_preferences

One-to-one with `users`. Stores theme, language, and notification settings.
Only accessible by registered users. The API returns defaults when no row exists.

### eq_profiles

Two types coexist in one table:

- **System presets** (`is_system=true`, `owner_user_id=NULL`): Read-only,
  seeded by migrations. Users fork these to create editable copies.
- **Custom profiles** (`is_system=false`, `owner_user_id=<uuid>`): Fully
  mutable by the owning user.

EQ bands are stored as a JSONB object with five integer fields (`subBass`,
`bass`, `mid`, `treble`, `presence`), each ranging from -12 to +12.

### devices

Paired speakers belonging to a user. Device types are constrained to:
`vespin_classic`, `vespin_mini`, `vespin_pro`.

Each device can have an `active_eq_profile_id` pointing to either a system
preset or a custom profile owned by the same user.

Simulated fields (never updated from real hardware):
- `firmware_version`
- `battery_level`
- `is_connected`

### party_sessions

Multi-device listening sessions. Status flow:

```text
active ──► paused ──► ended
  │                     ▲
  └─────────────────────┘
```

Only the session owner can modify status or membership.

### party_session_devices

Junction table linking devices to party sessions. A device can only join
sessions owned by the same user.

## Conventions

| Convention | Rule |
|-----------|------|
| Primary keys | UUID via `gen_random_uuid()` |
| Timestamps | Always `timestamptz`, never plain `timestamp` |
| Column naming | `snake_case` |
| JSON naming | `camelCase` (via Go struct tags) |
| Soft deletes | Not used — hard deletes with `ON DELETE CASCADE` |

## Migrations

Migrations are raw SQL files in `backend/internal/db/migrations/` with the
naming pattern:

```
000001_description.up.sql
000001_description.down.sql
```

**Rules:**
- Never modify a committed migration. Always create a new one.
- System presets are seeded by migrations, not application code.
- Migrations run in a one-shot container before the API starts.

## sqlc

Database queries are written in raw SQL at `backend/internal/db/queries/` and
compiled to type-safe Go code by sqlc. After editing query files:

```bash
cd backend
sqlc generate
```
