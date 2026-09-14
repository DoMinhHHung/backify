package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

type GetProject struct {
	projects port.ProjectRepository
}

func NewGetProject(projects port.ProjectRepository) *GetProject {
	return &GetProject{projects: projects}
}

func (uc *GetProject) Execute(ctx context.Context, id string) (*domain.Project, error) {
	return uc.projects.GetByID(ctx, id)
}

func (uc *GetProject) List(ctx context.Context) ([]*domain.Project, error) {
	return uc.projects.List(ctx)
}

func (uc *GetProject) Delete(ctx context.Context, id string) error {
	project, err := uc.projects.GetByID(ctx, id)
	if err != nil {
		return err
	}
	project.MarkDeleted()
	return uc.projects.Update(ctx, project)
}
