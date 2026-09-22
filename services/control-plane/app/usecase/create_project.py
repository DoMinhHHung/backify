from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ProjectSlugExistsError
from app.domain.events import ProjectEvent
from app.domain.project import Project
from app.port.outbox_repository import OutboxRepository
from app.port.project_repository import ProjectRepository
from app.port.transaction import TransactionManager


@dataclass(frozen=True)
class CreateProjectInput:
    name: str
    slug: str
    owner_id: UUID


@dataclass(frozen=True)
class CreateProjectOutput:
    project: Project


class CreateProject:
    def __init__(
        self,
        project_repository: ProjectRepository,
        outbox: OutboxRepository,
        tx: TransactionManager,
    ) -> None:
        self._projects = project_repository
        self._outbox = outbox
        self._tx = tx

    async def execute(self, input_data: CreateProjectInput) -> CreateProjectOutput:
        if await self._projects.exists_by_slug(input_data.slug):
            raise ProjectSlugExistsError(input_data.slug)

        project = Project.create(
            name=input_data.name,
            slug=input_data.slug,
            owner_id=input_data.owner_id,
        )
        async with self._tx.transaction() as conn:
            await self._projects.create(project, conn)
            await self._outbox.add(
                ProjectEvent.created(project.id, project.slug, project.schema_name),
                conn,
            )
        return CreateProjectOutput(project=project)
