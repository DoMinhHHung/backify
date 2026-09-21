import json
from typing import Any

from app.adapter.postgres.connection import Database
from app.domain.events import ProjectEvent


class PostgresOutboxRepository:
    def __init__(self, database: Database) -> None:
        self._db = database

    async def add(self, event: ProjectEvent, conn: Any = None) -> None:
        executor = conn if conn is not None else self._db
        await executor.execute(
            """
            INSERT INTO control.outbox (event_type, payload)
            VALUES ($1, $2::jsonb)
            """,
            event.event_type.value,
            json.dumps(event.to_payload()),
        )
