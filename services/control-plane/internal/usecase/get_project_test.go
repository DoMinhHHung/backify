package usecase

import (
	"context"
	"testing"

	"backify/services/control-plane/internal/domain"
)

func TestGetProject_Execute(t *testing.T) {
	projects := newMockProjectRepo()
	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)

	uc := NewGetProject(projects)
	fetched, err := uc.Execute(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fetched.ID != project.ID {
		t.Fatalf("expected project id %s, got %s", project.ID, fetched.ID)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	uc := NewGetProject(newMockProjectRepo())
	_, err := uc.Execute(context.Background(), "missing")
	if err != domain.ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestGetProject_List(t *testing.T) {
	projects := newMockProjectRepo()
	p1, _ := domain.NewProject("Shop App", "shop-app")
	p2, _ := domain.NewProject("Blog App", "blog-app")
	_ = projects.Create(context.Background(), p1)
	_ = projects.Create(context.Background(), p2)

	uc := NewGetProject(projects)
	list, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(list))
	}
}

func TestGetProject_Delete(t *testing.T) {
	projects := newMockProjectRepo()
	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)

	uc := NewGetProject(projects)
	if err := uc.Delete(context.Background(), project.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	fetched, _ := projects.GetByID(context.Background(), project.ID)
	if fetched.Status != domain.ProjectStatusDeleted {
		t.Fatalf("expected status deleted, got %s", fetched.Status)
	}
}
