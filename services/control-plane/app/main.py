from contextlib import asynccontextmanager
from typing import AsyncIterator

import structlog
from fastapi import FastAPI
from fastapi.responses import JSONResponse

from app.adapter.http.error_handlers import register_error_handlers
from app.adapter.http.routers import projects_router
from app.adapter.messaging.noop_publisher import NoopEventPublisher
from app.adapter.messaging.rabbitmq_publisher import RabbitMQEventPublisher
from app.adapter.postgres.connection import Database
from app.config import get_settings
from app.container import Container

logger = structlog.get_logger()


def configure_logging() -> None:
    settings = get_settings()
    processors: list[structlog.types.Processor] = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
    ]
    if settings.log_json:
        processors.append(structlog.processors.JSONRenderer())
    else:
        processors.append(structlog.dev.ConsoleRenderer())
    structlog.configure(
        processors=processors,
        wrapper_class=structlog.make_filtering_bound_logger(
            getattr(structlog.stdlib, settings.log_level.upper(), 20)
        ),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
        cache_logger_on_first_use=True,
    )


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    configure_logging()
    settings = get_settings()
    logger.info(
        "app_starting",
        env=settings.app_env,
        host=settings.app_host,
        port=settings.app_port,
    )

    database = Database(settings)
    if settings.rabbitmq_enabled:
        event_publisher = RabbitMQEventPublisher(settings)
    else:
        event_publisher = NoopEventPublisher()

    try:
        await database.connect()
        await event_publisher.connect()
        container = Container(settings, database, event_publisher)
        app.state.container = container
        app.state.db = database
        app.state.event_publisher = event_publisher
        logger.info("db_connected")
    except Exception as exc:
        logger.error("startup_failed", error=str(exc))
        app.state.db = None
        app.state.container = None
        app.state.event_publisher = None
    yield

    publisher = getattr(app.state, "event_publisher", None)
    if publisher is not None:
        await publisher.disconnect()
    db = getattr(app.state, "db", None)
    if db is not None:
        await db.disconnect()
    logger.info("app_stopped")


app = FastAPI(title="Backify Control Plane", lifespan=lifespan)
register_error_handlers(app)
app.include_router(projects_router)


@app.get("/health")
async def health() -> JSONResponse:
    db_status = "error"
    db = getattr(app.state, "db", None)
    if db is not None:
        ok = await db.health_check()
        db_status = "ok" if ok else "error"
    status = "ok" if db_status == "ok" else "degraded"
    return JSONResponse(
        status_code=200 if db_status == "ok" else 503,
        content={"status": status, "db": db_status},
    )