package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

type AddEntityInput struct {
	ProjectID string
	Name      string
}

type AddEntity struct {
	projects port.ProjectRepository
	entities port.EntityRepository
}

// NewAddEntity tạo use case quản lý entity từ các kho lưu trữ cần thiết.
func NewAddEntity(projects port.ProjectRepository, entities port.EntityRepository) *AddEntity {
	return &AddEntity{projects: projects, entities: entities}
}

// Execute kiểm tra project tồn tại, tạo entity hợp lệ rồi lưu và trả về entity đó.
func (uc *AddEntity) Execute(ctx context.Context, input AddEntityInput) (*domain.Entity, error) {
	if _, err := uc.projects.GetByID(ctx, input.ProjectID); err != nil {
		return nil, err
	}

	entity, err := domain.NewEntity(input.ProjectID, input.Name)
	if err != nil {
		return nil, err
	}

	if err := uc.entities.Create(ctx, entity); err != nil {
		return nil, err
	}

	return entity, nil
}

// List trả về các entity của project sau khi xác nhận project tồn tại.
func (uc *AddEntity) List(ctx context.Context, projectID string) ([]*domain.Entity, error) {
	if _, err := uc.projects.GetByID(ctx, projectID); err != nil {
		return nil, err
	}
	return uc.entities.ListByProject(ctx, projectID)
}
