from enum import StrEnum
from typing import Self


class ModuleName(StrEnum):
    AUTH = "auth"


class FunctionName(StrEnum):
    SIGNUP = "signup"
    SIGNIN = "signin"
    FORGOT_PASSWORD = "forgot_password"


class FunctionConfig:
    def __init__(
        self,
        function: FunctionName,
        enabled_fields: list[str] | None = None,
    ) -> None:
        self.function = function
        self._enabled_fields: set[str] = set(enabled_fields or [])

    @property
    def enabled_fields(self) -> list[str]:
        return sorted(self._enabled_fields)

    def is_enabled(self, field_name: str) -> bool:
        return field_name in self._enabled_fields

    def enable_field(self, field_name: str) -> None:
        self._enabled_fields.add(field_name)

    def disable_field(self, field_name: str) -> None:
        self._enabled_fields.discard(field_name)

    def to_dict(self) -> dict[str, object]:
        return {"enabledFields": self.enabled_fields}

    @classmethod
    def from_dict(cls, function: FunctionName, data: dict[str, object]) -> Self:
        raw = data.get("enabledFields", [])
        fields = [str(f) for f in raw] if isinstance(raw, list) else []
        return cls(function, fields)


class ModuleConfig:
    def __init__(
        self,
        module: ModuleName,
        *,
        enabled: bool = False,
        functions: dict[FunctionName, FunctionConfig] | None = None,
    ) -> None:
        self.module = module
        self.enabled = enabled
        self._functions: dict[FunctionName, FunctionConfig] = functions or {}

    def get_function(self, function: FunctionName) -> FunctionConfig | None:
        return self._functions.get(function)

    def set_function(self, config: FunctionConfig) -> None:
        self._functions[config.function] = config

    @property
    def functions(self) -> dict[FunctionName, FunctionConfig]:
        return dict(self._functions)

    def to_dict(self) -> dict[str, object]:
        return {
            "enabled": self.enabled,
            "functions": {
                fn.value: cfg.to_dict() for fn, cfg in self._functions.items()
            },
        }

    @classmethod
    def from_dict(cls, module: ModuleName, data: dict[str, object]) -> Self:
        enabled = bool(data.get("enabled", False))
        functions: dict[FunctionName, FunctionConfig] = {}
        raw_functions = data.get("functions", {})
        if isinstance(raw_functions, dict):
            for key, value in raw_functions.items():
                try:
                    fn = FunctionName(str(key))
                except ValueError:
                    continue
                if isinstance(value, dict):
                    functions[fn] = FunctionConfig.from_dict(fn, value)
        return cls(module, enabled=enabled, functions=functions)

    @classmethod
    def default_auth(cls) -> Self:
        signup = FunctionConfig(
            FunctionName.SIGNUP,
            enabled_fields=["email", "password", "fullName"],
        )
        signin = FunctionConfig(
            FunctionName.SIGNIN,
            enabled_fields=["email", "password"],
        )
        forgot = FunctionConfig(
            FunctionName.FORGOT_PASSWORD,
            enabled_fields=["email"],
        )
        return cls(
            ModuleName.AUTH,
            enabled=True,
            functions={
                FunctionName.SIGNUP: signup,
                FunctionName.SIGNIN: signin,
                FunctionName.FORGOT_PASSWORD: forgot,
            },
        )