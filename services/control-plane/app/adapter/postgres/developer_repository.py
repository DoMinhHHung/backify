from uuid import UUID

from app.adapter.postgres.connection import Database
from app.domain.developer import Developer
from app.port.developer_repository import DeveloperRepository


class PostgresDeveloperRepository(DeveloperRepository):
    def __init__(self, database: Database) -> None:
        self._db = database

    async def save(self, developer: Developer) -> None:
        await self._db.execute(
            """
            INSERT INTO control.developers (id, email, password_hash, name)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (id) DO UPDATE SET
                email = EXCLUDED.email,
                password_hash = EXCLUDED.password_hash,
                name = EXCLUDED.name,
                updated_at = NOW()
            """,
            developer.id,
            developer.email,
            developer.password_hash,
            developer.name,
        )

    async def get_by_id(self, developer_id: UUID) -> Developer | None:
        row = await self._db.fetchrow(
            """
            SELECT id, email, password_hash, name
            FROM control.developers
            WHERE id = $1
            """,
            developer_id,
        )
        if row is None:
            return None
        return Developer(
            email=row["email"],
            password_hash=row["password_hash"],
            developer_id=row["id"],
            name=row["name"] or "",
        )

    async def get_by_email(self, email: str) -> Developer | None:
        row = await self._db.fetchrow(
            """
            SELECT id, email, password_hash, name
            FROM control.developers
            WHERE email = $1
            """,
            email.strip().lower(),
        )
        if row is None:
            return None
        return Developer(
            email=row["email"],
            password_hash=row["password_hash"],
            developer_id=row["id"],
            name=row["name"] or "",
        )

    async def exists_by_email(self, email: str) -> bool:
        value = await self._db.fetchval(
            "SELECT EXISTS(SELECT 1 FROM control.developers WHERE email = $1)",
            email.strip().lower(),
        )
        return bool(value)