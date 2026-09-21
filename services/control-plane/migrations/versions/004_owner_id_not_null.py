"""enforce owner_id NOT NULL on control.projects

Revision ID: 004
Revises: 003
Create Date: 2026-09-21
"""

from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op

revision: str = "004"
down_revision: str | None = "003"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    connection = op.get_bind()
    orphan_count = connection.execute(
        sa.text("SELECT COUNT(*) FROM control.projects WHERE owner_id IS NULL")
    ).scalar_one()
    if orphan_count:
        raise RuntimeError(
            f"{orphan_count} row(s) in control.projects have owner_id IS NULL. "
            "Backfill an owner or delete these rows manually before running "
            "this migration, e.g.:\n"
            "  SELECT id, slug FROM control.projects WHERE owner_id IS NULL;"
        )
    op.execute("ALTER TABLE control.projects ALTER COLUMN owner_id SET NOT NULL")


def downgrade() -> None:
    op.execute("ALTER TABLE control.projects ALTER COLUMN owner_id DROP NOT NULL")
