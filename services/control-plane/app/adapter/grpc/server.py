import json
from uuid import UUID

import grpc
import structlog

from app.adapter.grpc.interceptors import InternalKeyInterceptor
from app.adapter.grpc.pb import control_plane_pb2, control_plane_pb2_grpc
from app.config import Settings
from app.container import Container
from app.domain.errors import ProjectNotFoundError
from app.usecase.get_project_config import GetProjectConfigInput
from app.usecase.get_project_internal import GetProjectInternalInput

logger = structlog.get_logger()


async def _parse_project_id(raw: str, context: grpc.aio.ServicerContext) -> UUID:
    try:
        return UUID(raw)
    except ValueError:
        await context.abort(
            grpc.StatusCode.INVALID_ARGUMENT, f"invalid project_id '{raw}'"
        )
        raise AssertionError("unreachable") from None


class ControlPlaneServicer(control_plane_pb2_grpc.ControlPlaneServiceServicer):
    def __init__(self, container: Container) -> None:
        self._container = container

    async def GetProject(
        self,
        request: control_plane_pb2.GetProjectRequest,
        context: grpc.aio.ServicerContext,
    ) -> control_plane_pb2.GetProjectResponse:
        project_id = await _parse_project_id(request.project_id, context)
        try:
            result = await self._container.get_project_internal.execute(
                GetProjectInternalInput(project_id=project_id)
            )
        except ProjectNotFoundError as exc:
            await context.abort(grpc.StatusCode.NOT_FOUND, exc.message)
            raise
        return control_plane_pb2.GetProjectResponse(
            id=str(result.id),
            name=result.name,
            slug=result.slug,
            schema_name=result.schema_name,
            owner_id=str(result.owner_id) if result.owner_id else "",
            version=result.version,
        )

    async def GetProjectConfig(
        self,
        request: control_plane_pb2.GetProjectConfigRequest,
        context: grpc.aio.ServicerContext,
    ) -> control_plane_pb2.GetProjectConfigResponse:
        project_id = await _parse_project_id(request.project_id, context)
        try:
            result = await self._container.get_project_config.execute(
                GetProjectConfigInput(project_id=project_id)
            )
        except ProjectNotFoundError as exc:
            await context.abort(grpc.StatusCode.NOT_FOUND, exc.message)
            raise
        return control_plane_pb2.GetProjectConfigResponse(
            project_id=str(result.project_id),
            slug=result.slug,
            schema_name=result.schema_name,
            version=result.version,
            config_json=json.dumps(result.config),
        )


def create_grpc_server(container: Container, settings: Settings) -> grpc.aio.Server:
    server = grpc.aio.server(
        interceptors=[InternalKeyInterceptor(settings.internal_api_key)]
    )
    control_plane_pb2_grpc.add_ControlPlaneServiceServicer_to_server(
        ControlPlaneServicer(container), server
    )
    server.add_insecure_port(f"{settings.app_host}:{settings.grpc_port}")
    return server
