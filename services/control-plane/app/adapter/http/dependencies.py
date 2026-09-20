from typing import Annotated

from fastapi import Depends, Request

from app.container import Container
from app.usecase.add_entity import AddEntity
from app.usecase.add_field import AddField
from app.usecase.create_project import CreateProject
from app.usecase.get_project import GetProject
from app.usecase.remove_field import RemoveField
from app.usecase.set_function_fields import SetFunctionFields


def get_container(request: Request) -> Container:
    return request.app.state.container


def get_create_project(
    container: Annotated[Container, Depends(get_container)],
) -> CreateProject:
    return container.create_project


def get_get_project(
    container: Annotated[Container, Depends(get_container)],
) -> GetProject:
    return container.get_project


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

def get_delete_project(
    container: Annotated[Container, Depends(get_container)],
) -> DeleteProject:
    return container.delete_project