import pytest
from app.domain.errors import InvalidFieldConfigError, InvalidFieldNameError
from app.domain.field import Field, FieldType, RelationCardinality


def test_enum_requires_values():
    with pytest.raises(InvalidFieldConfigError):
        Field("role", FieldType.ENUM)


def test_enum_values_rejected_on_other_types():
    with pytest.raises(InvalidFieldConfigError):
        Field("age", FieldType.INT, enum_values=["a"])


def test_relation_requires_cardinality():
    with pytest.raises(InvalidFieldConfigError):
        Field("owner", FieldType.RELATION, relation_to="User")


def test_relation_requires_target():
    with pytest.raises(InvalidFieldConfigError):
        Field(
            "owner",
            FieldType.RELATION,
            relation_cardinality=RelationCardinality.MANY_TO_ONE,
        )


def test_bad_cardinality_is_domain_error():
    with pytest.raises(InvalidFieldConfigError):
        RelationCardinality.parse("n-n")


@pytest.mark.parametrize("name", ["", "1abc", "has-dash", "có dấu"])
def test_invalid_field_names(name):
    with pytest.raises(InvalidFieldNameError):
        Field(name, FieldType.STRING)
