from typing import Annotated
from uuid import UUID

from fastapi import APIRouter, Depends, status

from app.adapter.http.dependencies import get_create_project, get_get_project
from app.adapter.http.schemas import CreateProjectRequest, ProjectResponse
from app.usecase.create_project import CreateProject, CreateProjectInput
from app.usecase.get_project import GetProject, GetProjectInput

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