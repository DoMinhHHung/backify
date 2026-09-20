import pytest

from app.domain.errors import (
    EntityAlreadyExistsError,
    FieldInUseError,
    FieldNotFoundError,
    InvalidProjectNameError,
    InvalidSlugError,
    RequiredFieldToggleError,
    SystemFieldProtectedError,
)
from app.domain.field import Field, FieldType
from app.domain.module import FunctionName, ModuleName
from app.domain.project import Project
from uuid import uuid4


def test_create_project_success():
    project = Project.create("My Shop", "my-shop", owner_id=uuid4())
    assert project.name == "My Shop"
    assert project.slug == "my-shop"
    assert project.schema_name == "proj_my_shop"
    assert "User" in project.entities
    assert ModuleName.AUTH in project.modules
    assert project.modules[ModuleName.AUTH].enabled is True


def test_invalid_slug_too_short():
    with pytest.raises(InvalidSlugError):
        Project.create("Test", "ab")


def test_invalid_slug_pattern():
    with pytest.raises(InvalidSlugError):
        Project.create("Test", "My_Shop")


def test_empty_name():
    with pytest.raises(InvalidProjectNameError):
        Project.create("   ", "valid-slug")


def test_add_entity_duplicate():
    project = Project.create("Shop", "shop-a")
    with pytest.raises(EntityAlreadyExistsError):
        from app.domain.entity import Entity

        project.add_entity(Entity("User"))


def test_add_and_remove_custom_field():
    project = Project.create("Shop", "shop-b")
    field = Field("address", FieldType.STRING)
    project.add_field_to_entity("User", field)
    assert project.require_entity("User").has_field("address")
    project.remove_field_from_entity("User", "address")
    assert not project.require_entity("User").has_field("address")


def test_cannot_remove_system_field():
    project = Project.create("Shop", "shop-c")
    with pytest.raises(SystemFieldProtectedError):
        project.remove_field_from_entity("User", "email")


def test_remove_field_in_use_without_force():
    project = Project.create("Shop", "shop-d")
    with pytest.raises(FieldInUseError):
        project.remove_field_from_entity("User", "fullName")


def test_remove_field_in_use_with_force():
    project = Project.create("Shop", "shop-e")
    project.remove_field_from_entity("User", "fullName", force=True)
    assert not project.require_entity("User").has_field("fullName")
    signup = project.require_module(ModuleName.AUTH).get_function(FunctionName.SIGNUP)
    assert signup is not None
    assert not signup.is_enabled("fullName")


def test_set_function_fields_success():
    project = Project.create("Shop", "shop-f")
    project.add_field_to_entity("User", Field("address", FieldType.STRING))
    project.set_function_fields(
        ModuleName.AUTH,
        FunctionName.SIGNUP,
        ["email", "password", "fullName", "address"],
    )
    signup = project.require_module(ModuleName.AUTH).get_function(FunctionName.SIGNUP)
    assert signup is not None
    assert "address" in signup.enabled_fields


def test_set_function_fields_missing_required():
    project = Project.create("Shop", "shop-g")
    with pytest.raises(RequiredFieldToggleError):
        project.set_function_fields(
            ModuleName.AUTH,
            FunctionName.SIGNUP,
            ["fullName"],
        )


def test_set_function_fields_unknown_field():
    project = Project.create("Shop", "shop-h")
    with pytest.raises(FieldNotFoundError):
        project.set_function_fields(
            ModuleName.AUTH,
            FunctionName.SIGNUP,
            ["email", "password", "unknownField"],
        )