from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ProjectNotFoundError
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class GetProjectConfigInput:
    project_id: UUID


@dataclass(frozen=True)
class GetProjectConfigOutput:
    project_id: UUID
    slug: str
    schema_name: str
    version: int
    config: dict[str, object]


class GetProjectConfig:
    def __init__(self, project_repository: ProjectRepository) -> None:
        self._projects = project_repository

    async def execute(self, input_data: GetProjectConfigInput) -> GetProjectConfigOutput:
        project = await self._projects.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))
        return GetProjectConfigOutput(
            project_id=project.id,
            slug=project.slug,
            schema_name=project.schema_name,
            version=project.version,
            config=project.to_config_dict(),
        )
