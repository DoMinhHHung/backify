from uuid import uuid4

import grpc
import pytest
from app.adapter.grpc.pb import control_plane_pb2


async def test_get_project_success(stub, existing_project, internal_key_metadata):
    response = await stub.GetProject(
        control_plane_pb2.GetProjectRequest(project_id=str(existing_project.id)),
        metadata=internal_key_metadata,
    )
    assert response.slug == "grpc-probe"
    assert response.schema_name == "proj_grpc_probe"
    assert response.owner_id == str(existing_project.owner_id)


async def test_get_project_config_success(stub, existing_project, internal_key_metadata):
    response = await stub.GetProjectConfig(
        control_plane_pb2.GetProjectConfigRequest(project_id=str(existing_project.id)),
        metadata=internal_key_metadata,
    )
    assert response.slug == "grpc-probe"
    assert '"entities"' in response.config_json
    assert '"modules"' in response.config_json


async def test_missing_internal_key_rejected(stub, existing_project):
    with pytest.raises(grpc.aio.AioRpcError) as exc_info:
        await stub.GetProject(
            control_plane_pb2.GetProjectRequest(project_id=str(existing_project.id))
        )
    assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED


async def test_wrong_internal_key_rejected(stub, existing_project):
    with pytest.raises(grpc.aio.AioRpcError) as exc_info:
        await stub.GetProject(
            control_plane_pb2.GetProjectRequest(project_id=str(existing_project.id)),
            metadata=(("x-internal-key", "wrong-key"),),
        )
    assert exc_info.value.code() == grpc.StatusCode.UNAUTHENTICATED


async def test_unknown_project_returns_not_found(stub, internal_key_metadata):
    with pytest.raises(grpc.aio.AioRpcError) as exc_info:
        await stub.GetProject(
            control_plane_pb2.GetProjectRequest(project_id=str(uuid4())),
            metadata=internal_key_metadata,
        )
    assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND


async def test_malformed_project_id_returns_invalid_argument(stub, internal_key_metadata):
    with pytest.raises(grpc.aio.AioRpcError) as exc_info:
        await stub.GetProject(
            control_plane_pb2.GetProjectRequest(project_id="not-a-uuid"),
            metadata=internal_key_metadata,
        )
    assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT


async def test_malformed_project_id_on_config_returns_invalid_argument(
    stub, internal_key_metadata
):
    with pytest.raises(grpc.aio.AioRpcError) as exc_info:
        await stub.GetProjectConfig(
            control_plane_pb2.GetProjectConfigRequest(project_id="garbage"),
            metadata=internal_key_metadata,
        )
    assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
