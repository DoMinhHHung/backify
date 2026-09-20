from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ProjectNotFoundError
from app.domain.module import FunctionName, ModuleName
from app.domain.project import Project
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class SetFunctionFieldsInput:
    project_id: UUID
    module: ModuleName
    function: FunctionName
    field_names: list[str]


@dataclass(frozen=True)
class SetFunctionFieldsOutput:
    project: Project


class SetFunctionFields:
    def __init__(self, project_repository: ProjectRepository) -> None:
        self._project_repository = project_repository

    async def execute(
        self, input_data: SetFunctionFieldsInput
    ) -> SetFunctionFieldsOutput:
        project = await self._project_repository.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))

        project.set_function_fields(
            input_data.module,
            input_data.function,
            input_data.field_names,
        )
        await self._project_repository.save(project)
        return SetFunctionFieldsOutput(project=project)