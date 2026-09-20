from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ProjectNotFoundError
from app.domain.project import Project
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class RemoveFieldInput:
    project_id: UUID
    entity_name: str
    field_name: str
    force: bool = False


@dataclass(frozen=True)
class RemoveFieldOutput:
    project: Project


class RemoveField:
    def __init__(self, project_repository: ProjectRepository) -> None:
        self._project_repository = project_repository

    async def execute(self, input_data: RemoveFieldInput) -> RemoveFieldOutput:
        project = await self._project_repository.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))

        project.remove_field_from_entity(
            input_data.entity_name,
            input_data.field_name,
            force=input_data.force,
        )
        await self._project_repository.save(project)
        return RemoveFieldOutput(project=project)