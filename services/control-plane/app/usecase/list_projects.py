from dataclasses import dataclass
from uuid import UUID

from app.domain.project import Project
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class ListProjectsInput:
    owner_id: UUID


@dataclass(frozen=True)
class ListProjectsOutput:
    projects: list[Project]


class ListProjects:
    def __init__(self, project_repository: ProjectRepository) -> None:
        self._project_repository = project_repository

    async def execute(self, input_data: ListProjectsInput) -> ListProjectsOutput:
        projects = await self._project_repository.list_by_owner(input_data.owner_id)
        return ListProjectsOutput(projects=projects)