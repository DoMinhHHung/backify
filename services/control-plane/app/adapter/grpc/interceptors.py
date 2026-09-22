import hmac
from collections.abc import Awaitable, Callable

import grpc

_METADATA_KEY = "x-internal-key"


class InternalKeyInterceptor(grpc.aio.ServerInterceptor):
    """Rejects any call that doesn't present the shared internal key.

    This is the gRPC equivalent of verify_internal_key in the HTTP layer:
    ControlPlaneService is only meant to be called by other backend
    services (e.g. Auth Service), never directly by end users, and it
    reuses the same secret so operators only manage one internal
    credential per environment.
    """

    def __init__(self, expected_key: str) -> None:
        self._expected_key = expected_key

    async def intercept_service(
        self,
        continuation: Callable[
            [grpc.HandlerCallDetails], Awaitable[grpc.RpcMethodHandler]
        ],
        handler_call_details: grpc.HandlerCallDetails,
    ) -> grpc.RpcMethodHandler:
        metadata = dict(handler_call_details.invocation_metadata or ())
        provided = metadata.get(_METADATA_KEY, "")
        if not hmac.compare_digest(provided, self._expected_key):

            async def deny(_request, context: grpc.aio.ServicerContext):
                await context.abort(
                    grpc.StatusCode.UNAUTHENTICATED, "invalid internal key"
                )

            return grpc.unary_unary_rpc_method_handler(deny)
        return await continuation(handler_call_details)
