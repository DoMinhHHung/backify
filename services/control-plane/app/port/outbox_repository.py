from typing import Any, Protocol

from app.domain.events import ProjectEvent


class OutboxRepository(Protocol):
    async def add(self, event: ProjectEvent, conn: Any = None) -> None: ...
