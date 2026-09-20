from enum import StrEnum
from typing import Self
import re

from app.domain.errors import InvalidFieldNameError


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


_FIELD_NAME_PATTERN = re.compile(r"^[a-zA-Z][a-zA-Z0-9_]*$")


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
    ) -> None:
        self._validate_name(name)
        if field_type == FieldType.ENUM and not enum_values:
            raise ValueError("enum field requires enum_values")
        if field_type != FieldType.ENUM and enum_values is not None:
            raise ValueError("enum_values only allowed for enum type")
        self.name = name
        self.field_type = field_type
        self.required = required
        self.unique = unique
        self.system = system
        self.enum_values = list(enum_values) if enum_values else None

    @staticmethod
    def _validate_name(name: str) -> None:
        if not name or not _FIELD_NAME_PATTERN.match(name):
            raise InvalidFieldNameError(
                f"invalid field name '{name}': must start with letter, "
                "only [a-zA-Z0-9_], non-empty"
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
        return data

    @classmethod
    def from_dict(cls, data: dict[str, object]) -> Self:
        enum_values = data.get("enumValues")
        return cls(
            name=str(data["name"]),
            field_type=FieldType(str(data["type"])),
            required=bool(data.get("required", False)),
            unique=bool(data.get("unique", False)),
            system=bool(data.get("system", False)),
            enum_values=list(enum_values) if isinstance(enum_values, list) else None,
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