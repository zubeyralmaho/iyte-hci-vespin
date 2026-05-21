# CLAUDE.md — Root

Constraints for Claude Code agents operating anywhere in this repository.
Scoped conventions live in `backend/CLAUDE.md` and `frontend/CLAUDE.md`.
Both inherit the rules in this file.

---

## Project identity

- The project name is **Vespin**.
- Module names, image names, container names, hostnames use lowercase **`vespin`**.
- Image names: **`vespin-api`** and **`vespin-migrate`**.
- The product line of speakers this app companions is the **Vespin Retro** series.

## What this project is

This is a **university HCI course project**. It is a companion app for a
fictional smart-speaker product line. The deliverable is a working mobile app
with a real backend, real database, and real authentication.

## What this project is NOT (hard scope boundaries)

The hardware interaction is **fully simulated**. The following are explicitly
out of scope and MUST NOT be added, suggested, or scaffolded for:

- **Real Bluetooth or BLE integration.** Pairing, discovery, GATT services,
  any `react-native-ble-*` library. Devices are paired via a form, not radio.
- **Real audio streaming or playback.** No audio APIs, no streaming protocols,
  no codec work.
- **Real EQ audio processing.** EQ bands are integers in a database. The
  backend never applies them to a signal because there is no signal.
- **Real firmware OTA.** The firmware check endpoint compares strings. There
  is no binary delivery, no update apply, no rollback.
- **Real speaker communication.** No mDNS, no UDP discovery, no socket
  connections to physical devices. `isConnected` is a column in a table.
- **OAuth backend implementation.** Google/Apple OAuth buttons exist in the
  UI for the HCI deliverable but the backend implements email/password only.
  Do not implement OAuth on the backend even if asked.
- **Concurrency for "party mode".** No goroutines for audio sync, no
  WebSockets, no real-time anything. Party sessions are state in Postgres.
- **WebSockets, Server-Sent Events, or any push channel.** REST polling only.
- **Push notifications.** Out of scope.

When a feature request would require any of the above, push back. State that
it is out of scope per `CLAUDE.md` and propose a simulated alternative.

## Simulated fields

The following fields in the data model are **simulated** — the backend
seeds them at creation time and never derives them from real sources:

| Field | Source |
|---|---|
| `devices.firmware_version` | Static string seeded on create |
| `devices.battery_level` | Pseudo-random integer 20–100 seeded on create |
| `devices.is_connected` | Pseudo-random boolean seeded on create |

These fields are real columns in the database with real values. The
"simulation" is that they are never updated from a real device — there is no
real device.

## The OpenAPI spec is the source of truth

The HTTP contract lives at **`backend/api/openapi.yaml`**.

- The backend implements the spec by hand. Server-side codegen is NOT used.
- The frontend's `src/api/generated/` directory is produced by Orval against
  the spec.
- **Never edit `frontend/src/api/generated/` by hand.** Regenerate with
  `pnpm orval` instead. CI enforces this with a drift check.
- When changing API behavior, change the spec first, then implement.

## Monorepo structure

```
vespin/
├── backend/          Go API
├── frontend/         Expo + React Native app
├── deploy/           Production stack (compose, Caddyfile, scripts)
└── .github/workflows/  CI/CD
```

Each directory is independently developed but shares the repo. Cross-cutting
changes (spec + backend + frontend together) belong in one PR.

## Tech stack — locked decisions

These were debated and locked. Do not propose alternatives without strong cause.

**Backend:**
- Go 1.23 + chi router + sqlc + pgx + Postgres 16
- NOT Gin, NOT GORM, NOT Ent, NOT Bun
- stdlib `log/slog`, NOT zap, NOT zerolog
- caarlos0/env for config, NOT Viper
- golang-jwt, golang-migrate, bcrypt
- Project layout: `cmd/api` + `internal/<domain>` packages

**Frontend:**
- Expo managed workflow + Expo Router (file-based)
- NativeWind for styling, NOT Tamagui, NOT styled-components, NOT Restyle
- TanStack Query for server state, NOT RTK Query
- Zustand for client state, NOT Redux, NOT Jotai
- React Hook Form + Zod for forms
- Orval for API codegen, NOT openapi-typescript-codegen, NOT hey-api

**Infra:**
- Hetzner VPS + Docker Compose + Caddy
- NOT Kubernetes, NOT Traefik, NOT nginx (for this project)
- GitHub Container Registry for images
- DuckDNS for domain (HCI scope)
- Frontend has no automated distribution; daily dev is Expo Go over LAN

