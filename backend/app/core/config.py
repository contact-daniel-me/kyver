from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    app_name: str = "Kyver API"
    app_env: str = "development"
    app_host: str = "0.0.0.0"
    app_port: int = 8000

    frontend_url: str = "http://localhost:3000"
    database_url: str = "postgresql+asyncpg://localhost:5432/kyver"

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )


@lru_cache
def get_settings() -> Settings:
    return Settings()
