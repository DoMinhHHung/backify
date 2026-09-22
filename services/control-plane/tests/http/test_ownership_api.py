import pytest


@pytest.fixture
def other_headers(client) -> dict[str, str]:
    client.post(
        "/api/v1/auth/register",
        json={"email": "other@backify.vn", "password": "secret1234"},
    )
    r = client.post(
        "/api/v1/auth/login",
        json={"email": "other@backify.vn", "password": "secret1234"},
    )
    return {"Authorization": f"Bearer {r.json()['access_token']}"}


def test_other_developer_cannot_read(client, project_id, other_headers):
    r = client.get(f"/api/v1/projects/{project_id}", headers=other_headers)
    assert r.status_code == 403


def test_other_developer_cannot_mutate(client, project_id, other_headers):
    r = client.post(
        f"/api/v1/projects/{project_id}/entities",
        json={"name": "Product"},
        headers=other_headers,
    )
    assert r.status_code == 403


def test_other_developer_cannot_delete(client, project_id, other_headers):
    assert (
        client.delete(f"/api/v1/projects/{project_id}", headers=other_headers).status_code
        == 403
    )


def test_list_is_scoped_to_owner(client, headers, project_id, other_headers):
    assert len(client.get("/api/v1/projects", headers=headers).json()) == 1
    assert client.get("/api/v1/projects", headers=other_headers).json() == []
