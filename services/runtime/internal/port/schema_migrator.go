package port

import (
	"context"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type SchemaMigrator interface {
	EnsureSchema(ctx context.Context, cfg *domain.ProjectConfig) error
	AppliedVersion(ctx context.Context, schemaName string) (int32, error)
}
