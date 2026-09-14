package port

import (
	"context"

	"backify/services/control-plane/internal/domain"
)

type FieldUsage struct {
	ModuleName   string
	FunctionName string
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
	DisableFieldEverywhere(ctx context.Context, fieldID string) error
}
