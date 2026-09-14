package domain

import "testing"

func TestNewEntity_Valid(t *testing.T) {
	e, err := NewEntity("proj123", "product")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if e.IsSystem {
		t.Fatal("expected non-system entity for name product")
	}
}

func TestNewEntity_SystemUser(t *testing.T) {
	e, err := NewEntity("proj123", "user")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !e.IsSystem {
		t.Fatal("expected user entity to be system entity")
	}
}

func TestNewEntity_EmptyProjectID(t *testing.T) {
	_, err := NewEntity("", "product")
	if err == nil {
		t.Fatal("expected error for empty project id")
	}
}

func TestNewEntity_InvalidName(t *testing.T) {
	cases := []string{"", "Product", "product-name", "product name"}
	for _, name := range cases {
		_, err := NewEntity("proj123", name)
		if err == nil {
			t.Fatalf("expected error for name %q", name)
		}
	}
}

func TestEntity_CanDelete(t *testing.T) {
	system, _ := NewEntity("proj123", "user")
	if err := system.CanDelete(); err != ErrSystemEntityCannotDelete {
		t.Fatalf("expected ErrSystemEntityCannotDelete, got %v", err)
	}

	custom, _ := NewEntity("proj123", "product")
	if err := custom.CanDelete(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
