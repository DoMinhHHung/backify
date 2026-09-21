from dataclasses import dataclass
from uuid import UUID

from app.domain.module import FunctionName, ModuleName
from app.domain.project import Project
from app.usecase._project_mutation import ProjectMutationUseCase


@dataclass(frozen=True)
class SetFunctionFieldsInput:
    project_id: UUID
    owner_id: UUID
    module: ModuleName
    function: FunctionName
    field_names: list[str]
    entity_name: str | None = None


@dataclass(frozen=True)
class SetFunctionFieldsOutput:
    project: Project


class SetFunctionFields(ProjectMutationUseCase):
    async def execute(
        self, input_data: SetFunctionFieldsInput
    ) -> SetFunctionFieldsOutput:
        project = await self._load_owned(input_data.project_id, input_data.owner_id)
        project.set_function_fields(
            input_data.module,
            input_data.function,
            input_data.field_names,
            entity_name=input_data.entity_name,
        )
        await self._commit(project)
        return SetFunctionFieldsOutput(project=project)
