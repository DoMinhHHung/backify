package usecase

import (
	"context"
	"testing"

	"backify/services/control-plane/internal/domain"
)

func TestConfigModule_ToggleField_Success(t *testing.T) {
	projects := newMockProjectRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)
	field, _ := domain.NewField("entity1", "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)

	uc := NewConfigModule(projects, fields, modules, publisher)
	err := uc.ToggleField(context.Background(), ToggleFieldInput{
		ProjectID: project.ID,
		Module:    domain.ModuleAuth,
		Function:  "signup",
		FieldID:   field.ID,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(publisher.events) != 1 || publisher.events[0].Name != "project.config.updated" {
		t.Fatalf("expected project.config.updated event, got %v", publisher.events)
	}
}

func TestConfigModule_InvalidFunction(t *testing.T) {
	projects := newMockProjectRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)
	field, _ := domain.NewField("entity1", "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)

	uc := NewConfigModule(projects, fields, modules, publisher)
	err := uc.ToggleField(context.Background(), ToggleFieldInput{
		ProjectID: project.ID,
		Module:    domain.ModuleAuth,
		Function:  "unknownFn",
		FieldID:   field.ID,
		Enabled:   true,
	})
	if err != domain.ErrFunctionNotFound {
		t.Fatalf("expected ErrFunctionNotFound, got %v", err)
	}
}

func TestConfigModule_UnspecifiedModuleFunctionsRejected(t *testing.T) {
	projects := newMockProjectRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)
	field, _ := domain.NewField("entity1", "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)

	uc := NewConfigModule(projects, fields, modules, publisher)
	err := uc.ToggleField(context.Background(), ToggleFieldInput{
		ProjectID: project.ID,
		Module:    domain.ModuleCRUD,
		Function:  "list",
		FieldID:   field.ID,
		Enabled:   true,
	})
	if err != domain.ErrFunctionNotFound {
		t.Fatalf("expected ErrFunctionNotFound, got %v", err)
	}
}

func TestConfigModule_ListModules(t *testing.T) {
	projects := newMockProjectRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)
	field, _ := domain.NewField("entity1", "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)

	uc := NewConfigModule(projects, fields, modules, publisher)
	_ = uc.ToggleField(context.Background(), ToggleFieldInput{
		ProjectID: project.ID,
		Module:    domain.ModuleAuth,
		Function:  "signup",
		FieldID:   field.ID,
		Enabled:   true,
	})

	list, err := uc.ListModules(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 module, got %d", len(list))
	}
}
