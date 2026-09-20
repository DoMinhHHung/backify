from dataclasses import dataclass
from uuid import UUID

from app.domain.errors import ProjectNotFoundError
from app.domain.field import Field, FieldType
from app.domain.project import Project
from app.port.project_repository import ProjectRepository


@dataclass(frozen=True)
class AddFieldInput:
    project_id: UUID
    entity_name: str
    name: str
    field_type: FieldType
    required: bool = False
    unique: bool = False
    enum_values: list[str] | None = None


@dataclass(frozen=True)
class AddFieldOutput:
    project: Project


class AddField:
    def __init__(self, project_repository: ProjectRepository) -> None:
        self._project_repository = project_repository

    async def execute(self, input_data: AddFieldInput) -> AddFieldOutput:
        project = await self._project_repository.get_by_id(input_data.project_id)
        if project is None:
            raise ProjectNotFoundError(str(input_data.project_id))

        field = Field(
            name=input_data.name,
            field_type=input_data.field_type,
            required=input_data.required,
            unique=input_data.unique,
            enum_values=input_data.enum_values,
        )
        project.add_field_to_entity(input_data.entity_name, field)
        await self._project_repository.save(project)
        return AddFieldOutput(project=project)