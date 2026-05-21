# Vespin

This project is a collaborative initiative developed for an HCI course at the Izmir Institute of Technology, bringing together students from the Department of Industrial Design and the Department of Computer Engineering.

Vespin is a companion mobile app concept for the **Vespin Retro** series Bluetooth smart speakers. This project focuses on the mobile interaction model, supported by a fully functional API, database, and authentication flow.

Please note: Hardware interaction is simulated. There is no real-time Bluetooth connectivity, audio processing, firmware OTA, or physical device communication; the focus remains on the user experience and the backend architecture.

---

## What's in this repo

```text
vespin/
|-- backend/            Go API (chi + sqlc + Postgres)
|   `-- api/openapi.yaml
|                       HTTP contract and frontend codegen source
|-- frontend/           Expo + React Native app
|-- deploy/             Production compose, Caddy config, and backup script
|-- .github/workflows/  CI, OpenAPI drift checks, and guarded deploy workflow
|-- specs/              Project notes and design-system documentation
|-- CLAUDE.md           Repository conventions for coding agents
|-- README.md           This file
`-- .gitignore
```

Each major directory has its own `CLAUDE.md` with scoped conventions.

## Scope

The app models the expected user-facing flows for a smart speaker companion app:

- Authentication, guest access, and account preferences
- Device listing and simulated device management
- EQ profile browsing, creation, editing, and forking
- Party session creation and management
- Simulated firmware check behavior

The hardware-facing parts are intentionally mocked for the HCI scope. The code is
structured so those integrations could be added later, but they are not part of
this course project.

## Tech stack at a glance

**Backend**

- Go 1.23
- [chi](https://github.com/go-chi/chi) for HTTP routing
- [sqlc](https://sqlc.dev) for typed queries from raw SQL
- [golang-migrate](https://github.com/golang-migrate/migrate) for SQL migrations
- [pgx](https://github.com/jackc/pgx) as the Postgres driver
- stdlib `log/slog` for structured logging
- [golang-jwt](https://github.com/golang-jwt/jwt) for auth tokens
- [go-playground/validator](https://github.com/go-playground/validator) for request validation

**Frontend**

- [Expo](https://expo.dev) with React Native
- [Expo Router](https://docs.expo.dev/router/introduction/) for file-based routing
- [NativeWind](https://www.nativewind.dev) for Tailwind-style React Native styling
- [TanStack Query](https://tanstack.com/query) for server state
- [Zustand](https://zustand-demo.pmnd.rs) for client state
- [React Hook Form](https://react-hook-form.com) and [Zod](https://zod.dev) for forms
- [Orval](https://orval.dev) for OpenAPI-driven API client generation

**Infrastructure**

- Docker Compose for local and production stacks
- Caddy config for reverse proxy and TLS in the production stack
- GitHub Actions for backend checks, OpenAPI validation, generated-client drift
  checks, and guarded backend image deployment
- GitHub Container Registry support for backend and migration images

## Local development

### Prerequisites

- Docker and Docker Compose
- Node.js 20+
- pnpm 9+
- Expo Go on a phone for physical-device testing
- Optional: Go 1.23+ for backend tests and local tooling
- Optional: [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) CLI for query regeneration

### Backend

```bash
cd backend
docker compose up
```

The API listens on `http://localhost:8080`. Use `GET /healthz` to verify it is
running. You do not need Go installed just to run the Docker-based local stack.

Regenerate SQL types after editing `internal/db/queries/*.sql`:

```bash
cd backend
sqlc generate
```

Add a new migration:

```bash
cd backend/internal/db/migrations
touch 0000XX_describe_change.up.sql 0000XX_describe_change.down.sql
cd ../../..
docker compose run --rm migrate
```

Replace `0000XX` with the next migration number before committing.

### Frontend

```bash
cd frontend
pnpm install
cp .env.example .env.local
pnpm codegen
pnpm start
```

Set `EXPO_PUBLIC_API_URL` in `frontend/.env.local`:

```text
EXPO_PUBLIC_API_URL=http://<your-lan-ip>:8080
```

Use your LAN IP because a phone running Expo Go is on the network, not inside
your computer's localhost. `localhost` only works for local simulators.

`EXPO_PUBLIC_*` values are bundled into the app, so only put public
configuration there. API base URLs are fine; secrets, database passwords, and
JWT signing keys are not.

### OpenAPI spec

The HTTP contract lives at `backend/api/openapi.yaml`. It is the source of truth
for both backend behavior and frontend generated types.

When you edit the spec:

1. Validate it: `cd frontend && npx @redocly/cli@latest lint ../backend/api/openapi.yaml`
2. Regenerate the frontend client: `cd frontend && pnpm codegen`
3. Commit the spec change and regenerated `frontend/src/api/generated/` files together.

CI enforces this with the OpenAPI validation and Orval drift-check workflow.

## GitHub workflow and deployment

The repository currently has GitHub Actions for:

- PR checks with path filtering for backend and frontend changes
- OpenAPI validation and generated-client drift detection
- Backend build checks on pushes to `main`
- Optional backend image build, push, and VPS deployment through
  `.github/workflows/backend.yml`

Deployment is prepared but intentionally not the focus of this HCI submission.
The deploy workflow is guarded by the `DEPLOY_ENABLED` repository variable or a
manual `workflow_dispatch` run. When enabled, it builds backend and migration
images, pushes them to GitHub Container Registry, syncs the `deploy/` directory
to the VPS, restarts the Docker Compose stack, and checks `/healthz`.

## Working with Claude Code in this repo

This codebase is designed to be agent-friendly. Each major directory has a
`CLAUDE.md` file describing conventions and constraints for that area:

1. [`CLAUDE.md`](./CLAUDE.md) - root project scope and cross-cutting rules
2. [`backend/CLAUDE.md`](./backend/CLAUDE.md) - Go/backend conventions
3. [`frontend/CLAUDE.md`](./frontend/CLAUDE.md) - React Native/Expo conventions

The most important scope rule: this project simulates hardware. Do not add real
Bluetooth integration, real audio processing, real firmware updates, real
WebSockets, or real device communication unless the project scope changes.

## License

Course project. No license granted for redistribution.
