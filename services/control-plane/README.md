# Control Plane

Dashboard API for Backify — manage projects, entities, field pools, and modules.

## Stack

- Python / FastAPI
- PostgreSQL (schema `control`)
- RabbitMQ (optional, `RABBITMQ_ENABLED=false` uses noop publisher)
- JWT for developer auth
- gRPC internal API (`GetProject`, `GetProjectConfig`) with `x-internal-key`

## Setup

```bash
cd services/control-plane
cp .env.example .env
uv sync --extra dev
uv run alembic upgrade head
uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

From repo root:

```bash
make install
make migrate-up
make run
```

Regenerate gRPC stubs after changing `proto/control_plane.proto`:

```bash
make grpc-gen
```

## Main flows

1. `POST /api/v1/auth/register` → `POST /api/v1/auth/login`
2. `POST /api/v1/projects` (Bearer token)
3. Entities / fields / modules under `/api/v1/projects/{id}/...`
4. Internal HTTP: `GET /api/v1/projects/{id}/config` with header `X-Internal-Key`
5. Internal gRPC: `ControlPlaneService` on `GRPC_PORT` (default 9091), metadata `x-internal-key`

## Events (outbox → RabbitMQ topic `backify.events`)

- `project.created`
- `project.config.updated`
- `project.deleted`

Publish uses stable `message_id` = outbox row id for consumer idempotency.

## Tests

```bash
uv run pytest
# or from root:
make test
```
