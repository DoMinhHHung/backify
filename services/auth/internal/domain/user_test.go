package domain

import (
	"errors"
	"testing"
)

func TestNewUser_Valid(t *testing.T) {
	u, err := NewUser("proj1", "User@Example.com", "Alice", "+84901234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID == "" {
		t.Fatal("expected generated id")
	}
	if u.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", u.Email)
	}
	if u.ProjectID != "proj1" {
		t.Fatalf("expected project id proj1, got %s", u.ProjectID)
	}
}

func TestNewUser_InvalidEmail(t *testing.T) {
	_, err := NewUser("proj1", "not-an-email", "", "")
	if err == nil {
		t.Fatal("expected error")
	}
	var de *Error
	if !errors.As(err, &de) || de.Code != CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", err)
	}
}

func TestNewUser_EmptyProject(t *testing.T) {
	_, err := NewUser("", "a@b.com", "", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("short"); err == nil {
		t.Fatal("expected error for short password")
	}
	if err := ValidatePassword("longenough"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewUser_InvalidPhone(t *testing.T) {
	_, err := NewUser("proj1", "a@b.com", "", "abc")
	if err == nil {
		t.Fatal("expected error for invalid phone")
	}
}

func TestUser_SetPasswordHash(t *testing.T) {
	u, _ := NewUser("proj1", "a@b.com", "", "")
	if err := u.SetPasswordHash(""); err == nil {
		t.Fatal("expected error for empty hash")
	}
	if err := u.SetPasswordHash("$argon2id$..."); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.PasswordHash == "" {
		t.Fatal("expected hash set")
	}
}
