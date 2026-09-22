from dataclasses import dataclass
from uuid import UUID

from app.domain.entity import Entity
from app.domain.project import Project
from app.usecase._project_mutation import ProjectMutationUseCase


@dataclass(frozen=True)
class AddEntityInput:
    project_id: UUID
    owner_id: UUID
    name: str


@dataclass(frozen=True)
class AddEntityOutput:
    project: Project


class AddEntity(ProjectMutationUseCase):
    async def execute(self, input_data: AddEntityInput) -> AddEntityOutput:
        project = await self._load_owned(input_data.project_id, input_data.owner_id)
        project.add_entity(Entity(input_data.name))
        await self._commit(project)
        return AddEntityOutput(project=project)
