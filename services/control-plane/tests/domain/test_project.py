from uuid import uuid4

import pytest
from app.domain.entity import Entity
from app.domain.errors import (
    EntityAlreadyExistsError,
    EntityNotFoundError,
    FieldInUseError,
    FieldNotFoundError,
    InvalidEntityNameError,
    InvalidFieldConfigError,
    InvalidProjectNameError,
    InvalidSlugError,
    RequiredFieldToggleError,
    SystemFieldProtectedError,
)
from app.domain.field import Field, FieldType, RelationCardinality
from app.domain.module import FunctionName, ModuleName
from app.domain.project import Project


def make_project(name: str = "Shop", slug: str = "my-shop") -> Project:
    return Project.create(name, slug, owner_id=uuid4())


def test_create_project_success():
    project = make_project("My Shop", "my-shop")
    assert project.schema_name == "proj_my_shop"
    assert project.version == 1
    assert "User" in project.entities
    assert project.modules[ModuleName.AUTH].enabled is True


@pytest.mark.parametrize("slug", ["ab", "My_Shop", "-lead", "trail-", "dou--ble"])
def test_invalid_slug(slug):
    with pytest.raises(InvalidSlugError):
        make_project("Test", slug)


def test_empty_name():
    with pytest.raises(InvalidProjectNameError):
        make_project("   ", "valid-slug")


def test_add_entity_duplicate():
    project = make_project()
    with pytest.raises(EntityAlreadyExistsError):
        project.add_entity(Entity("User"))


def test_entity_name_must_be_ascii_identifier():
    with pytest.raises(InvalidEntityNameError):
        Entity("NgườiDùng")


def test_add_and_remove_custom_field():
    project = make_project()
    project.add_field_to_entity("User", Field("address", FieldType.STRING))
    assert project.require_entity("User").has_field("address")
    project.remove_field_from_entity("User", "address")
    assert not project.require_entity("User").has_field("address")


def test_cannot_remove_system_field():
    project = make_project()
    with pytest.raises(SystemFieldProtectedError):
        project.remove_field_from_entity("User", "email")


def test_remove_field_in_use_without_force():
    project = make_project()
    project.add_field_to_entity("User", Field("address", FieldType.STRING))
    project.set_function_fields(
        ModuleName.AUTH, FunctionName.SIGNUP, ["email", "password", "address"]
    )
    with pytest.raises(FieldInUseError):
        project.remove_field_from_entity("User", "address")


def test_remove_field_in_use_with_force_cleans_config():
    project = make_project()
    project.add_field_to_entity("User", Field("address", FieldType.STRING))
    project.set_function_fields(
        ModuleName.AUTH, FunctionName.SIGNUP, ["email", "password", "address"]
    )
    project.remove_field_from_entity("User", "address", force=True)
    signup = project.require_module(ModuleName.AUTH).get_function(FunctionName.SIGNUP)
    assert not signup.is_enabled("address")


def test_system_field_protection_beats_in_use_check():
    project = make_project()
    with pytest.raises(SystemFieldProtectedError):
        project.remove_field_from_entity("User", "fullName", force=True)


def test_same_field_name_on_other_entity_is_not_in_use():
    project = make_project()
    project.add_entity(Entity("Product"))
    project.add_field_to_entity("Product", Field("fullName", FieldType.STRING))
    project.remove_field_from_entity("Product", "fullName")
    assert not project.require_entity("Product").has_field("fullName")
    assert project.require_entity("User").has_field("fullName")


def test_crud_entity_function_field_is_detected_as_in_use():
    project = make_project()
    project.add_entity(Entity("Product"))
    project.add_field_to_entity("Product", Field("title", FieldType.STRING))
    project.enable_module_by_name(ModuleName.CRUD)
    project.set_function_fields(
        ModuleName.CRUD, FunctionName.CREATE, ["title"], entity_name="Product"
    )
    with pytest.raises(FieldInUseError):
        project.remove_field_from_entity("Product", "title")


def test_force_remove_cleans_crud_entity_function():
    project = make_project()
    project.add_entity(Entity("Product"))
    project.add_field_to_entity("Product", Field("title", FieldType.STRING))
    project.enable_module_by_name(ModuleName.CRUD)
    project.set_function_fields(
        ModuleName.CRUD, FunctionName.CREATE, ["title"], entity_name="Product"
    )
    project.remove_field_from_entity("Product", "title", force=True)
    crud = project.to_config_dict()["modules"]["crud"]
    assert crud["entityFunctions"]["Product"]["create"]["enabledFields"] == []


def test_enable_module_twice_preserves_toggles():
    project = make_project()
    project.set_function_fields(
        ModuleName.AUTH, FunctionName.SIGNUP, ["email", "password"]
    )
    project.enable_module_by_name(ModuleName.AUTH)
    signup = project.require_module(ModuleName.AUTH).get_function(FunctionName.SIGNUP)
    assert signup.enabled_fields == ["email", "password"]


def test_set_function_fields_success():
    project = make_project()
    project.add_field_to_entity("User", Field("address", FieldType.STRING))
    project.set_function_fields(
        ModuleName.AUTH,
        FunctionName.SIGNUP,
        ["email", "password", "fullName", "address"],
    )
    signup = project.require_module(ModuleName.AUTH).get_function(FunctionName.SIGNUP)
    assert "address" in signup.enabled_fields


def test_set_function_fields_missing_required():
    project = make_project()
    with pytest.raises(RequiredFieldToggleError):
        project.set_function_fields(
            ModuleName.AUTH, FunctionName.SIGNUP, ["email", "fullName"]
        )


def test_set_function_fields_unknown_field():
    project = make_project()
    with pytest.raises(FieldNotFoundError):
        project.set_function_fields(
            ModuleName.AUTH, FunctionName.SIGNUP, ["email", "password", "nope"]
        )


def test_crud_requires_entity_name():
    project = make_project()
    project.enable_module_by_name(ModuleName.CRUD)
    with pytest.raises(InvalidFieldConfigError):
        project.set_function_fields(ModuleName.CRUD, FunctionName.CREATE, [])


def test_relation_field_requires_existing_target():
    project = make_project()
    field = Field(
        "owner",
        FieldType.RELATION,
        relation_to="Ghost",
        relation_cardinality=RelationCardinality.MANY_TO_ONE,
    )
    with pytest.raises(EntityNotFoundError):
        project.add_field_to_entity("User", field)


def test_roundtrip_through_persistence():
    project = make_project()
    project.add_entity(Entity("Product"))
    project.add_field_to_entity("Product", Field("title", FieldType.STRING))
    project.enable_module_by_name(ModuleName.CRUD)
    project.set_function_fields(
        ModuleName.CRUD, FunctionName.CREATE, ["title"], entity_name="Product"
    )
    restored = Project.from_persistence(
        project_id=project.id,
        name=project.name,
        slug=project.slug,
        config=project.to_config_dict(),
        owner_id=project.owner_id,
        version=project.version,
    )
    assert restored.to_config_dict() == project.to_config_dict()
