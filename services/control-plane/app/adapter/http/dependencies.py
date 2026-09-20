from typing import Annotated

from fastapi import Depends, Request

from app.container import Container
from app.usecase.create_project import CreateProject
from app.usecase.get_project import GetProject


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