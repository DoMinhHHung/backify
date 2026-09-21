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


class ConcurrentModificationError(DomainError):
    def __init__(self, project_id: str) -> None:
        super().__init__(
            "CONCURRENT_MODIFICATION",
            f"project '{project_id}' was modified by another request, retry",
        )


class EntityNotFoundError(DomainError):
    def __init__(self, entity_name: str) -> None:
        super().__init__("ENTITY_NOT_FOUND", f"entity '{entity_name}' not found")


class EntityAlreadyExistsError(DomainError):
    def __init__(self, entity_name: str) -> None:
        super().__init__(
            "ENTITY_ALREADY_EXISTS",
            f"entity '{entity_name}' already exists",
        )


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


class InvalidFieldConfigError(DomainError):
    def __init__(self, message: str) -> None:
        super().__init__("INVALID_FIELD_CONFIG", message)


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


class DeveloperAlreadyExistsError(DomainError):
    def __init__(self, email: str) -> None:
        super().__init__(
            "DEVELOPER_ALREADY_EXISTS",
            f"developer with email '{email}' already exists",
        )


class InvalidCredentialsError(DomainError):
    def __init__(self) -> None:
        super().__init__("INVALID_CREDENTIALS", "invalid email or password")


class UnauthorizedError(DomainError):
    def __init__(self, message: str = "unauthorized") -> None:
        super().__init__("UNAUTHORIZED", message)


class ForbiddenError(DomainError):
    def __init__(self, message: str = "forbidden") -> None:
        super().__init__("FORBIDDEN", message)
