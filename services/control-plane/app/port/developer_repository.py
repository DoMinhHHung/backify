from typing import Protocol
from uuid import UUID

from app.domain.developer import Developer


class DeveloperRepository(Protocol):
    async def save(self, developer: Developer) -> None: ...

    async def get_by_id(self, developer_id: UUID) -> Developer | None: ...

    async def get_by_email(self, email: str) -> Developer | None: ...

    async def exists_by_email(self, email: str) -> bool: ...