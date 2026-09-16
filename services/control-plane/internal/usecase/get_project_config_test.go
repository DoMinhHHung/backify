package usecase

import (
	"context"
	"testing"

	"backify/services/control-plane/internal/domain"
)

func TestGetProjectConfig_Execute(t *testing.T) {
	ctx := context.Background()
	projects := newMockProjectRepo()
	entities := newMockEntityRepo()
	fields := newMockFieldRepo()
	modules := newMockModuleRepo()

	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(ctx, project)

	entity, _ := domain.NewEntity(project.ID, "user")
	_ = entities.Create(ctx, entity)

	emailField, _ := domain.NewField(entity.ID, "email", domain.FieldTypeEmail)
	_ = fields.Create(ctx, emailField)

	module, _ := domain.NewModule(project.ID, domain.ModuleAuth)
	_ = modules.Create(ctx, module)

	functionID, _ := modules.EnsureFunction(ctx, module.ID, "signup")
	_ = modules.ToggleFunctionField(ctx, functionID, emailField.ID, true)

	uc := NewGetProjectConfig(projects, entities, fields, modules)
	cfg, err := uc.Execute(ctx, project.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Project.ID != project.ID {
		t.Fatalf("expected project id %s, got %s", project.ID, cfg.Project.ID)
	}
	if len(cfg.Entities) != 1 || len(cfg.Entities[0].Fields) != 1 {
		t.Fatalf("expected 1 entity with 1 field, got %+v", cfg.Entities)
	}
	if cfg.Entities[0].Fields[0].ID != emailField.ID {
		t.Fatalf("expected field id %s, got %s", emailField.ID, cfg.Entities[0].Fields[0].ID)
	}
	if len(cfg.Modules) != 1 || len(cfg.Modules[0].Functions) != 1 {
		t.Fatalf("expected 1 module with 1 function, got %+v", cfg.Modules)
	}
	fn := cfg.Modules[0].Functions[0]
	if fn.Name != "signup" || len(fn.EnabledFieldIDs) != 1 || fn.EnabledFieldIDs[0] != emailField.ID {
		t.Fatalf("expected signup function with email enabled, got %+v", fn)
	}
}

func TestGetProjectConfig_EmptyProject(t *testing.T) {
	ctx := context.Background()
	projects := newMockProjectRepo()
	project, _ := domain.NewProject("Blog App", "blog-app")
	_ = projects.Create(ctx, project)

	uc := NewGetProjectConfig(projects, newMockEntityRepo(), newMockFieldRepo(), newMockModuleRepo())
	cfg, err := uc.Execute(ctx, project.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Entities) != 0 || len(cfg.Modules) != 0 {
		t.Fatalf("expected empty entities/modules, got %+v", cfg)
	}
}

func TestGetProjectConfig_NotFound(t *testing.T) {
	uc := NewGetProjectConfig(newMockProjectRepo(), newMockEntityRepo(), newMockFieldRepo(), newMockModuleRepo())
	_, err := uc.Execute(context.Background(), "missing")
	if err != domain.ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}
