package port

import (
	"context"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type ConfigCache interface {
	Get(ctx context.Context, projectID string) (*domain.ProjectConfig, bool)
	Set(ctx context.Context, cfg *domain.ProjectConfig)
	Invalidate(ctx context.Context, projectID string)
}
