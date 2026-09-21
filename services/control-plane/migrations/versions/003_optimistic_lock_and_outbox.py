from collections.abc import Sequence

from alembic import op

revision: str = "003"
down_revision: str | None = "002"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.execute(
        "ALTER TABLE control.projects ADD COLUMN version INTEGER NOT NULL DEFAULT 1"
    )
    op.execute(
        """
        CREATE TABLE control.outbox (
            id           BIGSERIAL PRIMARY KEY,
            event_type   TEXT NOT NULL,
            payload      JSONB NOT NULL,
            created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            published_at TIMESTAMPTZ
        )
        """
    )
    op.execute(
        """
        CREATE INDEX idx_outbox_unpublished
        ON control.outbox (id) WHERE published_at IS NULL
        """
    )
    op.execute(
        """
        CREATE INDEX idx_projects_owner_created
        ON control.projects (owner_id, created_at DESC)
        """
    )


def downgrade() -> None:
    op.execute("DROP INDEX IF EXISTS control.idx_projects_owner_created")
    op.execute("DROP TABLE IF EXISTS control.outbox")
    op.execute("ALTER TABLE control.projects DROP COLUMN IF EXISTS version")
