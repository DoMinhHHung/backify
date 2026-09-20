import pytest

from app.domain.errors import InvalidProjectNameError, InvalidSlugError
from app.domain.module import ModuleName
from app.domain.project import Project


def test_create_project_success():
    project = Project.create("My Shop", "my-shop")
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