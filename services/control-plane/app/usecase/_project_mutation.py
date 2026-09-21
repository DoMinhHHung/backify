from uuid import UUID

from app.domain.errors import ForbiddenError, ProjectNotFoundError
from app.domain.events import ProjectEvent
from app.domain.project import Project
from app.port.outbox_repository import OutboxRepository
from app.port.project_repository import ProjectRepository
from app.port.transaction import TransactionManager


class ProjectMutationUseCase:
    def __init__(
        self,
        project_repository: ProjectRepository,
        outbox: OutboxRepository,
        tx: TransactionManager,
    ) -> None:
        self._projects = project_repository
        self._outbox = outbox
        self._tx = tx

    async def _load_owned(self, project_id: UUID, owner_id: UUID) -> Project:
        project = await self._projects.get_by_id(project_id)
        if project is None:
            raise ProjectNotFoundError(str(project_id))
        if project.owner_id != owner_id:
            raise ForbiddenError("you do not own this project")
        return project

    async def _commit(self, project: Project) -> None:
        async with self._tx.transaction() as conn:
            await self._projects.save(project, conn)
            await self._outbox.add(
                ProjectEvent.config_updated(
                    project.id, project.slug, project.schema_name
                ),
                conn,
            )
