package port

import (
	"context"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type ConfigClient interface {
	GetProject(ctx context.Context, projectID string) (*domain.ProjectMeta, error)
	GetProjectConfig(ctx context.Context, projectID string) (*domain.ProjectConfig, error)
}
