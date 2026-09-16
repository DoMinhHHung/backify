package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

// EntityConfig là entity kèm theo field của nó — read-model dùng riêng cho
// GetProjectConfig, tách khỏi domain.Entity/domain.Field vì đây là dữ liệu
// tổng hợp chỉ phục vụ đọc qua gRPC, không phải aggregate có hành vi ghi.
type EntityConfig struct {
	Entity *domain.Entity
	Fields []*domain.Field
}

// ModuleConfig là module kèm function và field đang bật của từng function.
type ModuleConfig struct {
	Module    *domain.Module
	Functions []port.FunctionConfig
}

// ProjectConfig gộp toàn bộ dữ liệu Auth/Runtime cần để phục vụ request của
// end-user: project, entity/field, module/function và field nào đang bật.
type ProjectConfig struct {
	Project  *domain.Project
	Entities []EntityConfig
	Modules  []ModuleConfig
}

// GetProjectConfig tổng hợp config đầy đủ của một project cho gRPC
// ControlPlaneService.GetProjectConfig. Tách khỏi GetProject vì đây là một
// tổ hợp nhiều repository, không phải một lượt đọc đơn.
type GetProjectConfig struct {
	projects port.ProjectRepository
	entities port.EntityRepository
	fields   port.FieldRepository
	modules  port.ModuleRepository
}

func NewGetProjectConfig(projects port.ProjectRepository, entities port.EntityRepository, fields port.FieldRepository, modules port.ModuleRepository) *GetProjectConfig {
	return &GetProjectConfig{projects: projects, entities: entities, fields: fields, modules: modules}
}

// Execute trả về ErrProjectNotFound nếu project không tồn tại; mọi lỗi đọc
// entity/field/module khác được trả nguyên trạng để caller (gRPC handler)
// map sang mã lỗi phù hợp.
func (uc *GetProjectConfig) Execute(ctx context.Context, projectID string) (*ProjectConfig, error) {
	project, err := uc.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	entities, err := uc.entities.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	entityConfigs := make([]EntityConfig, 0, len(entities))
	for _, entity := range entities {
		fields, err := uc.fields.ListByEntity(ctx, entity.ID)
		if err != nil {
			return nil, err
		}
		entityConfigs = append(entityConfigs, EntityConfig{Entity: entity, Fields: fields})
	}

	modules, err := uc.modules.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	moduleConfigs := make([]ModuleConfig, 0, len(modules))
	for _, module := range modules {
		functions, err := uc.modules.ListFunctionsByModule(ctx, module.ID)
		if err != nil {
			return nil, err
		}
		moduleConfigs = append(moduleConfigs, ModuleConfig{Module: module, Functions: functions})
	}

	return &ProjectConfig{Project: project, Entities: entityConfigs, Modules: moduleConfigs}, nil
}
