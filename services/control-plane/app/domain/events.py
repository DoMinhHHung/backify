from dataclasses import dataclass
from datetime import datetime, timezone
from enum import StrEnum
from uuid import UUID


class ProjectEventType(StrEnum):
    CREATED = "project.created"
    CONFIG_UPDATED = "project.config.updated"
    DELETED = "project.deleted"


@dataclass(frozen=True)
class ProjectEvent:
    event_type: ProjectEventType
    project_id: UUID
    slug: str
    schema_name: str
    occurred_at: datetime

    @classmethod
    def created(cls, project_id: UUID, slug: str, schema_name: str) -> "ProjectEvent":
        return cls(
            event_type=ProjectEventType.CREATED,
            project_id=project_id,
            slug=slug,
            schema_name=schema_name,
            occurred_at=datetime.now(timezone.utc),
        )

    @classmethod
    def config_updated(
        cls, project_id: UUID, slug: str, schema_name: str
    ) -> "ProjectEvent":
        return cls(
            event_type=ProjectEventType.CONFIG_UPDATED,
            project_id=project_id,
            slug=slug,
            schema_name=schema_name,
            occurred_at=datetime.now(timezone.utc),
        )

    @classmethod
    def deleted(cls, project_id: UUID, slug: str, schema_name: str) -> "ProjectEvent":
        return cls(
            event_type=ProjectEventType.DELETED,
            project_id=project_id,
            slug=slug,
            schema_name=schema_name,
            occurred_at=datetime.now(timezone.utc),
        )

    def to_payload(self) -> dict[str, object]:
        return {
            "event": self.event_type.value,
            "projectId": str(self.project_id),
            "slug": self.slug,
            "schemaName": self.schema_name,
            "occurredAt": self.occurred_at.isoformat(),
        }