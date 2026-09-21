import re
from enum import StrEnum
from typing import Self

from app.domain.errors import InvalidFieldConfigError, InvalidFieldNameError


class FieldType(StrEnum):
    UUID = "uuid"
    STRING = "string"
    EMAIL = "email"
    PASSWORD = "password"
    PHONE = "phone"
    DATE = "date"
    INT = "int"
    BOOL = "bool"
    ENUM = "enum"
    RELATION = "relation"


class RelationCardinality(StrEnum):
    ONE_TO_MANY = "1-n"
    MANY_TO_ONE = "n-1"

    @classmethod
    def parse(cls, raw: str) -> "RelationCardinality":
        try:
            return cls(raw)
        except ValueError:
            allowed = ", ".join(c.value for c in cls)
            raise InvalidFieldConfigError(
                f"invalid relation cardinality '{raw}': allowed values are {allowed}"
            ) from None


_FIELD_NAME_PATTERN = re.compile(r"^[a-zA-Z][a-zA-Z0-9_]*$")
_FIELD_NAME_MAX = 63


class Field:
    def __init__(
        self,
        name: str,
        field_type: FieldType,
        *,
        required: bool = False,
        unique: bool = False,
        system: bool = False,
        enum_values: list[str] | None = None,
        relation_to: str | None = None,
        relation_cardinality: RelationCardinality | None = None,
    ) -> None:
        self._validate_name(name)
        if field_type == FieldType.ENUM and not enum_values:
            raise InvalidFieldConfigError(
                f"field '{name}': enum type requires a non-empty enumValues list"
            )
        if field_type != FieldType.ENUM and enum_values is not None:
            raise InvalidFieldConfigError(
                f"field '{name}': enumValues is only allowed for enum type"
            )
        if field_type == FieldType.RELATION:
            if not relation_to:
                raise InvalidFieldConfigError(
                    f"field '{name}': relation type requires relationTo"
                )
            if relation_cardinality is None:
                raise InvalidFieldConfigError(
                    f"field '{name}': relation type requires relationCardinality"
                )
        elif relation_to is not None or relation_cardinality is not None:
            raise InvalidFieldConfigError(
                f"field '{name}': relationTo/relationCardinality are only "
                "allowed for relation type"
            )
        self.name = name
        self.field_type = field_type
        self.required = required
        self.unique = unique
        self.system = system
        self.enum_values = list(enum_values) if enum_values else None
        self.relation_to = relation_to
        self.relation_cardinality = relation_cardinality

    @staticmethod
    def _validate_name(name: str) -> None:
        if not name or not _FIELD_NAME_PATTERN.match(name):
            raise InvalidFieldNameError(
                f"invalid field name '{name}': must start with a letter and "
                "contain only [a-zA-Z0-9_]"
            )
        if len(name) > _FIELD_NAME_MAX:
            raise InvalidFieldNameError(
                f"field name '{name}' exceeds {_FIELD_NAME_MAX} characters"
            )

    def to_dict(self) -> dict[str, object]:
        data: dict[str, object] = {
            "name": self.name,
            "type": self.field_type.value,
            "system": self.system,
        }
        if self.required:
            data["required"] = True
        if self.unique:
            data["unique"] = True
        if self.enum_values is not None:
            data["enumValues"] = self.enum_values
        if self.relation_to is not None:
            data["relationTo"] = self.relation_to
        if self.relation_cardinality is not None:
            data["relationCardinality"] = self.relation_cardinality.value
        return data

    @classmethod
    def from_dict(cls, data: dict[str, object]) -> Self:
        enum_values = data.get("enumValues")
        cardinality_raw = data.get("relationCardinality")
        cardinality = (
            RelationCardinality.parse(str(cardinality_raw))
            if cardinality_raw is not None
            else None
        )
        relation_to = data.get("relationTo")
        return cls(
            name=str(data["name"]),
            field_type=FieldType(str(data["type"])),
            required=bool(data.get("required", False)),
            unique=bool(data.get("unique", False)),
            system=bool(data.get("system", False)),
            enum_values=list(enum_values) if isinstance(enum_values, list) else None,
            relation_to=str(relation_to) if relation_to is not None else None,
            relation_cardinality=cardinality,
        )

    @classmethod
    def system_id(cls) -> Self:
        return cls("id", FieldType.UUID, required=True, unique=True, system=True)

    @classmethod
    def system_email(cls) -> Self:
        return cls("email", FieldType.EMAIL, required=True, unique=True, system=True)

    @classmethod
    def system_password(cls) -> Self:
        return cls("password", FieldType.PASSWORD, required=True, system=True)

    @classmethod
    def system_full_name(cls) -> Self:
        return cls("fullName", FieldType.STRING, system=True)

    @classmethod
    def system_phone(cls) -> Self:
        return cls("phone", FieldType.PHONE, system=True)

    @classmethod
    def default_user_pool(cls) -> list[Self]:
        return [
            cls.system_id(),
            cls.system_email(),
            cls.system_password(),
            cls.system_full_name(),
            cls.system_phone(),
        ]
