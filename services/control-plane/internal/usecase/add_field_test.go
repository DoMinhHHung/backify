package usecase

import (
	"context"
	"testing"

	"backify/services/control-plane/internal/domain"
)

func TestAddField_Success(t *testing.T) {
	entities := newMockEntityRepo()
	fields := newMockFieldRepo()
	entity, _ := domain.NewEntity("proj1", "product")
	_ = entities.Create(context.Background(), entity)

	uc := NewAddField(entities, fields)
	field, err := uc.Execute(context.Background(), AddFieldInput{EntityID: entity.ID, Name: "title", Type: domain.FieldTypeString})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if field.IsSystem {
		t.Fatal("expected non-system field")
	}
}

func TestAddField_EntityNotFound(t *testing.T) {
	uc := NewAddField(newMockEntityRepo(), newMockFieldRepo())
	_, err := uc.Execute(context.Background(), AddFieldInput{EntityID: "missing", Name: "title", Type: domain.FieldTypeString})
	if err != domain.ErrEntityNotFound {
		t.Fatalf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestAddField_InvalidType(t *testing.T) {
	entities := newMockEntityRepo()
	fields := newMockFieldRepo()
	entity, _ := domain.NewEntity("proj1", "product")
	_ = entities.Create(context.Background(), entity)

	uc := NewAddField(entities, fields)
	_, err := uc.Execute(context.Background(), AddFieldInput{EntityID: entity.ID, Name: "title", Type: domain.FieldType("unknown")})
	if err == nil {
		t.Fatal("expected error for invalid field type")
	}
}

func TestAddField_List(t *testing.T) {
	entities := newMockEntityRepo()
	fields := newMockFieldRepo()
	entity, _ := domain.NewEntity("proj1", "product")
	_ = entities.Create(context.Background(), entity)

	uc := NewAddField(entities, fields)
	ctx := context.Background()
	_, _ = uc.Execute(ctx, AddFieldInput{EntityID: entity.ID, Name: "title", Type: domain.FieldTypeString})
	_, _ = uc.Execute(ctx, AddFieldInput{EntityID: entity.ID, Name: "price", Type: domain.FieldTypeNumber})

	list, err := uc.List(ctx, entity.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(list))
	}
}
