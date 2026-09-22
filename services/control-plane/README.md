# Control Plane

Dashboard API for Backify — manage projects, entities, field pools, and modules.

## Stack

- Python / FastAPI
- PostgreSQL (schema `control`)
- RabbitMQ (optional, `RABBITMQ_ENABLED=false` uses noop publisher)
- JWT for developer auth
- gRPC internal API (`GetProject`, `GetProjectConfig`) with `x-internal-key`
- Optional gRPC TLS / mutual TLS

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

### Local secrets

`.env.example` sets `ALLOW_INSECURE_DEFAULTS=true` so binding `0.0.0.0` works with
placeholder secrets. For any non-loopback bind without that flag, `JWT_SECRET` and
`INTERNAL_API_KEY` must each be at least 32 characters and not a known insecure
default. Production / staging always enforce secrets; production also requires
gRPC TLS when gRPC is enabled.

### gRPC TLS

```bash
GRPC_TLS_CERT_FILE=/path/to/server.crt
GRPC_TLS_KEY_FILE=/path/to/server.key
# mutual TLS (optional):
GRPC_TLS_CLIENT_CA_FILE=/path/to/ca.crt
```

Without cert/key files the server still binds insecure (dev only). Production
refuses to start gRPC without TLS.

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
