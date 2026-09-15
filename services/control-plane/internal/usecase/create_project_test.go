package usecase

import (
	"context"
	"testing"

	"backify/services/control-plane/internal/domain"
)

func TestCreateProject_Success(t *testing.T) {
	projects := newMockProjectRepo()
	publisher := &mockEventPublisher{}
	uc := NewCreateProject(projects, publisher)

	project, err := uc.Execute(context.Background(), CreateProjectInput{Name: "Shop App", Subdomain: "shop-app"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if project.ID == "" {
		t.Fatal("expected project to have an id")
	}
	if len(publisher.events) != 1 || publisher.events[0].Name != "project.created" {
		t.Fatalf("expected project.created event to be published, got %v", publisher.events)
	}
}

func TestCreateProject_InvalidSubdomain(t *testing.T) {
	projects := newMockProjectRepo()
	publisher := &mockEventPublisher{}
	uc := NewCreateProject(projects, publisher)

	_, err := uc.Execute(context.Background(), CreateProjectInput{Name: "Shop App", Subdomain: "a"})
	if err == nil {
		t.Fatal("expected error for invalid subdomain")
	}
	if len(publisher.events) != 0 {
		t.Fatal("expected no event published on validation failure")
	}
}

func TestCreateProject_DuplicateSubdomain(t *testing.T) {
	projects := newMockProjectRepo()
	publisher := &mockEventPublisher{}
	uc := NewCreateProject(projects, publisher)

	ctx := context.Background()
	if _, err := uc.Execute(ctx, CreateProjectInput{Name: "Shop App", Subdomain: "shop-app"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err := uc.Execute(ctx, CreateProjectInput{Name: "Another Shop", Subdomain: "shop-app"})
	if err != domain.ErrSubdomainTaken {
		t.Fatalf("expected ErrSubdomainTaken, got %v", err)
	}
}
