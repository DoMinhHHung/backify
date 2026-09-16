package port

import (
	"context"

	"backify/services/control-plane/internal/domain"
)

type FieldUsage struct {
	ModuleName   string
	FunctionName string
}

// FunctionConfig là read-model cho một function: tên và field ID đang bật.
// Tách khỏi domain.Module vì đây là dữ liệu tổng hợp chỉ phục vụ đọc
// (GetProjectConfig qua gRPC), không phải aggregate có hành vi ghi.
type FunctionConfig struct {
	ID              string
	Name            string
	EnabledFieldIDs []string
}

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*domain.Project, error)
	List(ctx context.Context) ([]*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
}

type EntityRepository interface {
	Create(ctx context.Context, entity *domain.Entity) error
	GetByID(ctx context.Context, id string) (*domain.Entity, error)
	GetByProjectAndName(ctx context.Context, projectID, name string) (*domain.Entity, error)
	ListByProject(ctx context.Context, projectID string) ([]*domain.Entity, error)
}

type FieldRepository interface {
	Create(ctx context.Context, field *domain.Field) error
	GetByID(ctx context.Context, id string) (*domain.Field, error)
	GetByEntityAndName(ctx context.Context, entityID, name string) (*domain.Field, error)
	ListByEntity(ctx context.Context, entityID string) ([]*domain.Field, error)
	Delete(ctx context.Context, id string) error
}

type ModuleRepository interface {
	Create(ctx context.Context, module *domain.Module) error
	GetByProjectAndName(ctx context.Context, projectID string, name domain.ModuleName) (*domain.Module, error)
	ListByProject(ctx context.Context, projectID string) ([]*domain.Module, error)
	EnsureFunction(ctx context.Context, moduleID, functionName string) (string, error)
	ToggleFunctionField(ctx context.Context, functionID, fieldID string, enabled bool) error
	ListFieldUsages(ctx context.Context, fieldID string) ([]FieldUsage, error)
	// ListFunctionsByModule trả về mọi function của module kèm field ID đang
	// bật (enabled = true); function chưa toggle field nào có EnabledFieldIDs rỗng.
	ListFunctionsByModule(ctx context.Context, moduleID string) ([]FunctionConfig, error)
	DisableFieldEverywhere(ctx context.Context, fieldID string) error
	DisableAndDeleteField(ctx context.Context, fieldID string) error
}
