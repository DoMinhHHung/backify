import json

import aio_pika
import structlog
from aio_pika import ExchangeType

from app.config import Settings
from app.domain.events import ProjectEvent

logger = structlog.get_logger()

EXCHANGE_NAME = "backify.events"


class RabbitMQEventPublisher:
    def __init__(self, settings: Settings) -> None:
        self._url = settings.rabbitmq_url
        self._connection: aio_pika.RobustConnection | None = None
        self._channel: aio_pika.Channel | None = None
        self._exchange: aio_pika.Exchange | None = None

    async def connect(self) -> None:
        if self._connection is not None:
            return
        self._connection = await aio_pika.connect_robust(self._url)
        self._channel = await self._connection.channel()
        self._exchange = await self._channel.declare_exchange(
            EXCHANGE_NAME,
            ExchangeType.TOPIC,
            durable=True,
        )
        logger.info("event_publisher_rabbitmq_connected", exchange=EXCHANGE_NAME)

    async def disconnect(self) -> None:
        if self._connection is not None:
            await self._connection.close()
            self._connection = None
            self._channel = None
            self._exchange = None
            logger.info("event_publisher_rabbitmq_disconnected")

    async def publish(self, event: ProjectEvent) -> None:
        if self._exchange is None:
            raise RuntimeError("event publisher is not connected")
        body = json.dumps(event.to_payload()).encode("utf-8")
        message = aio_pika.Message(
            body=body,
            content_type="application/json",
            delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
        )
        await self._exchange.publish(message, routing_key=event.event_type.value)
        logger.info(
            "event_published",
            event=event.event_type.value,
            project_id=str(event.project_id),
            slug=event.slug,
        )