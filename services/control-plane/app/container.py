from app.adapter.postgres.connection import Database
from app.adapter.postgres.developer_repository import PostgresDeveloperRepository
from app.adapter.postgres.outbox_repository import PostgresOutboxRepository
from app.adapter.postgres.project_repository import PostgresProjectRepository
from app.config import Settings
from app.port.developer_repository import DeveloperRepository
from app.port.outbox_repository import OutboxRepository
from app.port.project_repository import ProjectRepository
from app.port.transaction import TransactionManager
from app.usecase.add_entity import AddEntity
from app.usecase.add_field import AddField
from app.usecase.create_project import CreateProject
from app.usecase.delete_project import DeleteProject
from app.usecase.enable_module import EnableModule
from app.usecase.get_project import GetProject
from app.usecase.get_project_config import GetProjectConfig
from app.usecase.get_project_internal import GetProjectInternal
from app.usecase.list_projects import ListProjects
from app.usecase.login_developer import LoginDeveloper
from app.usecase.register_developer import RegisterDeveloper
from app.usecase.remove_field import RemoveField
from app.usecase.set_function_fields import SetFunctionFields


class Container:
    def __init__(
        self,
        settings: Settings,
        *,
        projects: ProjectRepository,
        developers: DeveloperRepository,
        outbox: OutboxRepository,
        tx: TransactionManager,
    ) -> None:
        self._settings = settings
        self._projects = projects
        self._developers = developers
        self._outbox = outbox
        self._tx = tx
        self._wire_usecases()

    @classmethod
    def from_database(cls, settings: Settings, database: Database) -> "Container":
        return cls(
            settings,
            projects=PostgresProjectRepository(database),
            developers=PostgresDeveloperRepository(database),
            outbox=PostgresOutboxRepository(database),
            tx=database,
        )

    def _wire_usecases(self) -> None:
        mutation_args = (self._projects, self._outbox, self._tx)
        self.create_project = CreateProject(*mutation_args)
        self.delete_project = DeleteProject(*mutation_args)
        self.add_entity = AddEntity(*mutation_args)
        self.add_field = AddField(*mutation_args)
        self.remove_field = RemoveField(*mutation_args)
        self.set_function_fields = SetFunctionFields(*mutation_args)
        self.enable_module = EnableModule(*mutation_args)
        self.get_project = GetProject(self._projects)
        self.get_project_config = GetProjectConfig(self._projects)
        self.get_project_internal = GetProjectInternal(self._projects)
        self.list_projects = ListProjects(self._projects)
        self.register_developer = RegisterDeveloper(self._developers)
        self.login_developer = LoginDeveloper(self._developers, self._settings)

    @property
    def settings(self) -> Settings:
        return self._settings
