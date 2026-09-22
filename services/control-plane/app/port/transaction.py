from contextlib import AbstractAsyncContextManager
from typing import Any, Protocol


class TransactionManager(Protocol):
    def transaction(self) -> AbstractAsyncContextManager[Any]: ...
