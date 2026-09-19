package usecase

import (
	"context"
	"testing"

	"backify/services/auth/internal/domain"
)

func TestSignUp_Success(t *testing.T) {
	users := newMockUserRepo()
	refresh := newMockRefreshTokenRepo()
	hasher := newMockHasher()
	issuer := newMockIssuer()
	pub := newMockEventPublisher()
	uc := NewSignUp(users, refresh, hasher, issuer, pub)

	out, err := uc.Execute(context.Background(), SignUpInput{
		ProjectID:       "proj1",
		Email:           "alice@example.com",
		Password:        "password123",
		FullName:        "Alice",
		Phone:           "+84901234567",
		RefreshDuration: "7d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User.Email != "alice@example.com" {
		t.Errorf("email = %q", out.User.Email)
	}
	if out.User.PasswordHash == "" {
		t.Error("password hash should be set")
	}
	if out.Tokens.AccessToken == "" || out.Tokens.RefreshToken == "" {
		t.Error("tokens should be issued")
	}
	if out.Tokens.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", out.Tokens.ExpiresIn)
	}
	if pub.lastEventName() != "auth.user.signup" {
		t.Errorf("event = %q, want auth.user.signup", pub.lastEventName())
	}
}

func TestSignUp_DuplicateEmail(t *testing.T) {
	users := newMockUserRepo()
	refresh := newMockRefreshTokenRepo()
	uc := NewSignUp(users, refresh, newMockHasher(), newMockIssuer(), newMockEventPublisher())

	in := SignUpInput{
		ProjectID:       "proj1",
		Email:           "bob@example.com",
		Password:        "password123",
		RefreshDuration: "7d",
	}
	if _, err := uc.Execute(context.Background(), in); err != nil {
		t.Fatalf("first signup: %v", err)
	}
	_, err := uc.Execute(context.Background(), in)
	if err != domain.ErrEmailTaken {
		t.Errorf("got %v, want ErrEmailTaken", err)
	}
}

func TestSignUp_InvalidPassword(t *testing.T) {
	uc := NewSignUp(newMockUserRepo(), newMockRefreshTokenRepo(), newMockHasher(), newMockIssuer(), newMockEventPublisher())
	_, err := uc.Execute(context.Background(), SignUpInput{
		ProjectID:       "proj1",
		Email:           "a@b.com",
		Password:        "short",
		RefreshDuration: "7d",
	})
	if err == nil {
		t.Fatal("expected error for short password")
	}
	de, ok := err.(*domain.Error)
	if !ok || de.Code != domain.CodeInvalidInput {
		t.Errorf("got %v, want INVALID_INPUT", err)
	}
}

func TestSignUp_InvalidRefreshDuration(t *testing.T) {
	uc := NewSignUp(newMockUserRepo(), newMockRefreshTokenRepo(), newMockHasher(), newMockIssuer(), newMockEventPublisher())
	_, err := uc.Execute(context.Background(), SignUpInput{
		ProjectID:       "proj1",
		Email:           "a@b.com",
		Password:        "password123",
		RefreshDuration: "1d",
	})
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}
