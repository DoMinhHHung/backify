from datetime import UTC, datetime, timedelta
from uuid import UUID

import jwt

from app.config import Settings

_ISSUER = "backify-control-plane"


def create_access_token(developer_id: UUID, settings: Settings) -> str:
    now = datetime.now(UTC)
    payload = {
        "sub": str(developer_id),
        "iss": _ISSUER,
        "iat": now,
        "exp": now + timedelta(minutes=settings.jwt_expire_minutes),
    }
    return jwt.encode(payload, settings.jwt_secret, algorithm=settings.jwt_algorithm)


def decode_access_token(token: str, settings: Settings) -> UUID:
    payload = jwt.decode(
        token,
        settings.jwt_secret,
        algorithms=[settings.jwt_algorithm],
        issuer=_ISSUER,
        options={"require": ["exp", "iat", "sub", "iss"]},
    )
    return UUID(str(payload["sub"]))
