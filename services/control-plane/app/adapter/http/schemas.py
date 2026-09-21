from typing import Any
from uuid import UUID

from pydantic import BaseModel, EmailStr, Field

from app.domain.field import FieldType
from app.domain.project import Project
from app.usecase.get_project_config import GetProjectConfigOutput


class CreateProjectRequest(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    slug: str = Field(min_length=3, max_length=40)


class AddEntityRequest(BaseModel):
    name: str = Field(min_length=1, max_length=48)


class AddFieldRequest(BaseModel):
    name: str = Field(min_length=1, max_length=63)
    type: FieldType
    required: bool = False
    unique: bool = False
    enum_values: list[str] | None = None
    relation_to: str | None = None
    relation_cardinality: str | None = None


class SetFunctionFieldsRequest(BaseModel):
    fields: list[str] = Field(default_factory=list, max_length=200)
    entity_name: str | None = None


class RegisterRequest(BaseModel):
    email: EmailStr
    password: str = Field(min_length=8, max_length=1024)
    name: str = Field(default="", max_length=100)


class LoginRequest(BaseModel):
    email: EmailStr
    password: str = Field(min_length=1, max_length=1024)


class TokenResponse(BaseModel):
    access_token: str
    token_type: str
    expires_in: int
    developer_id: str
    email: str


class DeveloperResponse(BaseModel):
    id: UUID
    email: str
    name: str


class FieldResponse(BaseModel):
    name: str
    type: FieldType
    system: bool = False
    required: bool = False
    unique: bool = False
    enumValues: list[str] | None = None
    relationTo: str | None = None
    relationCardinality: str | None = None


class EntityResponse(BaseModel):
    name: str
    pool: list[FieldResponse]


class FunctionResponse(BaseModel):
    enabledFields: list[str]


class ModuleResponse(BaseModel):
    enabled: bool
    functions: dict[str, FunctionResponse]
    entityFunctions: dict[str, dict[str, FunctionResponse]] = Field(default_factory=dict)


class ProjectResponse(BaseModel):
    id: UUID
    owner_id: UUID | None
    name: str
    slug: str
    schema_name: str
    version: int
    entities: dict[str, EntityResponse]
    modules: dict[str, ModuleResponse]

    @classmethod
    def from_domain(cls, project: Project) -> "ProjectResponse":
        data: dict[str, Any] = project.to_dict()
        return cls(
            id=project.id,
            owner_id=project.owner_id,
            name=project.name,
            slug=project.slug,
            schema_name=project.schema_name,
            version=project.version,
            entities=data["entities"],
            modules=data["modules"],
        )


class ProjectConfigResponse(BaseModel):
    project_id: UUID
    slug: str
    schema_name: str
    version: int
    entities: dict[str, EntityResponse]
    modules: dict[str, ModuleResponse]

    @classmethod
    def from_output(cls, out: GetProjectConfigOutput) -> "ProjectConfigResponse":
        config: dict[str, Any] = out.config
        return cls(
            project_id=out.project_id,
            slug=out.slug,
            schema_name=out.schema_name,
            version=out.version,
            entities=config["entities"],
            modules=config["modules"],
        )


class ErrorResponse(BaseModel):
    code: str
    message: str
