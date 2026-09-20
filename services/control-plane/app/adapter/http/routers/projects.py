from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, Query, status

from app.adapter.http.dependencies import (
    get_add_entity,
    get_add_field,
    get_create_project,
    get_get_project,
    get_remove_field,
    get_set_function_fields,
)
from app.adapter.http.schemas import (
    AddEntityRequest,
    AddFieldRequest,
    CreateProjectRequest,
    ProjectResponse,
    SetFunctionFieldsRequest,
)
from app.domain.module import FunctionName, ModuleName
from app.usecase.add_entity import AddEntity, AddEntityInput
from app.usecase.add_field import AddField, AddFieldInput
from app.usecase.create_project import CreateProject, CreateProjectInput
from app.usecase.get_project import GetProject, GetProjectInput
from app.usecase.remove_field import RemoveField, RemoveFieldInput
from app.usecase.set_function_fields import SetFunctionFields, SetFunctionFieldsInput

router = APIRouter(prefix="/api/v1/projects", tags=["projects"])


@router.post(
    "",
    response_model=ProjectResponse,
    status_code=status.HTTP_201_CREATED,
)
async def create_project(
    body: CreateProjectRequest,
    usecase: Annotated[CreateProject, Depends(get_create_project)],
) -> ProjectResponse:
    result = await usecase.execute(
        CreateProjectInput(name=body.name, slug=body.slug)
    )
    return ProjectResponse.from_domain(result.project)


@router.get(
    "/{project_id}",
    response_model=ProjectResponse,
)
async def get_project(
    project_id: UUID,
    usecase: Annotated[GetProject, Depends(get_get_project)],
) -> ProjectResponse:
    result = await usecase.execute(GetProjectInput(project_id=project_id))
    return ProjectResponse.from_domain(result.project)


@router.post(
    "/{project_id}/entities",
    response_model=ProjectResponse,
    status_code=status.HTTP_201_CREATED,
)
async def add_entity(
    project_id: UUID,
    body: AddEntityRequest,
    usecase: Annotated[AddEntity, Depends(get_add_entity)],
) -> ProjectResponse:
    result = await usecase.execute(
        AddEntityInput(project_id=project_id, name=body.name)
    )
    return ProjectResponse.from_domain(result.project)


@router.post(
    "/{project_id}/entities/{entity_name}/fields",
    response_model=ProjectResponse,
    status_code=status.HTTP_201_CREATED,
)
async def add_field(
    project_id: UUID,
    entity_name: str,
    body: AddFieldRequest,
    usecase: Annotated[AddField, Depends(get_add_field)],
) -> ProjectResponse:
    result = await usecase.execute(
        AddFieldInput(
            project_id=project_id,
            entity_name=entity_name,
            name=body.name,
            field_type=body.type,
            required=body.required,
            unique=body.unique,
            enum_values=body.enum_values,
        )
    )
    return ProjectResponse.from_domain(result.project)


@router.delete(
    "/{project_id}/entities/{entity_name}/fields/{field_name}",
    response_model=ProjectResponse,
)
async def remove_field(
    project_id: UUID,
    entity_name: str,
    field_name: str,
    usecase: Annotated[RemoveField, Depends(get_remove_field)],
    force: bool = Query(default=False),
) -> ProjectResponse:
    result = await usecase.execute(
        RemoveFieldInput(
            project_id=project_id,
            entity_name=entity_name,
            field_name=field_name,
            force=force,
        )
    )
    return ProjectResponse.from_domain(result.project)


@router.put(
    "/{project_id}/modules/{module}/functions/{function}/fields",
    response_model=ProjectResponse,
)
async def set_function_fields(
    project_id: UUID,
    module: ModuleName,
    function: FunctionName,
    body: SetFunctionFieldsRequest,
    usecase: Annotated[SetFunctionFields, Depends(get_set_function_fields)],
) -> ProjectResponse:
    result = await usecase.execute(
        SetFunctionFieldsInput(
            project_id=project_id,
            module=module,
            function=function,
            field_names=body.fields,
        )
    )
    return ProjectResponse.from_domain(result.project)