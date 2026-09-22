import re
from typing import Self
from uuid import UUID, uuid4

from app.domain.errors import DomainError

_EMAIL_PATTERN = re.compile(r"^[^@\s]+@[^@\s]+\.[^@\s]+$")


class InvalidEmailError(DomainError):
    def __init__(self, email: str) -> None:
        super().__init__("INVALID_EMAIL", f"invalid email '{email}'")


class Developer:
    def __init__(
        self,
        email: str,
        password_hash: str,
        *,
        developer_id: UUID | None = None,
        name: str = "",
    ) -> None:
        normalized = email.strip().lower()
        if not _EMAIL_PATTERN.match(normalized):
            raise InvalidEmailError(email)
        self.id = developer_id or uuid4()
        self.email = normalized
        self.password_hash = password_hash
        self.name = name.strip()

    @classmethod
    def create(cls, email: str, password_hash: str, name: str = "") -> Self:
        return cls(email=email, password_hash=password_hash, name=name)
