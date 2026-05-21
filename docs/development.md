# Development Guide

## Prerequisites

| Tool | Version | Required |
|------|---------|----------|
| Docker + Docker Compose | Latest | Yes |
| Node.js | 20+ | Yes |
| pnpm | 9+ | Yes |
| Expo Go (mobile) | Latest | Yes (for device testing) |
| Go | 1.23+ | Optional (for local backend tests) |
| sqlc CLI | Latest | Optional (for query regeneration) |

## Quick Start

### Backend

```bash
cd backend
docker compose up
```

The API starts at `http://localhost:8080`. Verify with:

```bash
curl http://localhost:8080/healthz
```

You do not need Go installed — Docker handles the build.

### Frontend

```bash
cd frontend
pnpm install
cp .env.example .env.local
pnpm codegen
pnpm start
```

Set `EXPO_PUBLIC_API_URL` in `.env.local` to your LAN IP:

```
EXPO_PUBLIC_API_URL=http://<your-lan-ip>:8080
```

Use your LAN IP (not `localhost`) because Expo Go runs on a separate device.

## Common Tasks

### Add a Database Migration

```bash
cd backend/internal/db/migrations
touch 0000XX_describe_change.up.sql 0000XX_describe_change.down.sql
```

Replace `0000XX` with the next sequential number. Then apply:

```bash
cd backend
docker compose run --rm migrate
```

**Never modify a committed migration** — always create a new one.

### Regenerate sqlc Types

After editing SQL query files in `backend/internal/db/queries/`:

```bash
cd backend
sqlc generate
```

### Regenerate Frontend API Client

After modifying `backend/api/openapi.yaml`:

```bash
cd frontend
pnpm codegen
```

This runs Orval to regenerate `src/api/generated/`. Never edit those files
by hand.

### Validate the OpenAPI Spec

```bash
cd frontend
npx @redocly/cli@latest lint ../backend/api/openapi.yaml
```

### Run Backend Tests

```bash
cd backend
go test ./...
```

### Run Frontend Lint/Type Check

```bash
cd frontend
pnpm lint
pnpm typecheck
```

## Workflow for API Changes

When changing API behavior, follow this order:

1. Edit `backend/api/openapi.yaml` (the spec is the source of truth)
2. Validate the spec
3. Implement the backend changes
4. Regenerate the frontend client (`pnpm codegen`)
5. Update frontend code to use new/changed endpoints
6. Commit the spec, backend, and regenerated frontend files together

CI enforces OpenAPI validation and Orval drift-check on PRs.

## Environment Variables

### Backend

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (`development`, `production`) | — |
| `APP_ADDR` | Listen address | `:8080` |
| `DATABASE_URL` | PostgreSQL connection string | — |
| `JWT_SECRET` | JWT signing secret | — |
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |

### Frontend

| Variable | Description |
|----------|-------------|
| `EXPO_PUBLIC_API_URL` | Backend API base URL |

`EXPO_PUBLIC_*` values are bundled into the app — only put public configuration
there. Never put secrets in these variables.

## Project Conventions

- See [`CLAUDE.md`](../CLAUDE.md) for cross-cutting conventions
- See [`backend/CLAUDE.md`](../backend/CLAUDE.md) for backend-specific rules
- See [`frontend/CLAUDE.md`](../frontend/CLAUDE.md) for frontend-specific rules

## Branching and PRs

- One concern per PR
- PR titles use conventional commit prefixes: `feat:`, `fix:`, `chore:`,
  `docs:`, `refactor:`
- All status checks must pass before merge
