from dataclasses import dataclass

from app.adapter.security.password import hash_password
from app.domain.developer import Developer
from app.domain.errors import DeveloperAlreadyExistsError
from app.port.developer_repository import DeveloperRepository


@dataclass(frozen=True)
class RegisterDeveloperInput:
    email: str
    password: str
    name: str = ""


@dataclass(frozen=True)
class RegisterDeveloperOutput:
    developer: Developer


class RegisterDeveloper:
    def __init__(self, developer_repository: DeveloperRepository) -> None:
        self._developer_repository = developer_repository

    async def execute(
        self, input_data: RegisterDeveloperInput
    ) -> RegisterDeveloperOutput:
        email = input_data.email.strip().lower()
        if await self._developer_repository.exists_by_email(email):
            raise DeveloperAlreadyExistsError(email)

        developer = Developer.create(
            email=email,
            password_hash=hash_password(input_data.password),
            name=input_data.name,
        )
        await self._developer_repository.save(developer)
        return RegisterDeveloperOutput(developer=developer)