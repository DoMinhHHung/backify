from uuid import UUID

from pydantic import BaseModel, Field

from app.domain.field import FieldType


class CreateProjectRequest(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    slug: str = Field(min_length=3, max_length=40)


class AddEntityRequest(BaseModel):
    name: str = Field(min_length=1, max_length=64)


class AddFieldRequest(BaseModel):
    name: str = Field(min_length=1, max_length=64)
    type: FieldType
    required: bool = False
    unique: bool = False
    enum_values: list[str] | None = None


class SetFunctionFieldsRequest(BaseModel):
    fields: list[str] = Field(min_length=1)


class RegisterRequest(BaseModel):
    email: str = Field(min_length=3, max_length=255)
    password: str = Field(min_length=8, max_length=128)
    name: str = Field(default="", max_length=100)


class LoginRequest(BaseModel):
    email: str
    password: str


class TokenResponse(BaseModel):
    access_token: str
    token_type: str
    developer_id: str
    email: str


class DeveloperResponse(BaseModel):
    id: UUID
    email: str
    name: str


class ProjectResponse(BaseModel):
    id: UUID
    owner_id: UUID | None
    name: str
    slug: str
    schema_name: str
    entities: dict
    modules: dict

    @classmethod
    def from_domain(cls, project) -> "ProjectResponse":
        data = project.to_dict()
        return cls(
            id=project.id,
            owner_id=project.owner_id,
            name=project.name,
            slug=project.slug,
            schema_name=project.schema_name,
            entities=data.get("entities", {}),
            modules=data.get("modules", {}),
        )


class ErrorResponse(BaseModel):
    code: str
    message: str