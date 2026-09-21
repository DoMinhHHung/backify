from contextlib import asynccontextmanager
from typing import Any
from uuid import UUID

from app.domain.developer import Developer
from app.domain.errors import (
    ConcurrentModificationError,
    DeveloperAlreadyExistsError,
    ProjectSlugExistsError,
)
from app.domain.events import ProjectEvent
from app.domain.project import Project


class FakeProjectRepository:
    def __init__(self) -> None:
        self.rows: dict[UUID, tuple[dict[str, Any], int]] = {}

    def _store(self, project: Project, version: int) -> None:
        self.rows[project.id] = (
            {
                "name": project.name,
                "slug": project.slug,
                "schema_name": project.schema_name,
                "owner_id": project.owner_id,
                "config": project.to_config_dict(),
            },
            version,
        )

    async def create(self, project: Project, conn: Any = None) -> None:
        if any(r["slug"] == project.slug for r, _ in self.rows.values()):
            raise ProjectSlugExistsError(project.slug)
        project.version = 1
        self._store(project, 1)

    async def save(self, project: Project, conn: Any = None) -> None:
        existing = self.rows.get(project.id)
        if existing is None or existing[1] != project.version:
            raise ConcurrentModificationError(str(project.id))
        project.version += 1
        self._store(project, project.version)

    async def get_by_id(self, project_id: UUID) -> Project | None:
        entry = self.rows.get(project_id)
        if entry is None:
            return None
        row, version = entry
        return Project.from_persistence(
            project_id=project_id,
            name=row["name"],
            slug=row["slug"],
            config=row["config"],
            owner_id=row["owner_id"],
            version=version,
        )

    async def get_by_slug(self, slug: str) -> Project | None:
        for pid, (row, _) in self.rows.items():
            if row["slug"] == slug:
                return await self.get_by_id(pid)
        return None

    async def exists_by_slug(self, slug: str) -> bool:
        return any(row["slug"] == slug for row, _ in self.rows.values())

    async def delete(self, project_id: UUID, conn: Any = None) -> bool:
        return self.rows.pop(project_id, None) is not None

    async def list_by_owner(self, owner_id: UUID) -> list[Project]:
        result = []
        for pid, (row, _) in self.rows.items():
            if row["owner_id"] == owner_id:
                project = await self.get_by_id(pid)
                if project is not None:
                    result.append(project)
        return result


class FakeDeveloperRepository:
    def __init__(self) -> None:
        self.items: dict[UUID, Developer] = {}

    async def create(self, developer: Developer, conn: Any = None) -> None:
        if await self.exists_by_email(developer.email):
            raise DeveloperAlreadyExistsError(developer.email)
        self.items[developer.id] = developer

    async def get_by_id(self, developer_id: UUID) -> Developer | None:
        return self.items.get(developer_id)

    async def get_by_email(self, email: str) -> Developer | None:
        normalized = email.strip().lower()
        return next((d for d in self.items.values() if d.email == normalized), None)

    async def exists_by_email(self, email: str) -> bool:
        return await self.get_by_email(email) is not None


class FakeOutboxRepository:
    def __init__(self) -> None:
        self.events: list[ProjectEvent] = []

    async def add(self, event: ProjectEvent, conn: Any = None) -> None:
        self.events.append(event)

    def types(self) -> list[str]:
        return [e.event_type.value for e in self.events]


class FakeTransactionManager:
    @asynccontextmanager
    async def transaction(self):
        yield None
