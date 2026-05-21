# Vespin

Companion mobile app for the **Vespin Retro** series Bluetooth smart speakers.

University HCI course project. **Hardware interaction is fully simulated** —
no real Bluetooth, audio processing, firmware OTA, or device communication.
The focus is a functional mobile app with a real backend, real database, and
real auth, minus the embedded/IoT complexity.

---

## What's in this repo

```
vespin/
├── backend/            Go API (chi + sqlc + Postgres)
│   └── api/openapi.yaml   The HTTP contract — source of truth
├── frontend/           Expo + React Native app
├── deploy/             Docker Compose + Caddy + backup script for the VPS
├── .github/workflows/  CI/CD pipelines
├── CLAUDE.md           Constraints for Claude Code agents (root)
├── README.md           This file
└── .gitignore
```

Each major directory has its own `CLAUDE.md` with scoped conventions.

## Tech stack at a glance

**Backend**
- Go 1.23, [chi](https://github.com/go-chi/chi) for HTTP routing
- [sqlc](https://sqlc.dev) for typed queries from raw SQL (NOT GORM)
- [golang-migrate](https://github.com/golang-migrate/migrate) for SQL-first migrations
- [pgx](https://github.com/jackc/pgx) as the Postgres driver
- stdlib `log/slog` for structured logging
- [golang-jwt](https://github.com/golang-jwt/jwt) for auth tokens
- [go-playground/validator](https://github.com/go-playground/validator) for request validation

**Frontend**
- [Expo](https://expo.dev) (managed workflow)
- [Expo Router](https://docs.expo.dev/router/introduction/) for file-based routing
- [NativeWind](https://www.nativewind.dev) (Tailwind for React Native)
- [TanStack Query](https://tanstack.com/query) for server state
- [Zustand](https://zustand-demo.pmnd.rs) for client state
- [React Hook Form](https://react-hook-form.com) + [Zod](https://zod.dev) for forms
- [Orval](https://orval.dev) for OpenAPI-driven API client codegen

**Infra**
- VPS, Ubuntu 24.04
- Docker Compose for orchestration
- Caddy for reverse proxy + automatic Let's Encrypt TLS
- DuckDNS for the public domain
- GitHub Container Registry for image hosting

## Local development

### Prerequisites

- Docker & Docker Compose
- Go 1.23+
- Node.js 20+ and pnpm 9+
- (Optional) [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) CLI for query regeneration

### Backend

```bash
cd backend
cp .env.example .env       # local secrets (gitignored)
docker compose up          # starts api + postgres + runs migrations
```

API listens on `http://localhost:8080`. Hit `GET /healthz` to verify it's up.

Regenerating SQL types after editing `internal/db/queries/*.sql`:

```bash
cd backend && sqlc generate
```

Adding a new migration:

```bash
cd backend/internal/db/migrations
# Create both files; replace XXXXXX with the next 6-digit number
touch 0000XX_describe_change.up.sql 0000XX_describe_change.down.sql
# Edit, then restart the migrate container:
docker compose -f docker-compose.yml run --rm migrate
```

### Frontend

```bash
cd frontend
pnpm install
pnpm orval                 # regenerate API client from openapi.yaml
pnpm start                 # opens Expo dev server, scan QR with Expo Go
```

Set `EXPO_PUBLIC_API_URL` in `frontend/.env.local`:

```
EXPO_PUBLIC_API_URL=http://<your-lan-ip>:8080
```

`<your-lan-ip>` because the phone running Expo Go is on the same network, not
localhost. `localhost` only works in iOS simulator.

### OpenAPI spec

The HTTP contract lives at `backend/api/openapi.yaml`. **It is the source of
truth.** Backend handlers implement it; frontend types are generated from it.

When you edit the spec:

1. Validate: `cd frontend && npx @redocly/cli@latest lint ../backend/api/openapi.yaml`
2. Regenerate the frontend client: `cd frontend && pnpm orval`
3. Commit both the spec change and the regenerated `frontend/src/api/generated/` together.

CI enforces this: PRs that change the spec without regenerating the client
will fail the **Orval drift check** job.

## Deployment

The full deploy guide lives in [`deploy/README.md`](./deploy/README.md). In
summary: a push to `main` triggers GitHub Actions, which builds the API and
migrate images, pushes them to GHCR, and SSHes into the VPS to pull and
restart the stack. Caddy handles TLS automatically.

The frontend has no automated distribution. Daily development runs in Expo
Go against the API on your LAN.

## Working with Claude Code in this repo

This codebase is designed to be agent-friendly. Each major directory has a
`CLAUDE.md` file describing the conventions and constraints for that area.
Read them in this order:

1. [`CLAUDE.md`](./CLAUDE.md) — root, cross-cutting rules (project scope, naming)
2. [`backend/CLAUDE.md`](./backend/CLAUDE.md) — Go conventions
3. [`frontend/CLAUDE.md`](./frontend/CLAUDE.md) — RN/Expo conventions

The single most important thing to know: **this project simulates hardware**.
Do not propose real Bluetooth integration, real audio processing, real
firmware updates, real WebSockets, or real device communication. See the
root `CLAUDE.md` for the full scope boundaries.

## License

Course project. No license granted for redistribution.
