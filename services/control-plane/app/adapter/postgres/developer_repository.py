from typing import Any
from uuid import UUID

import asyncpg

from app.adapter.postgres.connection import Database
from app.domain.developer import Developer
from app.domain.errors import DeveloperAlreadyExistsError


class PostgresDeveloperRepository:
    def __init__(self, database: Database) -> None:
        self._db = database

    def _executor(self, conn: Any) -> Any:
        return conn if conn is not None else self._db

    async def create(self, developer: Developer, conn: Any = None) -> None:
        try:
            await self._executor(conn).execute(
                """
                INSERT INTO control.developers (id, email, password_hash, name)
                VALUES ($1, $2, $3, $4)
                """,
                developer.id,
                developer.email,
                developer.password_hash,
                developer.name,
            )
        except asyncpg.UniqueViolationError as exc:
            raise DeveloperAlreadyExistsError(developer.email) from exc

    async def get_by_id(self, developer_id: UUID) -> Developer | None:
        row = await self._db.fetchrow(
            """
            SELECT id, email, password_hash, name
            FROM control.developers WHERE id = $1
            """,
            developer_id,
        )
        return self._row_to_developer(row) if row else None

    async def get_by_email(self, email: str) -> Developer | None:
        row = await self._db.fetchrow(
            """
            SELECT id, email, password_hash, name
            FROM control.developers WHERE email = $1
            """,
            email.strip().lower(),
        )
        return self._row_to_developer(row) if row else None

    async def exists_by_email(self, email: str) -> bool:
        value = await self._db.fetchval(
            "SELECT EXISTS(SELECT 1 FROM control.developers WHERE email = $1)",
            email.strip().lower(),
        )
        return bool(value)

    @staticmethod
    def _row_to_developer(row: Any) -> Developer:
        return Developer(
            email=row["email"],
            password_hash=row["password_hash"],
            developer_id=row["id"],
            name=row["name"] or "",
        )
