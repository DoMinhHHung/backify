from dataclasses import dataclass
from uuid import UUID

from app.domain.project import Project
from app.usecase._project_mutation import ProjectMutationUseCase


@dataclass(frozen=True)
class RemoveFieldInput:
    project_id: UUID
    owner_id: UUID
    entity_name: str
    field_name: str
    force: bool = False


@dataclass(frozen=True)
class RemoveFieldOutput:
    project: Project


class RemoveField(ProjectMutationUseCase):
    async def execute(self, input_data: RemoveFieldInput) -> RemoveFieldOutput:
        project = await self._load_owned(input_data.project_id, input_data.owner_id)
        project.remove_field_from_entity(
            input_data.entity_name,
            input_data.field_name,
            force=input_data.force,
        )
        await self._commit(project)
        return RemoveFieldOutput(project=project)