## Conventions that apply across the repo

### Naming

- `snake_case` in Postgres (tables, columns)
- `camelCase` in JSON request/response bodies (translated via Go struct tags)
- `snake_case` for Go identifiers where Go convention is `snake_case`
  (it isn't — Go is camelCase/PascalCase). The translation point is the JSON
  tags.
- `kebab-case` for URLs (`/eq-profiles`, `/party-sessions`)
- `kebab-case` for filenames in non-component contexts
- `PascalCase` for React component files

### Timestamps

- Always `timestamptz` in Postgres, NEVER plain `timestamp`
- Always UTC at the database layer
- Always RFC 3339 strings in JSON (`2026-05-18T14:30:00Z`)
- Never Unix epoch integers in API payloads

### IDs

- Always UUID. Generated by Postgres via `gen_random_uuid()` as a column default.
- Surfaced as strings in JSON (with `format: uuid` in the spec).
- Never auto-increment integers.

### Errors

- Backend returns `{ "error": { "code": "<snake_case>", "message": "..." } }`.
- The vocabulary of error codes is defined in `backend/api/openapi.yaml`.
  Do not invent new codes ad-hoc — add them to the spec first.

### Auth

- JWT bearer tokens, ~30-day expiry, no refresh tokens.
- Two roles: `guest` and `registered`.
- Guests are real users in the `users` table with `role='guest'`.
- Guest-to-registered conversion is an UPDATE on the same `users` row, not a
  new account creation. The user's ID never changes. Devices and EQ profiles
  remain attached because they reference the same user_id.

## Migrations

- Migrations are raw SQL in `backend/internal/db/migrations/`.
- Files are numbered: `000001_description.up.sql`, `000001_description.down.sql`.
- **NEVER modify a committed migration.** Always add a new one to fix state.
  This rule has no exceptions. If a previous migration was wrong, write a
  corrective migration with a higher number.
- System EQ presets are seeded by a migration, not by application code.
- Migrations run in a one-shot container at deploy time, before the API starts.

## Testing posture

- Light. This is an HCI course project, not a production system.
- Table-driven tests for the few pieces with real logic:
  - Guest-to-registered conversion
  - EQ profile copy-on-write (fork)
  - JWT signing and verification
  - Password hashing
- Do **NOT** propose:
  - Integration test infrastructure setup unless asked
  - Coverage targets or coverage gates
  - End-to-end testing setups
  - Snapshot tests on the frontend
  - Mocking the database — if you write a test that needs the DB, set up a
    short-lived Postgres container, don't mock sqlc.

## Documentation hygiene

- Keep CLAUDE.md files in sync with reality. If a decision changes, update
  the relevant CLAUDE.md in the same PR.
- README is for humans. CLAUDE.md is for agents. Do not duplicate large
  sections between them — link instead.
- Comments in code should explain *why*, not *what*. The code already says
  what it does.

## Pull requests

- One concern per PR. A spec change + backend implementation + frontend
  regen is one concern (they're inseparable). A spec change + an unrelated
  styling fix is two PRs.
- PR titles use conventional commit prefixes: `feat:`, `fix:`, `chore:`,
  `docs:`, `refactor:`. No body required if the diff is small.
- All status checks must be green before merge (enforced by branch protection).

## What to do when you're uncertain

- If a decision conflicts with this document, **this document wins**. If
  you believe the document is wrong, propose a change to the document first
  in a separate PR.
- If a request would require something on the "NOT in scope" list, push back.
- If you don't know which library or pattern to use, check the scoped
  `CLAUDE.md` in the relevant directory. If it's not there, ask before
  scaffolding — don't guess.

## Glossary

- **Guest user** — anonymous user with `role='guest'`, no email/password,
  full access except preferences and firmware check.
- **Registered user** — `role='registered'`, has email/password, full access
  to all endpoints.
- **Guest conversion** — a guest gaining email/password, becoming registered.
  Implemented as `UPDATE users SET ...` on the same row.
- **System preset** — read-only EQ profile (`is_system=true`, `owner_user_id=null`).
  Seeded by migration. Users cannot modify; they fork instead.
- **Custom profile** — user-owned EQ profile (`is_system=false`,
  `owner_user_id=<user>`). Fully mutable by the owner.
- **Fork (copy-on-write)** — creating a new custom profile based on a system
  preset's bands. The original preset is never modified.
- **Party session** — a state record in `party_sessions` linking 1+ devices.
  No real audio sync; just CRUD on session state.
