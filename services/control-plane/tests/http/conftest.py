import pytest
from app.adapter.http.dependencies import get_container
from app.config import get_settings
from app.container import Container
from app.main import create_app
from fastapi.testclient import TestClient

from tests.fakes import (
    FakeDeveloperRepository,
    FakeOutboxRepository,
    FakeProjectRepository,
    FakeTransactionManager,
)


@pytest.fixture
def outbox() -> FakeOutboxRepository:
    return FakeOutboxRepository()


@pytest.fixture
def container(outbox: FakeOutboxRepository) -> Container:
    return Container(
        get_settings(),
        projects=FakeProjectRepository(),
        developers=FakeDeveloperRepository(),
        outbox=outbox,
        tx=FakeTransactionManager(),
    )


@pytest.fixture
def client(container: Container):
    application = create_app()
    application.dependency_overrides[get_container] = lambda: container
    return TestClient(application, raise_server_exceptions=False)


@pytest.fixture
def headers(client) -> dict[str, str]:
    client.post(
        "/api/v1/auth/register",
        json={"email": "dev@backify.vn", "password": "secret1234"},
    )
    r = client.post(
        "/api/v1/auth/login",
        json={"email": "dev@backify.vn", "password": "secret1234"},
    )
    return {"Authorization": f"Bearer {r.json()['access_token']}"}


@pytest.fixture
def project_id(client, headers) -> str:
    r = client.post(
        "/api/v1/projects",
        json={"name": "My Shop", "slug": "my-shop"},
        headers=headers,
    )
    return r.json()["id"]
