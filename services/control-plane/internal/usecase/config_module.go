package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

type ConfigModuleEvent struct {
	ProjectID string `json:"project_id"`
	Module    string `json:"module"`
	Function  string `json:"function"`
	FieldID   string `json:"field_id"`
	Enabled   bool   `json:"enabled"`
}

type ToggleFieldInput struct {
	ProjectID string
	Module    domain.ModuleName
	Function  string
	FieldID   string
	Enabled   bool
}

type ConfigModule struct {
	projects  port.ProjectRepository
	fields    port.FieldRepository
	modules   port.ModuleRepository
	publisher port.EventPublisher
}

func NewConfigModule(projects port.ProjectRepository, fields port.FieldRepository, modules port.ModuleRepository, publisher port.EventPublisher) *ConfigModule {
	return &ConfigModule{projects: projects, fields: fields, modules: modules, publisher: publisher}
}

func (uc *ConfigModule) ToggleField(ctx context.Context, input ToggleFieldInput) error {
	if _, err := uc.projects.GetByID(ctx, input.ProjectID); err != nil {
		return err
	}

	if !input.Module.Valid() {
		return domain.ErrInvalidModuleName
	}

	if !input.Module.Enabled() {
		return domain.ErrModuleNotEnabled
	}

	if !domain.ValidFunction(input.Module, input.Function) {
		return domain.ErrFunctionNotFound
	}

	if _, err := uc.fields.GetByID(ctx, input.FieldID); err != nil {
		return err
	}

	module, err := uc.modules.GetByProjectAndName(ctx, input.ProjectID, input.Module)
	if err != nil {
		if err != domain.ErrModuleNotFound {
			return err
		}
		newModule, createErr := domain.NewModule(input.ProjectID, input.Module)
		if createErr != nil {
			return createErr
		}
		if createErr := uc.modules.Create(ctx, newModule); createErr != nil {
			return createErr
		}
		module = newModule
	}

	functionID, err := uc.modules.EnsureFunction(ctx, module.ID, input.Function)
	if err != nil {
		return err
	}

	if err := uc.modules.ToggleFunctionField(ctx, functionID, input.FieldID, input.Enabled); err != nil {
		return err
	}

	return uc.publisher.Publish(ctx, port.Event{
		Name: "project.config.updated",
		Payload: ConfigModuleEvent{
			ProjectID: input.ProjectID,
			Module:    string(input.Module),
			Function:  input.Function,
			FieldID:   input.FieldID,
			Enabled:   input.Enabled,
		},
	})
}

func (uc *ConfigModule) ListModules(ctx context.Context, projectID string) ([]*domain.Module, error) {
	if _, err := uc.projects.GetByID(ctx, projectID); err != nil {
		return nil, err
	}
	return uc.modules.ListByProject(ctx, projectID)
}
