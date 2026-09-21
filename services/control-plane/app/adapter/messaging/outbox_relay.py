import asyncio
import contextlib
import json

import structlog

from app.adapter.postgres.connection import Database
from app.port.event_publisher import EventPublisher

logger = structlog.get_logger()


class OutboxRelay:
    def __init__(
        self,
        database: Database,
        publisher: EventPublisher,
        *,
        poll_interval: float = 1.0,
        batch_size: int = 50,
    ) -> None:
        self._db = database
        self._publisher = publisher
        self._poll_interval = poll_interval
        self._batch_size = batch_size
        self._task: asyncio.Task[None] | None = None
        self._stopping = asyncio.Event()

    async def start(self) -> None:
        if self._task is not None:
            return
        self._stopping.clear()
        self._task = asyncio.create_task(self._run())
        logger.info("outbox_relay_started", interval=self._poll_interval)

    async def stop(self) -> None:
        if self._task is None:
            return
        self._stopping.set()
        self._task.cancel()
        with contextlib.suppress(asyncio.CancelledError):
            await self._task
        self._task = None
        logger.info("outbox_relay_stopped")

    async def _run(self) -> None:
        while not self._stopping.is_set():
            try:
                sent = await self.drain_once()
            except Exception as exc:
                sent = 0
                logger.error("outbox_relay_error", error=str(exc))
            if sent == 0:
                await asyncio.sleep(self._poll_interval)

    async def drain_once(self) -> int:
        sent = 0
        async with self._db.transaction() as conn:
            rows = await conn.fetch(
                """
                SELECT id, event_type, payload
                FROM control.outbox
                WHERE published_at IS NULL
                ORDER BY id
                LIMIT $1
                FOR UPDATE SKIP LOCKED
                """,
                self._batch_size,
            )
            for row in rows:
                payload = row["payload"]
                if isinstance(payload, str):
                    payload = json.loads(payload)
                await self._publisher.publish(row["event_type"], payload)
                await conn.execute(
                    "UPDATE control.outbox SET published_at = NOW() WHERE id = $1",
                    row["id"],
                )
                sent += 1
        if sent:
            logger.info("outbox_relay_published", count=sent)
        return sent
