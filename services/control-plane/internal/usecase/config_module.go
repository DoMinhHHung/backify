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
	entities  port.EntityRepository
	fields    port.FieldRepository
	modules   port.ModuleRepository
	publisher port.EventPublisher
}

func NewConfigModule(projects port.ProjectRepository, entities port.EntityRepository, fields port.FieldRepository, modules port.ModuleRepository, publisher port.EventPublisher) *ConfigModule {
	return &ConfigModule{projects: projects, entities: entities, fields: fields, modules: modules, publisher: publisher}
}

// ToggleField bật hoặc tắt một field thuộc project cho function được hỗ trợ của module auth.
// Module và function được tạo khi cần; cấu hình đã lưu không được hoàn tác nếu phát sự kiện thất bại.
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

	field, err := uc.fields.GetByID(ctx, input.FieldID)
	if err != nil {
		return err
	}

	entity, err := uc.entities.GetByID(ctx, field.EntityID)
	if err != nil {
		if err != domain.ErrEntityNotFound {
			return err
		}
		return domain.ErrFieldNotFound
	}
	if entity.ProjectID != input.ProjectID {
		return domain.ErrFieldNotFound
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

// ListModules trả về các module đã cấu hình sau khi xác nhận project tồn tại.
func (uc *ConfigModule) ListModules(ctx context.Context, projectID string) ([]*domain.Module, error) {
	if _, err := uc.projects.GetByID(ctx, projectID); err != nil {
		return nil, err
	}
	return uc.modules.ListByProject(ctx, projectID)
}
