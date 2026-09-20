from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from app.domain.errors import DomainError


def register_error_handlers(app: FastAPI) -> None:
    @app.exception_handler(DomainError)
    async def domain_error_handler(_request: Request, exc: DomainError) -> JSONResponse:
        status_map = {
            "PROJECT_NOT_FOUND": 404,
            "PROJECT_SLUG_EXISTS": 409,
            "INVALID_SLUG": 400,
            "INVALID_PROJECT_NAME": 400,
            "INVALID_FIELD_NAME": 400,
            "INVALID_ENTITY_NAME": 400,
        }
        status_code = status_map.get(exc.code, 400)
        return JSONResponse(
            status_code=status_code,
            content={"code": exc.code, "message": exc.message},
        )