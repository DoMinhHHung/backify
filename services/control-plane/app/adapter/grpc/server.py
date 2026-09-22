import json
from pathlib import Path
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


def _load_server_credentials(settings: Settings) -> grpc.ServerCredentials:
    cert_path = Path(settings.grpc_tls_cert_file or "")
    key_path = Path(settings.grpc_tls_key_file or "")
    private_key = key_path.read_bytes()
    certificate_chain = cert_path.read_bytes()
    if settings.grpc_tls_client_ca_file:
        root_certificates = Path(settings.grpc_tls_client_ca_file).read_bytes()
        return grpc.ssl_server_credentials(
            [(private_key, certificate_chain)],
            root_certificates=root_certificates,
            require_client_auth=True,
        )
    return grpc.ssl_server_credentials([(private_key, certificate_chain)])


def create_grpc_server(container: Container, settings: Settings) -> grpc.aio.Server:
    server = grpc.aio.server(
        interceptors=[InternalKeyInterceptor(settings.internal_api_key)]
    )
    control_plane_pb2_grpc.add_ControlPlaneServiceServicer_to_server(
        ControlPlaneServicer(container), server
    )
    bind_addr = f"{settings.app_host}:{settings.grpc_port}"
    if settings.grpc_tls_enabled:
        credentials = _load_server_credentials(settings)
        bound = server.add_secure_port(bind_addr, credentials)
        logger.info(
            "grpc_secure_port_bound",
            address=bind_addr,
            port=bound,
            mutual_tls=bool(settings.grpc_tls_client_ca_file),
        )
    else:
        bound = server.add_insecure_port(bind_addr)
        logger.warning(
            "grpc_insecure_port_bound",
            address=bind_addr,
            port=bound,
            hint="set GRPC_TLS_CERT_FILE and GRPC_TLS_KEY_FILE for TLS",
        )
    return server
