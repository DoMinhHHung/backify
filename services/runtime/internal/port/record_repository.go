package port

import (
	"context"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type RecordRepository interface {
	Create(ctx context.Context, schemaName, table string, columns map[string]any) (domain.Record, error)
	FindByID(ctx context.Context, schemaName, table, id string) (domain.Record, error)
	List(ctx context.Context, schemaName, table string, ownerColumn, ownerID string, limit, offset int) ([]domain.Record, int, error)
	Update(ctx context.Context, schemaName, table, id string, columns map[string]any) (domain.Record, error)
	Delete(ctx context.Context, schemaName, table, id string) error
}
