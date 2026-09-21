from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ProjectNotFoundError
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class GetProjectInternalInput:
    project_id: UUID


@dataclass(frozen=True)
class GetProjectInternalOutput:
    id: UUID
    name: str
    slug: str
    schema_name: str
    owner_id: UUID | None
    version: int


class GetProjectInternal:
    """Looks up a project for trusted internal callers (e.g. gRPC).

    Unlike GetProject, this performs no ownership check: it is only wired
    to transports that already authenticate the caller as a backend
    service (internal API key / gRPC metadata), never to end-user-facing
    HTTP routes.
    """

    def __init__(self, project_repository: ProjectRepository) -> None:
        self._projects = project_repository

    async def execute(
        self, input_data: GetProjectInternalInput
    ) -> GetProjectInternalOutput:
        project = await self._projects.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))
        return GetProjectInternalOutput(
            id=project.id,
            name=project.name,
            slug=project.slug,
            schema_name=project.schema_name,
            owner_id=project.owner_id,
            version=project.version,
        )
