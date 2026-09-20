from typing import Protocol

from app.domain.events import ProjectEvent


class EventPublisher(Protocol):
    async def publish(self, event: ProjectEvent) -> None: ...

    async def connect(self) -> None: ...

    async def disconnect(self) -> None: ...