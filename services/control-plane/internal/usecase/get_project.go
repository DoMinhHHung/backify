package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

// ProjectDeletedEvent là payload publish khi project bị xóa (mềm) — Auth
// Service subscribe sự kiện này để DROP DATABASE auth_proj_<project_id>
// (Bước 3 §Auth), Runtime subscribe để ngừng phục vụ request của project đó.
type ProjectDeletedEvent struct {
	ProjectID string `json:"project_id"`
}

// GetProject đọc và (soft-)xóa project. Gộp Delete vào đây thay vì tách
// usecase riêng vì Delete chỉ đổi Status qua ProjectRepository giống Execute,
// không có input phức tạp như CreateProject.
type GetProject struct {
	projects  port.ProjectRepository
	publisher port.EventPublisher
}

// NewGetProject nhận thêm publisher (khác với bản cũ chỉ nhận projects) vì
// Delete cần publish project.deleted — Auth/Runtime không có cách nào khác
// biết project đã bị xóa để dọn dẹp (không polling định kỳ ở MVP).
func NewGetProject(projects port.ProjectRepository, publisher port.EventPublisher) *GetProject {
	return &GetProject{projects: projects, publisher: publisher}
}

func (uc *GetProject) Execute(ctx context.Context, id string) (*domain.Project, error) {
	return uc.projects.GetByID(ctx, id)
}

func (uc *GetProject) List(ctx context.Context) ([]*domain.Project, error) {
	return uc.projects.List(ctx)
}

// Delete đánh dấu project đã xóa, lưu trạng thái mới rồi phát project.deleted.
// Publish lỗi được trả về nguyên trạng giống CreateProject.Execute (project
// đã Update không được hoàn tác) — chấp nhận được ở MVP, chưa có outbox
// pattern để đảm bảo publish-hay-không cùng transaction với Update.
func (uc *GetProject) Delete(ctx context.Context, id string) error {
	project, err := uc.projects.GetByID(ctx, id)
	if err != nil {
		return err
	}
	project.MarkDeleted()
	if err := uc.projects.Update(ctx, project); err != nil {
		return err
	}

	return uc.publisher.Publish(ctx, port.Event{
		Name:    "project.deleted",
		Payload: ProjectDeletedEvent{ProjectID: project.ID},
	})
}
