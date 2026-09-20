from uuid import UUID

from pydantic import BaseModel, Field


class CreateProjectRequest(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    slug: str = Field(min_length=3, max_length=40)


class ProjectResponse(BaseModel):
    id: UUID
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
            name=project.name,
            slug=project.slug,
            schema_name=project.schema_name,
            entities=data.get("entities", {}),
            modules=data.get("modules", {}),
        )


class ErrorResponse(BaseModel):
    code: str
    message: str