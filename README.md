# Sumni Finance Backend

Backend service for Sumni Finance — a personal/treasury finance platform. Built in Go as a modular monolith following Clean Architecture / Hexagonal (Ports & Adapters) principles, with Keycloak-based authentication and authorization.

## Tech Stack

- **Language**: Go 1.26
- **HTTP**: [Echo](https://echo.labstack.com/)
- **Database**: PostgreSQL (via [pgx](https://github.com/jackc/pgx))
- **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Auth**: [Keycloak](https://www.keycloak.org/) (OAuth2/OIDC + policy enforcement)
- **API contracts**: OpenAPI, code generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)
- **DB queries**: [sqlc](https://sqlc.dev/)
- **Task runner**: [Task](https://taskfile.dev/)

## Architecture

The codebase is organized as a modular monolith. Each business module under `internal/` is self-contained and typically follows the same internal layering:

- `domain/` — entities, value objects, and repository interfaces (pure business logic)
- `app/` — application layer: commands and queries orchestrating the domain
- `adapters/` — infrastructure implementations (Postgres via sqlc, external services)
- `api/` — inbound interfaces (HTTP handlers, module wiring)

```
internal/
├── identity/    # authentication, sessions, Keycloak integration, policy enforcement
├── treasury/    # fund sources, journal entries, bank lookups
├── envelope/    # budgeting/envelope domain
└── common/      # shared config, HTTP server, logging, module contracts, test utils
```

Modules are wired together in [internal/svc.go](internal/svc.go) and started from [cmd/main.go](cmd/main.go). Cross-module dependencies are exposed explicitly through `internal/common/module/contracts`.

## Prerequisites

- [Docker](https://www.docker.com/) and Docker Compose
- [Go 1.26+](https://go.dev/dl/) (for local tooling outside Docker)
- [Task](https://taskfile.dev/installation/)

## Getting Started

1. Copy the environment file and adjust values as needed:

   ```bash
   cp .env.example .env
   ```

2. Start the app, Postgres, and Keycloak:

   ```bash
   task up
   ```

   This builds and runs the backend in Docker with hot reload (via `reflex`). The API is available at `http://localhost:4000`, and Keycloak at `http://localhost:8080`.

   To restart with a clean state:

   ```bash
   task up-clean
   ```

3. Stop the stack:

   ```bash
   task down
   # or, to also drop volumes (e.g. Postgres data):
   task down-volumes
   ```

