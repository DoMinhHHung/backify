from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ProjectNotFoundError
from app.domain.project import Project
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class GetProjectConfigInput:
    project_id: UUID


@dataclass(frozen=True)
class GetProjectConfigOutput:
    project: Project


class GetProjectConfig:
    def __init__(self, project_repository: ProjectRepository) -> None:
        self._project_repository = project_repository

    async def execute(
        self, input_data: GetProjectConfigInput
    ) -> GetProjectConfigOutput:
        project = await self._project_repository.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))
        return GetProjectConfigOutput(project=project)