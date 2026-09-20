from typing import Annotated

from fastapi import APIRouter, Depends, status

from app.adapter.http.dependencies import get_login_developer, get_register_developer
from app.adapter.http.schemas import (
    DeveloperResponse,
    LoginRequest,
    RegisterRequest,
    TokenResponse,
)
from app.usecase.login_developer import LoginDeveloper, LoginDeveloperInput
from app.usecase.register_developer import RegisterDeveloper, RegisterDeveloperInput

router = APIRouter(prefix="/api/v1/auth", tags=["auth"])


@router.post(
    "/register",
    response_model=DeveloperResponse,
    status_code=status.HTTP_201_CREATED,
)
async def register(
    body: RegisterRequest,
    usecase: Annotated[RegisterDeveloper, Depends(get_register_developer)],
) -> DeveloperResponse:
    result = await usecase.execute(
        RegisterDeveloperInput(
            email=body.email,
            password=body.password,
            name=body.name,
        )
    )
    return DeveloperResponse(
        id=result.developer.id,
        email=result.developer.email,
        name=result.developer.name,
    )


@router.post("/login", response_model=TokenResponse)
async def login(
    body: LoginRequest,
    usecase: Annotated[LoginDeveloper, Depends(get_login_developer)],
) -> TokenResponse:
    result = await usecase.execute(
        LoginDeveloperInput(email=body.email, password=body.password)
    )
    return TokenResponse(
        access_token=result.access_token,
        token_type=result.token_type,
        developer_id=result.developer_id,
        email=result.email,
    )