package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

type ProjectCreatedEvent struct {
	ProjectID  string `json:"project_id"`
	Name       string `json:"name"`
	Subdomain  string `json:"subdomain"`
	SchemaName string `json:"schema_name"`
}

type CreateProjectInput struct {
	Name      string
	Subdomain string
}

type CreateProject struct {
	projects  port.ProjectRepository
	publisher port.EventPublisher
}

// NewCreateProject tạo use case tạo project từ kho lưu trữ và bộ phát sự kiện.
func NewCreateProject(projects port.ProjectRepository, publisher port.EventPublisher) *CreateProject {
	return &CreateProject{projects: projects, publisher: publisher}
}

// Execute tạo và lưu project, sau đó phát sự kiện project.created.
// Project đã lưu không được hoàn tác nếu bộ phát sự kiện trả về lỗi.
func (uc *CreateProject) Execute(ctx context.Context, input CreateProjectInput) (*domain.Project, error) {
	project, err := domain.NewProject(input.Name, input.Subdomain)
	if err != nil {
		return nil, err
	}

	if err := uc.projects.Create(ctx, project); err != nil {
		return nil, err
	}

	if err := uc.publisher.Publish(ctx, port.Event{
		Name: "project.created",
		Payload: ProjectCreatedEvent{
			ProjectID:  project.ID,
			Name:       project.Name,
			Subdomain:  project.Subdomain,
			SchemaName: project.SchemaName,
		},
	}); err != nil {
		return nil, err
	}

	return project, nil
}
