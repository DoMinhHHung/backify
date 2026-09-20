from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


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
    database_pool_min_size: int = 1
    database_pool_max_size: int = 10

    log_level: str = "INFO"
    log_json: bool = True

    rabbitmq_url: str = "amqp://backify:backify@localhost:5672/"


@lru_cache
def get_settings() -> Settings:
    return Settings()