from typing import Self
from uuid import UUID, uuid4


class Developer:
    def __init__(
        self,
        email: str,
        password_hash: str,
        *,
        developer_id: UUID | None = None,
        name: str = "",
    ) -> None:
        if not email or "@" not in email:
            raise ValueError("invalid email")
        self.id = developer_id or uuid4()
        self.email = email.strip().lower()
        self.password_hash = password_hash
        self.name = name.strip()

    @classmethod
    def create(cls, email: str, password_hash: str, name: str = "") -> Self:
        return cls(email=email, password_hash=password_hash, name=name)