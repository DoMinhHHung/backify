import hmac
from typing import Annotated
from uuid import UUID

import jwt
from fastapi import Depends, Header, HTTPException, Request, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer

from app.adapter.security.jwt import decode_access_token
from app.container import Container
from app.domain.errors import UnauthorizedError
from app.usecase.add_entity import AddEntity
from app.usecase.add_field import AddField
from app.usecase.create_project import CreateProject
from app.usecase.delete_project import DeleteProject
from app.usecase.enable_module import EnableModule
from app.usecase.get_project import GetProject
from app.usecase.get_project_config import GetProjectConfig
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
            detail="service unavailable",
        )
    return container


ContainerDep = Annotated[Container, Depends(get_container)]


def get_create_project(c: ContainerDep) -> CreateProject:
    return c.create_project


def get_get_project(c: ContainerDep) -> GetProject:
    return c.get_project


def get_get_project_config(c: ContainerDep) -> GetProjectConfig:
    return c.get_project_config


def get_list_projects(c: ContainerDep) -> ListProjects:
    return c.list_projects


def get_delete_project(c: ContainerDep) -> DeleteProject:
    return c.delete_project


def get_add_entity(c: ContainerDep) -> AddEntity:
    return c.add_entity


def get_add_field(c: ContainerDep) -> AddField:
    return c.add_field


def get_remove_field(c: ContainerDep) -> RemoveField:
    return c.remove_field


def get_set_function_fields(c: ContainerDep) -> SetFunctionFields:
    return c.set_function_fields


def get_enable_module(c: ContainerDep) -> EnableModule:
    return c.enable_module


def get_register_developer(c: ContainerDep) -> RegisterDeveloper:
    return c.register_developer


def get_login_developer(c: ContainerDep) -> LoginDeveloper:
    return c.login_developer


async def get_current_developer_id(
    container: ContainerDep,
    credentials: Annotated[HTTPAuthorizationCredentials | None, Depends(security)] = None,
) -> UUID:
    if credentials is None or not credentials.credentials:
        raise UnauthorizedError("missing bearer token")
    try:
        return decode_access_token(credentials.credentials, container.settings)
    except (jwt.PyJWTError, ValueError, KeyError) as exc:
        raise UnauthorizedError("invalid or expired token") from exc


async def verify_internal_key(
    container: ContainerDep,
    x_internal_key: Annotated[str | None, Header(alias="X-Internal-Key")] = None,
) -> None:
    expected = container.settings.internal_api_key
    if x_internal_key is None or not hmac.compare_digest(x_internal_key, expected):
        raise UnauthorizedError("invalid internal key")
