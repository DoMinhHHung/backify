package usecase

import (
	"context"
	"testing"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

func TestDeleteField_SystemFieldBlocked(t *testing.T) {
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}
	field, _ := domain.NewField("entity1", "id", domain.FieldTypeUUID)
	_ = fields.Create(context.Background(), field)

	uc := NewDeleteField(fields, modules, publisher)
	err := uc.Execute(context.Background(), DeleteFieldInput{FieldID: field.ID})
	if err != domain.ErrSystemFieldCannotDelete {
		t.Fatalf("expected ErrSystemFieldCannotDelete, got %v", err)
	}
}

func TestDeleteField_InUseWithoutForce(t *testing.T) {
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}
	field, _ := domain.NewField("entity1", "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)
	modules.usages[field.ID] = []port.FieldUsage{{ModuleName: "auth", FunctionName: "signup"}}

	uc := NewDeleteField(fields, modules, publisher)
	err := uc.Execute(context.Background(), DeleteFieldInput{FieldID: field.ID})

	domainErr, ok := err.(*domain.Error)
	if !ok || domainErr.Code != domain.CodeFieldInUse {
		t.Fatalf("expected CodeFieldInUse, got %v", err)
	}
}

func TestDeleteField_ForceDeletesAndDisables(t *testing.T) {
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}
	field, _ := domain.NewField("entity1", "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)
	modules.usages[field.ID] = []port.FieldUsage{{ModuleName: "auth", FunctionName: "signup"}}

	uc := NewDeleteField(fields, modules, publisher)
	if err := uc.Execute(context.Background(), DeleteFieldInput{FieldID: field.ID, Force: true}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(publisher.events) != 1 || publisher.events[0].Name != "field.deleted" {
		t.Fatalf("expected field.deleted event, got %v", publisher.events)
	}
}

func TestDeleteField_CustomNotInUse(t *testing.T) {
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}
	field, _ := domain.NewField("entity1", "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)

	uc := NewDeleteField(fields, modules, publisher)
	if err := uc.Execute(context.Background(), DeleteFieldInput{FieldID: field.ID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
