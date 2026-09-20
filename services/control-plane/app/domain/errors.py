class DomainError(Exception):
    def __init__(self, code: str, message: str) -> None:
        self.code = code
        self.message = message
        super().__init__(message)


class ProjectNotFoundError(DomainError):
    def __init__(self, project_id: str) -> None:
        super().__init__("PROJECT_NOT_FOUND", f"project '{project_id}' not found")


class ProjectSlugExistsError(DomainError):
    def __init__(self, slug: str) -> None:
        super().__init__("PROJECT_SLUG_EXISTS", f"slug '{slug}' already exists")


class EntityNotFoundError(DomainError):
    def __init__(self, entity_name: str) -> None:
        super().__init__("ENTITY_NOT_FOUND", f"entity '{entity_name}' not found")


class FieldNotFoundError(DomainError):
    def __init__(self, field_name: str) -> None:
        super().__init__("FIELD_NOT_FOUND", f"field '{field_name}' not found")


class FieldAlreadyExistsError(DomainError):
    def __init__(self, field_name: str) -> None:
        super().__init__("FIELD_ALREADY_EXISTS", f"field '{field_name}' already exists")


class SystemFieldProtectedError(DomainError):
    def __init__(self, field_name: str) -> None:
        super().__init__(
            "SYSTEM_FIELD_PROTECTED",
            f"system field '{field_name}' cannot be deleted",
        )


class FieldInUseError(DomainError):
    def __init__(self, field_name: str) -> None:
        super().__init__(
            "FIELD_IN_USE",
            f"field '{field_name}' is enabled in one or more functions; use force=true",
        )


class ModuleNotFoundError(DomainError):
    def __init__(self, module: str) -> None:
        super().__init__("MODULE_NOT_FOUND", f"module '{module}' not found")


class FunctionNotFoundError(DomainError):
    def __init__(self, function: str) -> None:
        super().__init__("FUNCTION_NOT_FOUND", f"function '{function}' not found")


class RequiredFieldToggleError(DomainError):
    def __init__(self, field_name: str, function: str) -> None:
        super().__init__(
            "REQUIRED_FIELD_TOGGLE",
            f"field '{field_name}' cannot be disabled for function '{function}'",
        )


class InvalidSlugError(DomainError):
    def __init__(self, message: str) -> None:
        super().__init__("INVALID_SLUG", message)


class InvalidFieldNameError(DomainError):
    def __init__(self, message: str) -> None:
        super().__init__("INVALID_FIELD_NAME", message)


class InvalidEntityNameError(DomainError):
    def __init__(self, message: str) -> None:
        super().__init__("INVALID_ENTITY_NAME", message)


class InvalidProjectNameError(DomainError):
    def __init__(self, message: str) -> None:
        super().__init__("INVALID_PROJECT_NAME", message)