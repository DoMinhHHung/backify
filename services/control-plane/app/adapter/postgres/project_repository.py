import json
from typing import Any
from uuid import UUID

import asyncpg

from app.adapter.postgres.connection import Database
from app.domain.errors import ConcurrentModificationError, ProjectSlugExistsError
from app.domain.project import Project


class PostgresProjectRepository:
    def __init__(self, database: Database) -> None:
        self._db = database

    def _executor(self, conn: Any) -> Any:
        return conn if conn is not None else self._db

    async def create(self, project: Project, conn: Any = None) -> None:
        try:
            await self._executor(conn).execute(
                """
                INSERT INTO control.projects
                    (id, name, slug, schema_name, config, owner_id, version)
                VALUES ($1, $2, $3, $4, $5::jsonb, $6, 1)
                """,
                project.id,
                project.name,
                project.slug,
                project.schema_name,
                json.dumps(project.to_config_dict()),
                project.owner_id,
            )
        except asyncpg.UniqueViolationError as exc:
            if "slug" in str(exc) or "schema_name" in str(exc):
                raise ProjectSlugExistsError(project.slug) from exc
            raise
        project.version = 1

    async def save(self, project: Project, conn: Any = None) -> None:
        result = await self._executor(conn).execute(
            """
            UPDATE control.projects
            SET name = $2,
                config = $3::jsonb,
                version = version + 1,
                updated_at = NOW()
            WHERE id = $1 AND version = $4
            """,
            project.id,
            project.name,
            json.dumps(project.to_config_dict()),
            project.version,
        )
        if result == "UPDATE 0":
            raise ConcurrentModificationError(str(project.id))
        project.version += 1

    async def get_by_id(self, project_id: UUID) -> Project | None:
        row = await self._db.fetchrow(
            """
            SELECT id, name, slug, config, owner_id, version
            FROM control.projects WHERE id = $1
            """,
            project_id,
        )
        return self._row_to_project(row) if row else None

    async def get_by_slug(self, slug: str) -> Project | None:
        row = await self._db.fetchrow(
            """
            SELECT id, name, slug, config, owner_id, version
            FROM control.projects WHERE slug = $1
            """,
            slug,
        )
        return self._row_to_project(row) if row else None

    async def exists_by_slug(self, slug: str) -> bool:
        value = await self._db.fetchval(
            "SELECT EXISTS(SELECT 1 FROM control.projects WHERE slug = $1)",
            slug,
        )
        return bool(value)

    async def delete(self, project_id: UUID, conn: Any = None) -> bool:
        result = await self._executor(conn).execute(
            "DELETE FROM control.projects WHERE id = $1",
            project_id,
        )
        return result == "DELETE 1"

    async def list_by_owner(self, owner_id: UUID) -> list[Project]:
        rows = await self._db.fetch(
            """
            SELECT id, name, slug, config, owner_id, version
            FROM control.projects
            WHERE owner_id = $1
            ORDER BY created_at DESC
            """,
            owner_id,
        )
        return [self._row_to_project(row) for row in rows]

    @staticmethod
    def _row_to_project(row: Any) -> Project:
        config = row["config"]
        if isinstance(config, str):
            config = json.loads(config)
        if not isinstance(config, dict):
            config = {}
        return Project.from_persistence(
            project_id=row["id"],
            name=row["name"],
            slug=row["slug"],
            config=config,
            owner_id=row["owner_id"],
            version=row["version"],
        )
