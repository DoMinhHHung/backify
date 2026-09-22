import asyncio
import base64
import hashlib

import bcrypt

_DUMMY_HASH = bcrypt.hashpw(b"backify-dummy-password", bcrypt.gensalt()).decode("utf-8")


def _prehash(plain: str) -> bytes:
    return base64.b64encode(hashlib.sha256(plain.encode("utf-8")).digest())


def _hash_sync(plain: str) -> str:
    return bcrypt.hashpw(_prehash(plain), bcrypt.gensalt()).decode("utf-8")


def _verify_sync(plain: str, hashed: str) -> bool:
    try:
        return bcrypt.checkpw(_prehash(plain), hashed.encode("utf-8"))
    except ValueError:
        return False


async def hash_password(plain: str) -> str:
    return await asyncio.to_thread(_hash_sync, plain)


async def verify_password(plain: str, hashed: str) -> bool:
    return await asyncio.to_thread(_verify_sync, plain, hashed)


def dummy_hash() -> str:
    return _DUMMY_HASH
