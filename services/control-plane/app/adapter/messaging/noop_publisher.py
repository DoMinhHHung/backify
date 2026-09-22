import structlog

logger = structlog.get_logger()


class NoopEventPublisher:
    async def connect(self) -> None:
        logger.warning("event_publisher_noop_enabled")

    async def disconnect(self) -> None:
        return None

    async def publish(
        self,
        event_type: str,
        payload: dict[str, object],
        *,
        message_id: str | None = None,
    ) -> None:
        logger.info(
            "event_dropped_noop",
            event_type=event_type,
            message_id=message_id,
            payload=payload,
        )
