from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ProjectNotFoundError
from app.domain.events import ProjectEvent
from app.usecase._project_mutation import ProjectMutationUseCase


@dataclass(frozen=True)
class DeleteProjectInput:
    project_id: UUID
    owner_id: UUID


class DeleteProject(ProjectMutationUseCase):
    async def execute(self, input_data: DeleteProjectInput) -> None:
        project = await self._load_owned(input_data.project_id, input_data.owner_id)
        async with self._tx.transaction() as conn:
            deleted = await self._projects.delete(project.id, conn)
            if not deleted:
                raise ProjectNotFoundError(str(input_data.project_id))
            await self._outbox.add(
                ProjectEvent.deleted(project.id, project.slug, project.schema_name),
                conn,
            )
