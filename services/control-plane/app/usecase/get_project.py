from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ForbiddenError, ProjectNotFoundError
from app.domain.project import Project
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class GetProjectInput:
    project_id: UUID
    owner_id: UUID


@dataclass(frozen=True)
class GetProjectOutput:
    project: Project


class GetProject:
    def __init__(self, project_repository: ProjectRepository) -> None:
        self._project_repository = project_repository

    async def execute(self, input_data: GetProjectInput) -> GetProjectOutput:
        project = await self._project_repository.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))
        if project.owner_id != input_data.owner_id:
            raise ForbiddenError("you do not own this project")
        return GetProjectOutput(project=project)