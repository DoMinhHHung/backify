import structlog
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from app.domain.errors import DomainError

logger = structlog.get_logger()

STATUS_MAP: dict[str, int] = {
    "PROJECT_NOT_FOUND": 404,
    "PROJECT_SLUG_EXISTS": 409,
    "CONCURRENT_MODIFICATION": 409,
    "ENTITY_NOT_FOUND": 404,
    "ENTITY_ALREADY_EXISTS": 409,
    "FIELD_NOT_FOUND": 404,
    "FIELD_ALREADY_EXISTS": 409,
    "SYSTEM_FIELD_PROTECTED": 403,
    "FIELD_IN_USE": 409,
    "INVALID_FIELD_CONFIG": 400,
    "MODULE_NOT_FOUND": 404,
    "FUNCTION_NOT_FOUND": 404,
    "REQUIRED_FIELD_TOGGLE": 400,
    "INVALID_SLUG": 400,
    "INVALID_PROJECT_NAME": 400,
    "INVALID_FIELD_NAME": 400,
    "INVALID_ENTITY_NAME": 400,
    "INVALID_EMAIL": 400,
    "DEVELOPER_ALREADY_EXISTS": 409,
    "INVALID_CREDENTIALS": 401,
    "UNAUTHORIZED": 401,
    "FORBIDDEN": 403,
}


def register_error_handlers(app: FastAPI) -> None:
    @app.exception_handler(DomainError)
    async def domain_error_handler(_request: Request, exc: DomainError) -> JSONResponse:
        return JSONResponse(
            status_code=STATUS_MAP.get(exc.code, 400),
            content={"code": exc.code, "message": exc.message},
        )

    @app.exception_handler(ValueError)
    async def value_error_handler(_request: Request, exc: ValueError) -> JSONResponse:
        logger.exception("unhandled_value_error")
        return JSONResponse(
            status_code=500,
            content={
                "code": "INTERNAL_ERROR",
                "message": "internal server error",
            },
        )
