from dataclasses import dataclass

from app.adapter.security.jwt import create_access_token
from app.adapter.security.password import verify_password
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
    developer_id: str
    email: str


class LoginDeveloper:
    def __init__(
        self,
        developer_repository: DeveloperRepository,
        settings: Settings,
    ) -> None:
        self._developer_repository = developer_repository
        self._settings = settings

    async def execute(self, input_data: LoginDeveloperInput) -> LoginDeveloperOutput:
        developer = await self._developer_repository.get_by_email(input_data.email)
        if developer is None:
            raise InvalidCredentialsError()
        if not verify_password(input_data.password, developer.password_hash):
            raise InvalidCredentialsError()

        token = create_access_token(developer.id, self._settings)
        return LoginDeveloperOutput(
            access_token=token,
            token_type="bearer",
            developer_id=str(developer.id),
            email=developer.email,
        )