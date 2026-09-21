from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, Query, status

from app.adapter.http.dependencies import (
    get_add_entity,
    get_add_field,
    get_create_project,
    get_current_developer_id,
    get_delete_project,
    get_get_project,
    get_list_projects,
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
from app.usecase.delete_project import DeleteProject, DeleteProjectInput
from app.usecase.get_project import GetProject, GetProjectInput
from app.usecase.list_projects import ListProjects, ListProjectsInput
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
    owner_id: Annotated[UUID, Depends(get_current_developer_id)],
    usecase: Annotated[CreateProject, Depends(get_create_project)],
) -> ProjectResponse:
    result = await usecase.execute(
        CreateProjectInput(name=body.name, slug=body.slug, owner_id=owner_id)
    )
    return ProjectResponse.from_domain(result.project)


@router.get(
    "",
    response_model=list[ProjectResponse],
)
async def list_projects(
    owner_id: Annotated[UUID, Depends(get_current_developer_id)],
    usecase: Annotated[ListProjects, Depends(get_list_projects)],
) -> list[ProjectResponse]:
    result = await usecase.execute(ListProjectsInput(owner_id=owner_id))
    return [ProjectResponse.from_domain(p) for p in result.projects]


@router.get(
    "/{project_id}",
    response_model=ProjectResponse,
)
async def get_project(
    project_id: UUID,
    owner_id: Annotated[UUID, Depends(get_current_developer_id)],
    usecase: Annotated[GetProject, Depends(get_get_project)],
) -> ProjectResponse:
    result = await usecase.execute(
        GetProjectInput(project_id=project_id, owner_id=owner_id)
    )
    return ProjectResponse.from_domain(result.project)


@router.delete(
    "/{project_id}",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def delete_project(
    project_id: UUID,
    owner_id: Annotated[UUID, Depends(get_current_developer_id)],
    usecase: Annotated[DeleteProject, Depends(get_delete_project)],
) -> None:
    await usecase.execute(
        DeleteProjectInput(project_id=project_id, owner_id=owner_id)
    )


@router.post(
    "/{project_id}/entities",
    response_model=ProjectResponse,
    status_code=status.HTTP_201_CREATED,
)
async def add_entity(
    project_id: UUID,
    body: AddEntityRequest,
    owner_id: Annotated[UUID, Depends(get_current_developer_id)],
    usecase: Annotated[AddEntity, Depends(get_add_entity)],
) -> ProjectResponse:
    result = await usecase.execute(
        AddEntityInput(project_id=project_id, owner_id=owner_id, name=body.name)
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
    owner_id: Annotated[UUID, Depends(get_current_developer_id)],
    usecase: Annotated[AddField, Depends(get_add_field)],
) -> ProjectResponse:
    result = await usecase.execute(
        AddFieldInput(
            project_id=project_id,
            owner_id=owner_id,
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
    owner_id: Annotated[UUID, Depends(get_current_developer_id)],
    usecase: Annotated[RemoveField, Depends(get_remove_field)],
    force: bool = Query(default=False),
) -> ProjectResponse:
    result = await usecase.execute(
        RemoveFieldInput(
            project_id=project_id,
            owner_id=owner_id,
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
    owner_id: Annotated[UUID, Depends(get_current_developer_id)],
    usecase: Annotated[SetFunctionFields, Depends(get_set_function_fields)],
) -> ProjectResponse:
    result = await usecase.execute(
        SetFunctionFieldsInput(
            project_id=project_id,
            owner_id=owner_id,
            module=module,
            function=function,
            field_names=body.fields,
        )
    )
    return ProjectResponse.from_domain(result.project)

@router.post(
    "/{project_id}/modules/{module}/enable",
    response_model=ProjectResponse,
)
async def enable_module(
    project_id: UUID,
    module: ModuleName,
    owner_id: Annotated[UUID, Depends(get_current_developer_id)],
    usecase: Annotated[EnableModule, Depends(get_enable_module)],
) -> ProjectResponse:
    result = await usecase.execute(
        EnableModuleInput(
            project_id=project_id, owner_id=owner_id, module=module
        )
    )
    return ProjectResponse.from_domain(result.project)

@router.get(
    "/{project_id}/config",
    response_model=ProjectResponse,
    tags=["internal"],
)
async def get_project_config(
    project_id: UUID,
    _: Annotated[None, Depends(verify_internal_key)],
    usecase: Annotated[GetProjectConfig, Depends(get_get_project_config)],
) -> ProjectResponse:
    result = await usecase.execute(GetProjectConfigInput(project_id=project_id))
    return ProjectResponse.from_domain(result.project)