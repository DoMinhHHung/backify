from dataclasses import dataclass

from app.domain.errors import ProjectSlugExistsError
from app.domain.project import Project
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class CreateProjectInput:
    name: str
    slug: str


@dataclass(frozen=True)
class CreateProjectOutput:
    project: Project


class CreateProject:
    def __init__(self, project_repository: ProjectRepository) -> None:
        self._project_repository = project_repository

    async def execute(self, input_data: CreateProjectInput) -> CreateProjectOutput:
        if await self._project_repository.exists_by_slug(input_data.slug):
            raise ProjectSlugExistsError(input_data.slug)

        project = Project.create(name=input_data.name, slug=input_data.slug)
        await self._project_repository.save(project)
        return CreateProjectOutput(project=project)