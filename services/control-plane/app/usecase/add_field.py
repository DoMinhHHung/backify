from dataclasses import dataclass
from uuid import UUID

from app.domain.field import Field, FieldType, RelationCardinality
from app.domain.project import Project
from app.usecase._project_mutation import ProjectMutationUseCase


@dataclass(frozen=True)
class AddFieldInput:
    project_id: UUID
    owner_id: UUID
    entity_name: str
    name: str
    field_type: FieldType
    required: bool = False
    unique: bool = False
    enum_values: list[str] | None = None
    relation_to: str | None = None
    relation_cardinality: str | None = None


@dataclass(frozen=True)
class AddFieldOutput:
    project: Project


class AddField(ProjectMutationUseCase):
    async def execute(self, input_data: AddFieldInput) -> AddFieldOutput:
        project = await self._load_owned(input_data.project_id, input_data.owner_id)

        cardinality = (
            RelationCardinality.parse(input_data.relation_cardinality)
            if input_data.relation_cardinality is not None
            else None
        )
        field = Field(
            name=input_data.name,
            field_type=input_data.field_type,
            required=input_data.required,
            unique=input_data.unique,
            enum_values=input_data.enum_values,
            relation_to=input_data.relation_to,
            relation_cardinality=cardinality,
        )
        project.add_field_to_entity(input_data.entity_name, field)
        await self._commit(project)
        return AddFieldOutput(project=project)
