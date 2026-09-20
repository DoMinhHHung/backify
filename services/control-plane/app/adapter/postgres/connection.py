from typing import Any
from urllib.parse import parse_qs, urlencode, urlparse, urlunparse

import asyncpg
import structlog

from app.config import Settings

logger = structlog.get_logger()


def _clean_dsn(dsn: str) -> str:
    parsed = urlparse(dsn)
    query = parse_qs(parsed.query)
    query.pop("sslmode", None)
    query.pop("ssl", None)
    clean_query = urlencode({k: v[0] for k, v in query.items()}) if query else ""
    return urlunparse(parsed._replace(query=clean_query))


class Database:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._pool: asyncpg.Pool | None = None

    @property
    def pool(self) -> asyncpg.Pool:
        if self._pool is None:
            raise RuntimeError("database pool is not connected")
        return self._pool

    async def connect(self) -> None:
        if self._pool is not None:
            return
        dsn = _clean_dsn(self._settings.database_url)
        self._pool = await asyncpg.create_pool(
            dsn=dsn,
            min_size=self._settings.database_pool_min_size,
            max_size=self._settings.database_pool_max_size,
            ssl="require",
        )
        async with self._pool.acquire() as conn:
            await conn.fetchval("SELECT 1")
        logger.info("db_pool_connected")

    async def disconnect(self) -> None:
        if self._pool is not None:
            await self._pool.close()
            self._pool = None
            logger.info("db_pool_closed")

    async def execute(self, query: str, *args: Any) -> str:
        async with self.pool.acquire() as conn:
            return await conn.execute(query, *args)

    async def fetch(self, query: str, *args: Any) -> list[asyncpg.Record]:
        async with self.pool.acquire() as conn:
            return await conn.fetch(query, *args)

    async def fetchrow(self, query: str, *args: Any) -> asyncpg.Record | None:
        async with self.pool.acquire() as conn:
            return await conn.fetchrow(query, *args)

    async def fetchval(self, query: str, *args: Any) -> Any:
        async with self.pool.acquire() as conn:
            return await conn.fetchval(query, *args)

    async def health_check(self) -> bool:
        try:
            val = await self.fetchval("SELECT 1")
            return val == 1
        except Exception:
            return False