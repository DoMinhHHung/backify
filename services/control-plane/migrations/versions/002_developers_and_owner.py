"""developers table and project owner_id

Revision ID: 002
Revises: 001
Create Date: 2026-09-20
"""

from typing import Sequence, Union

from alembic import op

revision: str = "002"
down_revision: Union[str, None] = "001"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.execute(
        """
        CREATE TABLE control.developers (
            id              UUID PRIMARY KEY,
            email           TEXT NOT NULL UNIQUE,
            password_hash   TEXT NOT NULL,
            name            TEXT NOT NULL DEFAULT '',
            created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
        )
        """
    )
    op.execute(
        """
        ALTER TABLE control.projects
        ADD COLUMN owner_id UUID REFERENCES control.developers(id)
        """
    )
    op.execute(
        "CREATE INDEX idx_projects_owner_id ON control.projects (owner_id)"
    )


def downgrade() -> None:
    op.execute("ALTER TABLE control.projects DROP COLUMN IF EXISTS owner_id")
    op.execute("DROP TABLE IF EXISTS control.developers")