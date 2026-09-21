from dataclasses import dataclass
from uuid import UUID

from app.domain.module import ModuleName
from app.domain.project import Project
from app.usecase._project_mutation import ProjectMutationUseCase


@dataclass(frozen=True)
class EnableModuleInput:
    project_id: UUID
    owner_id: UUID
    module: ModuleName


@dataclass(frozen=True)
class EnableModuleOutput:
    project: Project


class EnableModule(ProjectMutationUseCase):
    async def execute(self, input_data: EnableModuleInput) -> EnableModuleOutput:
        project = await self._load_owned(input_data.project_id, input_data.owner_id)
        project.enable_module_by_name(input_data.module)
        await self._commit(project)
        return EnableModuleOutput(project=project)
