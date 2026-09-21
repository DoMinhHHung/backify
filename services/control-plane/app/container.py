from app.adapter.messaging.noop_publisher import NoopEventPublisher
from app.adapter.messaging.rabbitmq_publisher import RabbitMQEventPublisher
from app.adapter.postgres.connection import Database
from app.adapter.postgres.developer_repository import PostgresDeveloperRepository
from app.adapter.postgres.project_repository import PostgresProjectRepository
from app.config import Settings
from app.port.event_publisher import EventPublisher
from app.usecase.add_entity import AddEntity
from app.usecase.add_field import AddField
from app.usecase.create_project import CreateProject
from app.usecase.delete_project import DeleteProject
from app.usecase.get_project import GetProject
from app.usecase.list_projects import ListProjects
from app.usecase.login_developer import LoginDeveloper
from app.usecase.register_developer import RegisterDeveloper
from app.usecase.remove_field import RemoveField
from app.usecase.set_function_fields import SetFunctionFields
from app.usecase.enable_module import EnableModule


class Container:
    def __init__(
        self,
        settings: Settings,
        database: Database,
        event_publisher: EventPublisher,
    ) -> None:
        self._settings = settings
        self._database = database
        self._event_publisher = event_publisher
        self._project_repository = PostgresProjectRepository(database)
        self._developer_repository = PostgresDeveloperRepository(database)

        self.create_project = CreateProject(
            self._project_repository, self._event_publisher
        )
        self.get_project = GetProject(self._project_repository)
        self.list_projects = ListProjects(self._project_repository)
        self.delete_project = DeleteProject(
            self._project_repository, self._event_publisher
        )
        self.add_entity = AddEntity(self._project_repository, self._event_publisher)
        self.add_field = AddField(self._project_repository, self._event_publisher)
        self.remove_field = RemoveField(
            self._project_repository, self._event_publisher
        )
        self.set_function_fields = SetFunctionFields(
            self._project_repository, self._event_publisher
        )
        self.register_developer = RegisterDeveloper(self._developer_repository)
        self.login_developer = LoginDeveloper(
            self._developer_repository, settings
        )
        self.enable_module = EnableModule(
            self._project_repository, self._event_publisher
        )

    @property
    def database(self) -> Database:
        return self._database

    @property
    def settings(self) -> Settings:
        return self._settings

    @property
    def event_publisher(self) -> EventPublisher:
        return self._event_publisher