from functools import lru_cache

from pydantic import model_validator
from pydantic_settings import BaseSettings, SettingsConfigDict

INSECURE_DEFAULTS = frozenset({"change-me-in-production", "dev-internal-key", ""})
MIN_SECRET_LENGTH = 32


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    app_name: str = "control-plane"
    app_env: str = "development"
    app_debug: bool = False
    app_host: str = "0.0.0.0"
    app_port: int = 8000

    database_url: str = "postgresql://backify:backify@localhost:5432/backify_control"
    database_pool_min_size: int = 2
    database_pool_max_size: int = 20
    database_statement_cache_size: int = 256
    database_command_timeout: float = 10.0

    log_level: str = "INFO"
    log_json: bool = True

    rabbitmq_url: str = "amqp://backify:backify@localhost:5672/"
    rabbitmq_enabled: bool = False
    outbox_poll_interval: float = 1.0
    outbox_batch_size: int = 50

    jwt_secret: str = "change-me-in-production"
    jwt_algorithm: str = "HS256"
    jwt_expire_minutes: int = 60

    internal_api_key: str = "dev-internal-key"

    cors_origins: str = "http://localhost:3000,http://localhost:5173"

    @property
    def is_production(self) -> bool:
        return self.app_env.lower() in {"production", "prod"}

    @property
    def cors_origin_list(self) -> list[str]:
        return [o.strip() for o in self.cors_origins.split(",") if o.strip()]

    @model_validator(mode="after")
    def _enforce_production_hardening(self) -> "Settings":
        if not self.is_production:
            return self
        if self.jwt_secret in INSECURE_DEFAULTS:
            raise ValueError("JWT_SECRET must be set explicitly in production")
        if len(self.jwt_secret) < MIN_SECRET_LENGTH:
            raise ValueError(
                f"JWT_SECRET must be at least {MIN_SECRET_LENGTH} characters"
            )
        if self.internal_api_key in INSECURE_DEFAULTS:
            raise ValueError("INTERNAL_API_KEY must be set explicitly in production")
        if len(self.internal_api_key) < MIN_SECRET_LENGTH:
            raise ValueError(
                f"INTERNAL_API_KEY must be at least {MIN_SECRET_LENGTH} characters"
            )
        if not self.rabbitmq_enabled:
            raise ValueError("RABBITMQ_ENABLED must be true in production")
        if self.app_debug:
            raise ValueError("APP_DEBUG must be false in production")
        if "*" in self.cors_origin_list:
            raise ValueError("CORS_ORIGINS must not be '*' in production")
        return self


@lru_cache
def get_settings() -> Settings:
    return Settings()
