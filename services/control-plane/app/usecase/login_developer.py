from dataclasses import dataclass

from app.adapter.security.jwt import create_access_token
from app.adapter.security.password import dummy_hash, verify_password
from app.config import Settings
from app.domain.errors import InvalidCredentialsError
from app.port.developer_repository import DeveloperRepository


@dataclass(frozen=True)
class LoginDeveloperInput:
    email: str
    password: str


@dataclass(frozen=True)
class LoginDeveloperOutput:
    access_token: str
    token_type: str
    expires_in: int
    developer_id: str
    email: str


class LoginDeveloper:
    def __init__(
        self,
        developer_repository: DeveloperRepository,
        settings: Settings,
    ) -> None:
        self._developers = developer_repository
        self._settings = settings

    async def execute(self, input_data: LoginDeveloperInput) -> LoginDeveloperOutput:
        developer = await self._developers.get_by_email(input_data.email)
        stored_hash = developer.password_hash if developer else dummy_hash()
        password_ok = await verify_password(input_data.password, stored_hash)
        if developer is None or not password_ok:
            raise InvalidCredentialsError()

        return LoginDeveloperOutput(
            access_token=create_access_token(developer.id, self._settings),
            token_type="bearer",
            expires_in=self._settings.jwt_expire_minutes * 60,
            developer_id=str(developer.id),
            email=developer.email,
        )
