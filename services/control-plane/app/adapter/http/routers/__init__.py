from app.adapter.http.routers.auth import router as auth_router
from app.adapter.http.routers.projects import router as projects_router

__all__ = ["auth_router", "projects_router"]