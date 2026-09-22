import logging
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

import structlog
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from app.adapter.grpc.server import create_grpc_server
from app.adapter.http.error_handlers import register_error_handlers
from app.adapter.http.middleware import RequestContextMiddleware
from app.adapter.http.routers import auth_router, projects_router
from app.adapter.messaging.noop_publisher import NoopEventPublisher
from app.adapter.messaging.outbox_relay import OutboxRelay
from app.adapter.messaging.rabbitmq_publisher import RabbitMQEventPublisher
from app.adapter.postgres.connection import Database
from app.config import Settings, get_settings
from app.container import Container

logger = structlog.get_logger()


def configure_logging(settings: Settings) -> None:
    level = logging.getLevelNamesMapping().get(settings.log_level.upper(), logging.INFO)
    processors: list[structlog.types.Processor] = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
    ]
    processors.append(
        structlog.processors.JSONRenderer()
        if settings.log_json
        else structlog.dev.ConsoleRenderer()
    )
    structlog.configure(
        processors=processors,
        wrapper_class=structlog.make_filtering_bound_logger(level),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
        cache_logger_on_first_use=True,
    )


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    settings = get_settings()
    configure_logging(settings)
    logger.info("app_starting", env=settings.app_env, port=settings.app_port)

    database = Database(settings)
    publisher = (
        RabbitMQEventPublisher(settings)
        if settings.rabbitmq_enabled
        else NoopEventPublisher()
    )
    relay = OutboxRelay(
        database,
        publisher,
        poll_interval=settings.outbox_poll_interval,
        batch_size=settings.outbox_batch_size,
    )

    container = Container.from_database(settings, database)
    grpc_server = (
        create_grpc_server(container, settings) if settings.grpc_enabled else None
    )

    try:
        await database.connect()
        await publisher.connect()
        await relay.start()
        if grpc_server is not None:
            await grpc_server.start()
            logger.info("grpc_server_started", port=settings.grpc_port)
    except Exception:
        if grpc_server is not None:
            await grpc_server.stop(grace=None)
        await relay.stop()
        await publisher.disconnect()
        await database.disconnect()
        logger.error("startup_failed")
        raise

    app.state.settings = settings
    app.state.db = database
    app.state.container = container
    app.state.relay = relay
    app.state.grpc_server = grpc_server
    logger.info("app_started")

    try:
        yield
    finally:
        if grpc_server is not None:
            await grpc_server.stop(grace=5.0)
        await relay.stop()
        await publisher.disconnect()
        await database.disconnect()
        logger.info("app_stopped")


def create_app() -> FastAPI:
    settings = get_settings()
    application = FastAPI(
        title="Backify Control Plane",
        version="0.1.0",
        lifespan=lifespan,
        docs_url=None if settings.is_production else "/docs",
        redoc_url=None,
    )
    application.add_middleware(RequestContextMiddleware)
    application.add_middleware(
        CORSMiddleware,
        allow_origins=settings.cors_origin_list,
        allow_credentials=True,
        allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        allow_headers=["Authorization", "Content-Type", "X-Request-ID"],
    )
    register_error_handlers(application)
    application.include_router(auth_router)
    application.include_router(projects_router)

    @application.get("/health", tags=["internal"])
    async def health() -> JSONResponse:
        database = getattr(application.state, "db", None)
        db_ok = await database.health_check() if database is not None else False
        return JSONResponse(
            status_code=200 if db_ok else 503,
            content={
                "status": "ok" if db_ok else "degraded",
                "db": "ok" if db_ok else "error",
            },
        )

    return application


app = create_app()
