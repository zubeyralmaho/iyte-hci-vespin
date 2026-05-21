# Vespin Documentation

Welcome to the Vespin project documentation. This directory contains structured
guides covering all aspects of the system.

## Contents

| Document | Description |
|----------|-------------|
| [Architecture](./architecture.md) | System architecture, component overview, and domain boundaries |
| [API Reference](./api.md) | REST API conventions, endpoints, authentication, and error handling |
| [Data Model](./data-model.md) | Database schema, migrations, and data conventions |
| [Authentication](./authentication.md) | Auth flows, JWT tokens, guest/registered roles |
| [Development](./development.md) | Local setup, tooling, workflow, and common tasks |
| [Deployment](./deployment.md) | Production infrastructure, CI/CD, and deploy process |

## Quick Links

- **OpenAPI Spec:** [`backend/api/openapi.yaml`](../backend/api/openapi.yaml)
- **Database Schema:** [`backend/internal/db/SCHEMA.md`](../backend/internal/db/SCHEMA.md)
- **Design System:** [`specs/design-system.md`](../specs/design-system.md)
- **Agent Conventions:** [`CLAUDE.md`](../CLAUDE.md)

## About This Project

Vespin is a companion mobile app for the fictional **Vespin Retro** series
Bluetooth smart speakers, developed as a university HCI course project at
IYTE (Izmir Institute of Technology). The project brings together students from
Industrial Design and Computer Engineering.

Hardware interaction is fully simulated — the focus is on user experience,
interface design, and a working backend architecture.
