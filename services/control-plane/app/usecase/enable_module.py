from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ForbiddenError, ProjectNotFoundError
from app.domain.events import ProjectEvent
from app.domain.module import ModuleName
from app.domain.project import Project
from app.port.event_publisher import EventPublisher
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class EnableModuleInput:
    project_id: UUID
    owner_id: UUID
    module: ModuleName


@dataclass(frozen=True)
class EnableModuleOutput:
    project: Project


class EnableModule:
    def __init__(
        self,
        project_repository: ProjectRepository,
        event_publisher: EventPublisher,
    ) -> None:
        self._project_repository = project_repository
        self._event_publisher = event_publisher

    async def execute(self, input_data: EnableModuleInput) -> EnableModuleOutput:
        project = await self._project_repository.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))
        if project.owner_id != input_data.owner_id:
            raise ForbiddenError("you do not own this project")

        project.enable_module_by_name(input_data.module)
        await self._project_repository.save(project)
        await self._event_publisher.publish(
            ProjectEvent.config_updated(project.id, project.slug, project.schema_name)
        )
        return EnableModuleOutput(project=project)