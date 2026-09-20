from dataclasses import dataclass
from uuid import UUID

from app.domain.entity import Entity
from app.domain.errors import ProjectNotFoundError
from app.domain.events import ProjectEvent
from app.domain.project import Project
from app.port.event_publisher import EventPublisher
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class AddEntityInput:
    project_id: UUID
    name: str


@dataclass(frozen=True)
class AddEntityOutput:
    project: Project


class AddEntity:
    def __init__(
        self,
        project_repository: ProjectRepository,
        event_publisher: EventPublisher,
    ) -> None:
        self._project_repository = project_repository
        self._event_publisher = event_publisher

    async def execute(self, input_data: AddEntityInput) -> AddEntityOutput:
        project = await self._project_repository.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))

        entity = Entity(input_data.name)
        project.add_entity(entity)
        await self._project_repository.save(project)
        await self._event_publisher.publish(
            ProjectEvent.config_updated(project.id, project.slug, project.schema_name)
        )
        return AddEntityOutput(project=project)