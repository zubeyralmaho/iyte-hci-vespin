# Deployment

## Infrastructure Overview

```text
┌─────────────────────────────────────────────────────┐
│                  Hetzner VPS                         │
│                                                     │
│  ┌─────────┐      ┌─────────┐      ┌───────────┐  │
│  │  Caddy  │─────►│   API   │─────►│ PostgreSQL│  │
│  │ :80/:443│      │  :8080  │      │   :5432   │  │
│  └─────────┘      └─────────┘      └───────────┘  │
│       ▲                                             │
│       │ TLS (Let's Encrypt)                         │
└───────┼─────────────────────────────────────────────┘
        │
    Internet
```

- **Caddy** — reverse proxy with automatic HTTPS (Let's Encrypt)
- **API** — Go binary in Docker container (`vespin-api`)
- **PostgreSQL** — database, internal-only access
- **Migrate** — one-shot container that applies migrations before API starts

Only Caddy is exposed to the internet. The API and Postgres are internal to the
Docker network.

## Container Images

| Image | Registry | Purpose |
|-------|----------|---------|
| `vespin-api` | GitHub Container Registry | API server |
| `vespin-migrate` | GitHub Container Registry | Migration runner |

Images are built and pushed by the CI workflow on merge to `main` (when deploy
is enabled).

## Production Docker Compose

The production stack is defined in [`deploy/docker-compose.prod.yml`](../deploy/docker-compose.prod.yml).

### Required Environment Variables

Create a `.env` file on the VPS (see `deploy/.env.example`):

| Variable | Description |
|----------|-------------|
| `DOMAIN` | Public domain for Caddy TLS |
| `ACME_EMAIL` | Email for Let's Encrypt certificate |
| `POSTGRES_PASSWORD` | Database password |
| `JWT_SECRET` | JWT signing secret |
| `GHCR_OWNER` | GitHub Container Registry owner |
| `IMAGE_TAG` | Image tag (default: `latest`) |

### Starting the Stack

```bash
cd /home/deploy/vespin
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

## CI/CD Pipeline

### GitHub Actions Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `pr-checks.yml` | Pull requests | Lint, build, test with path filtering |
| `openapi.yml` | Pull requests | OpenAPI validation + Orval drift check |
| `backend.yml` | Push to `main` | Build, push images, deploy to VPS |

### Deploy Workflow (`backend.yml`)

The deploy workflow is **guarded** — it only runs when:

- `DEPLOY_ENABLED` repository variable is set, OR
- Manually triggered via `workflow_dispatch`

When triggered, the workflow:

1. Builds `vespin-api` and `vespin-migrate` Docker images
2. Pushes images to GitHub Container Registry
3. Syncs the `deploy/` directory to the VPS
4. Restarts the Docker Compose stack on the VPS
5. Verifies the API is healthy via `/healthz`

### Required Secrets and Variables

| Name | Type | Description |
|------|------|-------------|
| `DEPLOY_ENABLED` | Variable | Enable automatic deploys |
| `VPS_HOST` | Secret | VPS hostname/IP |
| `VPS_SSH_KEY` | Secret | SSH private key for deployment |
| `GHCR_TOKEN` | Secret | GitHub Container Registry token |

## Backups

A backup script is available at [`deploy/backup.sh`](../deploy/backup.sh).
It performs `pg_dump` to the mounted `/backups` volume on the Postgres
container.

## Domain and DNS

The project uses DuckDNS for the domain (appropriate for the HCI course scope).
Caddy automatically handles certificate provisioning and renewal via the ACME
protocol.

## Deployment Is Not the Focus

Deployment infrastructure is prepared and functional but is intentionally not
the focus of this HCI course submission. The primary deliverable is the mobile
app experience and its supporting backend.
