import structlog

from app.domain.events import ProjectEvent

logger = structlog.get_logger()


class NoopEventPublisher:
    async def connect(self) -> None:
        logger.info("event_publisher_noop_connected")

    async def disconnect(self) -> None:
        logger.info("event_publisher_noop_disconnected")

    async def publish(self, event: ProjectEvent) -> None:
        logger.info(
            "event_published_noop",
            event=event.event_type.value,
            project_id=str(event.project_id),
            slug=event.slug,
        )