from typing import Annotated
from uuid import UUID

import jwt
from fastapi import Depends, HTTPException, Request, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer

from app.adapter.security.jwt import decode_access_token
from app.container import Container
from app.domain.errors import UnauthorizedError
from app.usecase.add_entity import AddEntity
from app.usecase.add_field import AddField
from app.usecase.create_project import CreateProject
from app.usecase.delete_project import DeleteProject
from app.usecase.get_project import GetProject
from app.usecase.list_projects import ListProjects
from app.usecase.login_developer import LoginDeveloper
from app.usecase.register_developer import RegisterDeveloper
from app.usecase.remove_field import RemoveField
from app.usecase.set_function_fields import SetFunctionFields

security = HTTPBearer(auto_error=False)



def get_container(request: Request) -> Container:
    container = getattr(request.app.state, "container", None)
    if container is None:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="service unavailable: database not connected",
        )
    return container


def get_create_project(
    container: Annotated[Container, Depends(get_container)],
) -> CreateProject:
    return container.create_project


def get_get_project(
    container: Annotated[Container, Depends(get_container)],
) -> GetProject:
    return container.get_project


def get_list_projects(
    container: Annotated[Container, Depends(get_container)],
) -> ListProjects:
    return container.list_projects


def get_delete_project(
    container: Annotated[Container, Depends(get_container)],
) -> DeleteProject:
    return container.delete_project


def get_add_entity(
    container: Annotated[Container, Depends(get_container)],
) -> AddEntity:
    return container.add_entity


def get_add_field(
    container: Annotated[Container, Depends(get_container)],
) -> AddField:
    return container.add_field


def get_remove_field(
    container: Annotated[Container, Depends(get_container)],
) -> RemoveField:
    return container.remove_field


def get_set_function_fields(
    container: Annotated[Container, Depends(get_container)],
) -> SetFunctionFields:
    return container.set_function_fields


def get_register_developer(
    container: Annotated[Container, Depends(get_container)],
) -> RegisterDeveloper:
    return container.register_developer


def get_login_developer(
    container: Annotated[Container, Depends(get_container)],
) -> LoginDeveloper:
    return container.login_developer


async def get_current_developer_id(
    credentials: Annotated[
        HTTPAuthorizationCredentials | None, Depends(security)
    ],
    container: Annotated[Container, Depends(get_container)],
) -> UUID:
    if credentials is None or not credentials.credentials:
        raise UnauthorizedError()
    try:
        return decode_access_token(credentials.credentials, container.settings)
    except (jwt.PyJWTError, ValueError):
        raise UnauthorizedError("invalid or expired token")