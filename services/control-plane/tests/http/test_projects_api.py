def test_create_project_emits_event(client, headers, outbox):
    r = client.post(
        "/api/v1/projects", json={"name": "Shop", "slug": "shop-one"}, headers=headers
    )
    assert r.status_code == 201
    assert r.json()["schema_name"] == "proj_shop_one"
    assert outbox.types() == ["project.created"]


def test_duplicate_slug_returns_409(client, headers, project_id):
    r = client.post(
        "/api/v1/projects", json={"name": "Other", "slug": "my-shop"}, headers=headers
    )
    assert r.status_code == 409


def test_enum_field_without_values_returns_400(client, headers, project_id):
    r = client.post(
        f"/api/v1/projects/{project_id}/entities/User/fields",
        json={"name": "role", "type": "enum"},
        headers=headers,
    )
    assert r.status_code == 400
    assert r.json()["code"] == "INVALID_FIELD_CONFIG"


def test_relation_without_cardinality_returns_400(client, headers, project_id):
    r = client.post(
        f"/api/v1/projects/{project_id}/entities/User/fields",
        json={"name": "owner", "type": "relation", "relation_to": "User"},
        headers=headers,
    )
    assert r.status_code == 400


def test_bad_cardinality_returns_400(client, headers, project_id):
    r = client.post(
        f"/api/v1/projects/{project_id}/entities/User/fields",
        json={
            "name": "owner",
            "type": "relation",
            "relation_to": "User",
            "relation_cardinality": "n-n",
        },
        headers=headers,
    )
    assert r.status_code == 400


def test_crud_without_entity_name_returns_400(client, headers, project_id):
    client.post(f"/api/v1/projects/{project_id}/modules/crud/enable", headers=headers)
    r = client.put(
        f"/api/v1/projects/{project_id}/modules/crud/functions/create/fields",
        json={"fields": []},
        headers=headers,
    )
    assert r.status_code == 400


def test_unicode_entity_name_returns_400(client, headers, project_id):
    r = client.post(
        f"/api/v1/projects/{project_id}/entities",
        json={"name": "NgườiDùng"},
        headers=headers,
    )
    assert r.status_code == 400
    assert r.json()["code"] == "INVALID_ENTITY_NAME"


def test_disable_required_field_returns_400(client, headers, project_id):
    r = client.put(
        f"/api/v1/projects/{project_id}/modules/auth/functions/signup/fields",
        json={"fields": ["email", "fullName"]},
        headers=headers,
    )
    assert r.status_code == 400
    assert r.json()["code"] == "REQUIRED_FIELD_TOGGLE"


def test_remove_system_field_returns_403(client, headers, project_id):
    r = client.delete(
        f"/api/v1/projects/{project_id}/entities/User/fields/email", headers=headers
    )
    assert r.status_code == 403


def test_field_pool_toggle_flow(client, headers, project_id, outbox):
    r = client.post(
        f"/api/v1/projects/{project_id}/entities/User/fields",
        json={"name": "address", "type": "string"},
        headers=headers,
    )
    assert r.status_code == 201

    r = client.put(
        f"/api/v1/projects/{project_id}/modules/auth/functions/signup/fields",
        json={"fields": ["email", "password", "fullName", "address"]},
        headers=headers,
    )
    assert r.status_code == 200
    signup = r.json()["modules"]["auth"]["functions"]["signup"]["enabledFields"]
    assert "address" in signup

    r = client.delete(
        f"/api/v1/projects/{project_id}/entities/User/fields/address", headers=headers
    )
    assert r.status_code == 409

    r = client.delete(
        f"/api/v1/projects/{project_id}/entities/User/fields/address?force=true",
        headers=headers,
    )
    assert r.status_code == 200
    signup = r.json()["modules"]["auth"]["functions"]["signup"]["enabledFields"]
    assert "address" not in signup
    assert outbox.types().count("project.config.updated") == 3


def test_enable_module_twice_preserves_toggles(client, headers, project_id):
    client.put(
        f"/api/v1/projects/{project_id}/modules/auth/functions/signup/fields",
        json={"fields": ["email", "password"]},
        headers=headers,
    )
    r = client.post(f"/api/v1/projects/{project_id}/modules/auth/enable", headers=headers)
    assert r.status_code == 200
    signup = r.json()["modules"]["auth"]["functions"]["signup"]["enabledFields"]
    assert signup == ["email", "password"]


def test_version_increments_on_mutation(client, headers, project_id):
    r = client.post(
        f"/api/v1/projects/{project_id}/entities",
        json={"name": "Product"},
        headers=headers,
    )
    assert r.json()["version"] == 2


def test_delete_project_emits_event(client, headers, project_id, outbox):
    r = client.delete(f"/api/v1/projects/{project_id}", headers=headers)
    assert r.status_code == 204
    assert outbox.types()[-1] == "project.deleted"
