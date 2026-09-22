from app.config import get_settings


def test_register_and_login(client):
    r = client.post(
        "/api/v1/auth/register",
        json={"email": "a@b.com", "password": "secret1234", "name": "A"},
    )
    assert r.status_code == 201

    r = client.post(
        "/api/v1/auth/login", json={"email": "a@b.com", "password": "secret1234"}
    )
    assert r.status_code == 200
    body = r.json()
    assert body["token_type"] == "bearer"
    assert body["expires_in"] == get_settings().jwt_expire_minutes * 60


def test_long_unicode_password_is_accepted(client):
    password = "Mật khẩu siêu dài của người Việt Nam có dấu" * 3
    assert len(password.encode("utf-8")) > 72
    r = client.post(
        "/api/v1/auth/register", json={"email": "vn@b.com", "password": password}
    )
    assert r.status_code == 201
    r = client.post(
        "/api/v1/auth/login", json={"email": "vn@b.com", "password": password}
    )
    assert r.status_code == 200


def test_duplicate_email_returns_409(client):
    payload = {"email": "dup@b.com", "password": "secret1234"}
    assert client.post("/api/v1/auth/register", json=payload).status_code == 201
    r = client.post("/api/v1/auth/register", json=payload)
    assert r.status_code == 409
    assert r.json()["code"] == "DEVELOPER_ALREADY_EXISTS"


def test_wrong_password_returns_401(client):
    client.post(
        "/api/v1/auth/register",
        json={"email": "x@b.com", "password": "secret1234"},
    )
    r = client.post(
        "/api/v1/auth/login", json={"email": "x@b.com", "password": "wrongpass"}
    )
    assert r.status_code == 401
    assert r.json()["code"] == "INVALID_CREDENTIALS"


def test_unknown_email_returns_same_error_shape(client):
    r = client.post(
        "/api/v1/auth/login", json={"email": "ghost@b.com", "password": "secret1234"}
    )
    assert r.status_code == 401
    assert r.json()["code"] == "INVALID_CREDENTIALS"


def test_invalid_email_rejected_by_schema(client):
    r = client.post(
        "/api/v1/auth/register", json={"email": "not-an-email", "password": "secret1234"}
    )
    assert r.status_code == 422


def test_missing_token_returns_401(client):
    assert client.get("/api/v1/projects").status_code == 401


def test_garbage_token_returns_401(client):
    r = client.get("/api/v1/projects", headers={"Authorization": "Bearer nope"})
    assert r.status_code == 401
