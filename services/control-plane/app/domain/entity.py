from typing import Self

from app.domain.errors import (
    FieldAlreadyExistsError,
    FieldNotFoundError,
    InvalidEntityNameError,
    SystemFieldProtectedError,
)
from app.domain.field import Field


class Entity:
    def __init__(self, name: str, fields: list[Field] | None = None) -> None:
        if not name or not name[0].isalpha() or not name.replace("_", "").isalnum():
            raise InvalidEntityNameError(
                f"invalid entity name '{name}': must start with letter, only [a-zA-Z0-9_]"
            )
        self.name = name
        self._fields: dict[str, Field] = {}
        if fields:
            for field in fields:
                self.add_field(field)

    @property
    def fields(self) -> list[Field]:
        return list(self._fields.values())

    def get_field(self, name: str) -> Field | None:
        return self._fields.get(name)

    def has_field(self, name: str) -> bool:
        return name in self._fields

    def add_field(self, field: Field) -> None:
        if field.name in self._fields:
            raise FieldAlreadyExistsError(field.name)
        self._fields[field.name] = field

    def remove_field(self, name: str) -> Field:
        field = self._fields.get(name)
        if field is None:
            raise FieldNotFoundError(name)
        if field.system:
            raise SystemFieldProtectedError(name)
        del self._fields[name]
        return field

    def to_dict(self) -> dict[str, object]:
        return {
            "name": self.name,
            "pool": [f.to_dict() for f in self._fields.values()],
        }

    @classmethod
    def from_dict(cls, data: dict[str, object]) -> Self:
        pool_raw = data.get("pool", [])
        fields: list[Field] = []
        if isinstance(pool_raw, list):
            for item in pool_raw:
                if isinstance(item, dict):
                    fields.append(Field.from_dict(item))
        return cls(str(data["name"]), fields)

    @classmethod
    def user_with_default_pool(cls) -> Self:
        return cls("User", Field.default_user_pool())