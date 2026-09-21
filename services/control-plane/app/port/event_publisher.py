from typing import Protocol


class EventPublisher(Protocol):
    async def connect(self) -> None: ...

    async def disconnect(self) -> None: ...

    async def publish(
        self,
        event_type: str,
        payload: dict[str, object],
        *,
        message_id: str | None = None,
    ) -> None: ...
