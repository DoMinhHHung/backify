package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

type GetProject struct {
	projects port.ProjectRepository
}

// NewGetProject tạo use case đọc và xóa mềm project.
func NewGetProject(projects port.ProjectRepository) *GetProject {
	return &GetProject{projects: projects}
}

// Execute trả về project theo mã định danh.
func (uc *GetProject) Execute(ctx context.Context, id string) (*domain.Project, error) {
	return uc.projects.GetByID(ctx, id)
}

// List trả về toàn bộ project từ kho lưu trữ.
func (uc *GetProject) List(ctx context.Context) ([]*domain.Project, error) {
	return uc.projects.List(ctx)
}

// Delete đánh dấu project đã xóa rồi lưu lại, không xóa bản ghi vật lý.
func (uc *GetProject) Delete(ctx context.Context, id string) error {
	project, err := uc.projects.GetByID(ctx, id)
	if err != nil {
		return err
	}
	project.MarkDeleted()
	return uc.projects.Update(ctx, project)
}
