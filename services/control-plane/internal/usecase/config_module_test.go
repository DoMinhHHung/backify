package usecase

import (
	"context"
	"testing"

	"backify/services/control-plane/internal/domain"
)

func TestConfigModule_ToggleField_Success(t *testing.T) {
	projects := newMockProjectRepo()
	entities := newMockEntityRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)
	entity, _ := domain.NewEntity(project.ID, "user")
	_ = entities.Create(context.Background(), entity)
	field, _ := domain.NewField(entity.ID, "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)

	uc := NewConfigModule(projects, entities, fields, modules, publisher)
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

func TestConfigModule_FieldFromOtherProject(t *testing.T) {
	projects := newMockProjectRepo()
	entities := newMockEntityRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)
	other, _ := domain.NewProject("Other", "other-app")
	_ = projects.Create(context.Background(), other)
	entity, _ := domain.NewEntity(other.ID, "product")
	_ = entities.Create(context.Background(), entity)
	field, _ := domain.NewField(entity.ID, "title", domain.FieldTypeString)
	_ = fields.Create(context.Background(), field)

	uc := NewConfigModule(projects, entities, fields, modules, publisher)
	err := uc.ToggleField(context.Background(), ToggleFieldInput{
		ProjectID: project.ID,
		Module:    domain.ModuleAuth,
		Function:  "signup",
		FieldID:   field.ID,
		Enabled:   true,
	})
	if err != domain.ErrFieldNotFound {
		t.Fatalf("expected ErrFieldNotFound, got %v", err)
	}
}

func TestConfigModule_InvalidFunction(t *testing.T) {
	projects := newMockProjectRepo()
	entities := newMockEntityRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)
	entity, _ := domain.NewEntity(project.ID, "user")
	_ = entities.Create(context.Background(), entity)
	field, _ := domain.NewField(entity.ID, "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)

	uc := NewConfigModule(projects, entities, fields, modules, publisher)
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

func TestConfigModule_NonAuthModuleDisabled(t *testing.T) {
	projects := newMockProjectRepo()
	entities := newMockEntityRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)
	entity, _ := domain.NewEntity(project.ID, "user")
	_ = entities.Create(context.Background(), entity)
	field, _ := domain.NewField(entity.ID, "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)

	uc := NewConfigModule(projects, entities, fields, modules, publisher)
	err := uc.ToggleField(context.Background(), ToggleFieldInput{
		ProjectID: project.ID,
		Module:    domain.ModuleCRUD,
		Function:  "list",
		FieldID:   field.ID,
		Enabled:   true,
	})
	if err != domain.ErrModuleNotEnabled {
		t.Fatalf("expected ErrModuleNotEnabled, got %v", err)
	}
}

func TestConfigModule_ListModules(t *testing.T) {
	projects := newMockProjectRepo()
	entities := newMockEntityRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()
	publisher := &mockEventPublisher{}

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)
	entity, _ := domain.NewEntity(project.ID, "user")
	_ = entities.Create(context.Background(), entity)
	field, _ := domain.NewField(entity.ID, "bio", domain.FieldTypeText)
	_ = fields.Create(context.Background(), field)

	uc := NewConfigModule(projects, entities, fields, modules, publisher)
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
