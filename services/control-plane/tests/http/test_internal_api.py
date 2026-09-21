from app.config import get_settings


def test_config_requires_key(client, project_id):
    r = client.get(f"/api/v1/projects/{project_id}/config")
    assert r.status_code == 401


def test_config_rejects_wrong_key(client, project_id):
    r = client.get(
        f"/api/v1/projects/{project_id}/config", headers={"X-Internal-Key": "wrong"}
    )
    assert r.status_code == 401


def test_config_with_valid_key_hides_owner(client, project_id):
    r = client.get(
        f"/api/v1/projects/{project_id}/config",
        headers={"X-Internal-Key": get_settings().internal_api_key},
    )
    assert r.status_code == 200
    body = r.json()
    assert "owner_id" not in body
    assert body["schema_name"] == "proj_my_shop"
    assert "User" in body["entities"]


def test_developer_token_does_not_work_on_internal(client, headers, project_id):
    r = client.get(f"/api/v1/projects/{project_id}/config", headers=headers)
    assert r.status_code == 401
