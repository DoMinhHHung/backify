# Control Plane

Dashboard API for Backify — manage projects, entities, field pools, and modules.

## Stack

- Python 3.12+ / FastAPI
- PostgreSQL (schema `control`)
- RabbitMQ (optional, `RABBITMQ_ENABLED=false` uses noop publisher)
- JWT for developer auth

## Setup

```bash
cd services/control-plane
cp .env.example .env
uv sync --extra dev
alembic upgrade head
uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000