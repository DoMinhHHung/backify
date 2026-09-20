from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ProjectNotFoundError
from app.domain.events import ProjectEvent
from app.port.event_publisher import EventPublisher
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class DeleteProjectInput:
    project_id: UUID


class DeleteProject:
    def __init__(
        self,
        project_repository: ProjectRepository,
        event_publisher: EventPublisher,
    ) -> None:
        self._project_repository = project_repository
        self._event_publisher = event_publisher

    async def execute(self, input_data: DeleteProjectInput) -> None:
        project = await self._project_repository.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))

        deleted = await self._project_repository.delete(project.id)
        if not deleted:
            raise ProjectNotFoundError(str(input_data.project_id))

        await self._event_publisher.publish(
            ProjectEvent.deleted(project.id, project.slug, project.schema_name)
        )