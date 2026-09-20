import json
from uuid import UUID

from app.adapter.postgres.connection import Database
from app.domain.project import Project
from app.port.project_repository import ProjectRepository


class PostgresProjectRepository(ProjectRepository):
    def __init__(self, database: Database) -> None:
        self._db = database

    async def save(self, project: Project) -> None:
        config = project.to_config_dict()
        await self._db.execute(
            """
            INSERT INTO control.projects (id, name, slug, schema_name, config)
            VALUES ($1, $2, $3, $4, $5::jsonb)
            ON CONFLICT (id) DO UPDATE SET
                name = EXCLUDED.name,
                slug = EXCLUDED.slug,
                schema_name = EXCLUDED.schema_name,
                config = EXCLUDED.config,
                updated_at = NOW()
            """,
            project.id,
            project.name,
            project.slug,
            project.schema_name,
            json.dumps(config),
        )

    async def get_by_id(self, project_id: UUID) -> Project | None:
        row = await self._db.fetchrow(
            """
            SELECT id, name, slug, config
            FROM control.projects
            WHERE id = $1
            """,
            project_id,
        )
        if row is None:
            return None
        return self._row_to_project(row)

    async def get_by_slug(self, slug: str) -> Project | None:
        row = await self._db.fetchrow(
            """
            SELECT id, name, slug, config
            FROM control.projects
            WHERE slug = $1
            """,
            slug,
        )
        if row is None:
            return None
        return self._row_to_project(row)

    async def exists_by_slug(self, slug: str) -> bool:
        value = await self._db.fetchval(
            "SELECT EXISTS(SELECT 1 FROM control.projects WHERE slug = $1)",
            slug,
        )
        return bool(value)

    async def delete(self, project_id: UUID) -> bool:
        result = await self._db.execute(
            "DELETE FROM control.projects WHERE id = $1",
            project_id,
        )
        return result == "DELETE 1"

    def _row_to_project(self, row: object) -> Project:
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
        )