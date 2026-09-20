from app.adapter.postgres.connection import Database
from app.adapter.postgres.project_repository import PostgresProjectRepository
from app.config import Settings
from app.usecase.create_project import CreateProject
from app.usecase.get_project import GetProject


class Container:
    def __init__(self, settings: Settings, database: Database) -> None:
        self._settings = settings
        self._database = database
        self._project_repository = PostgresProjectRepository(database)
        self.create_project = CreateProject(self._project_repository)
        self.get_project = GetProject(self._project_repository)

    @property
    def database(self) -> Database:
        return self._database

    @property
    def settings(self) -> Settings:
        return self._settings