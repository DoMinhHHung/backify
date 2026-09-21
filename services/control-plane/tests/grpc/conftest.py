from uuid import uuid4

import grpc
import pytest
import pytest_asyncio
from app.adapter.grpc.pb import control_plane_pb2_grpc
from app.adapter.grpc.server import create_grpc_server
from app.config import get_settings
from app.container import Container
from app.domain.project import Project

from tests.fakes import (
    FakeDeveloperRepository,
    FakeOutboxRepository,
    FakeProjectRepository,
    FakeTransactionManager,
)


@pytest.fixture
def project_repo() -> FakeProjectRepository:
    return FakeProjectRepository()


@pytest.fixture
def grpc_container(project_repo: FakeProjectRepository) -> Container:
    return Container(
        get_settings(),
        projects=project_repo,
        developers=FakeDeveloperRepository(),
        outbox=FakeOutboxRepository(),
        tx=FakeTransactionManager(),
    )


@pytest_asyncio.fixture
async def grpc_channel(grpc_container: Container):
    settings = get_settings()
    server = create_grpc_server(grpc_container, settings)
    port = server.add_insecure_port("127.0.0.1:0")
    await server.start()
    channel = grpc.aio.insecure_channel(f"127.0.0.1:{port}")
    try:
        yield channel
    finally:
        await channel.close()
        await server.stop(grace=None)


@pytest.fixture
def stub(
    grpc_channel: grpc.aio.Channel,
) -> control_plane_pb2_grpc.ControlPlaneServiceStub:
    return control_plane_pb2_grpc.ControlPlaneServiceStub(grpc_channel)


@pytest.fixture
def internal_key_metadata() -> tuple[tuple[str, str], ...]:
    return (("x-internal-key", get_settings().internal_api_key),)


@pytest_asyncio.fixture
async def existing_project(project_repo: FakeProjectRepository) -> Project:
    project = Project.create("Shop", "grpc-probe", owner_id=uuid4())
    await project_repo.create(project)
    return project
