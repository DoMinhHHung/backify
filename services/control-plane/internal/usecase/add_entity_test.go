package usecase

import (
	"context"
	"testing"

	"backify/services/control-plane/internal/domain"
)

func TestAddEntity_Success(t *testing.T) {
	projects := newMockProjectRepo()
	entities := newMockEntityRepo()
	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)

	uc := NewAddEntity(projects, entities)
	entity, err := uc.Execute(context.Background(), AddEntityInput{ProjectID: project.ID, Name: "product"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entity.IsSystem {
		t.Fatal("expected non-system entity")
	}
}

func TestAddEntity_ProjectNotFound(t *testing.T) {
	uc := NewAddEntity(newMockProjectRepo(), newMockEntityRepo())
	_, err := uc.Execute(context.Background(), AddEntityInput{ProjectID: "missing", Name: "product"})
	if err != domain.ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestAddEntity_DuplicateName(t *testing.T) {
	projects := newMockProjectRepo()
	entities := newMockEntityRepo()
	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)

	uc := NewAddEntity(projects, entities)
	ctx := context.Background()
	if _, err := uc.Execute(ctx, AddEntityInput{ProjectID: project.ID, Name: "product"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err := uc.Execute(ctx, AddEntityInput{ProjectID: project.ID, Name: "product"})
	if err != domain.ErrEntityNameTaken {
		t.Fatalf("expected ErrEntityNameTaken, got %v", err)
	}
}

func TestAddEntity_List(t *testing.T) {
	projects := newMockProjectRepo()
	entities := newMockEntityRepo()
	project, _ := domain.NewProject("Shop App", "shop-app")
	_ = projects.Create(context.Background(), project)

	uc := NewAddEntity(projects, entities)
	ctx := context.Background()
	_, _ = uc.Execute(ctx, AddEntityInput{ProjectID: project.ID, Name: "product"})
	_, _ = uc.Execute(ctx, AddEntityInput{ProjectID: project.ID, Name: "order"})

	list, err := uc.List(ctx, project.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(list))
	}
}
