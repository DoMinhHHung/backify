import re
from typing import Self
from uuid import UUID, uuid4

from app.domain.entity import Entity
from app.domain.errors import InvalidProjectNameError, InvalidSlugError
from app.domain.module import ModuleConfig, ModuleName


_SLUG_PATTERN = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
_SLUG_MIN = 3
_SLUG_MAX = 40


class Project:
    def __init__(
        self,
        name: str,
        slug: str,
        *,
        project_id: UUID | None = None,
        entities: dict[str, Entity] | None = None,
        modules: dict[ModuleName, ModuleConfig] | None = None,
    ) -> None:
        self._validate_slug(slug)
        if not name or not name.strip():
            raise InvalidProjectNameError("project name must not be empty")
        self.id = project_id or uuid4()
        self.name = name.strip()
        self.slug = slug
        self.schema_name = f"proj_{slug.replace('-', '_')}"
        self._entities: dict[str, Entity] = entities or {}
        self._modules: dict[ModuleName, ModuleConfig] = modules or {}

    @staticmethod
    def _validate_slug(slug: str) -> None:
        if len(slug) < _SLUG_MIN or len(slug) > _SLUG_MAX:
            raise InvalidSlugError(
                f"slug must be {_SLUG_MIN}-{_SLUG_MAX} characters, got {len(slug)}"
            )
        if not _SLUG_PATTERN.match(slug):
            raise InvalidSlugError(
                f"invalid slug '{slug}': only [a-z0-9-], no leading/trailing/double hyphen"
            )

    @property
    def entities(self) -> dict[str, Entity]:
        return dict(self._entities)

    @property
    def modules(self) -> dict[ModuleName, ModuleConfig]:
        return dict(self._modules)

    def get_entity(self, name: str) -> Entity | None:
        return self._entities.get(name)

    def add_entity(self, entity: Entity) -> None:
        if entity.name in self._entities:
            raise ValueError(f"entity '{entity.name}' already exists")
        self._entities[entity.name] = entity

    def get_module(self, module: ModuleName) -> ModuleConfig | None:
        return self._modules.get(module)

    def enable_module(self, config: ModuleConfig) -> None:
        config.enabled = True
        self._modules[config.module] = config

    def to_dict(self) -> dict[str, object]:
        return {
            "id": str(self.id),
            "name": self.name,
            "slug": self.slug,
            "schema": self.schema_name,
            "entities": {name: entity.to_dict() for name, entity in self._entities.items()},
            "modules": {
                module.value: config.to_dict()
                for module, config in self._modules.items()
            },
        }

    def to_config_dict(self) -> dict[str, object]:
        return {
            "entities": {name: entity.to_dict() for name, entity in self._entities.items()},
            "modules": {
                module.value: config.to_dict()
                for module, config in self._modules.items()
            },
        }

    @classmethod
    def from_persistence(
        cls,
        *,
        project_id: UUID,
        name: str,
        slug: str,
        config: dict[str, object],
    ) -> Self:
        entities: dict[str, Entity] = {}
        raw_entities = config.get("entities", {})
        if isinstance(raw_entities, dict):
            for key, value in raw_entities.items():
                if isinstance(value, dict):
                    entities[str(key)] = Entity.from_dict(value)

        modules: dict[ModuleName, ModuleConfig] = {}
        raw_modules = config.get("modules", {})
        if isinstance(raw_modules, dict):
            for key, value in raw_modules.items():
                try:
                    module_name = ModuleName(str(key))
                except ValueError:
                    continue
                if isinstance(value, dict):
                    modules[module_name] = ModuleConfig.from_dict(module_name, value)

        return cls(
            name=name,
            slug=slug,
            project_id=project_id,
            entities=entities,
            modules=modules,
        )

    @classmethod
    def create(cls, name: str, slug: str) -> Self:
        project = cls(name=name, slug=slug)
        project.add_entity(Entity.user_with_default_pool())
        project.enable_module(ModuleConfig.default_auth())
        return project