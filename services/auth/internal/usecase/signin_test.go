package usecase

import (
	"context"
	"testing"

	"backify/services/auth/internal/domain"
)

func seedUser(t *testing.T, users *mockUserRepo, projectID, email, password string) *domain.User {
	t.Helper()
	hasher := newMockHasher()
	user, err := domain.NewUser(projectID, email, "Test", "")
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	hash, _ := hasher.Hash(password)
	_ = user.SetPasswordHash(hash)
	if err := users.Create(context.Background(), projectID, user); err != nil {
		t.Fatalf("Create: %v", err)
	}
	return user
}

func TestSignIn_Success(t *testing.T) {
	users := newMockUserRepo()
	refresh := newMockRefreshTokenRepo()
	pub := newMockEventPublisher()
	seedUser(t, users, "proj1", "alice@example.com", "password123")

	uc := NewSignIn(users, refresh, newMockHasher(), newMockIssuer(), pub)
	out, err := uc.Execute(context.Background(), SignInInput{
		ProjectID:       "proj1",
		Email:           "alice@example.com",
		Password:        "password123",
		RefreshDuration: "30d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Tokens.AccessToken == "" {
		t.Error("access token empty")
	}
	if pub.lastEventName() != "auth.user.signin" {
		t.Errorf("event = %q", pub.lastEventName())
	}
}

func TestSignIn_WrongPassword(t *testing.T) {
	users := newMockUserRepo()
	seedUser(t, users, "proj1", "alice@example.com", "password123")

	uc := NewSignIn(users, newMockRefreshTokenRepo(), newMockHasher(), newMockIssuer(), newMockEventPublisher())
	_, err := uc.Execute(context.Background(), SignInInput{
		ProjectID:       "proj1",
		Email:           "alice@example.com",
		Password:        "wrongpassword",
		RefreshDuration: "7d",
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("got %v, want ErrInvalidCredentials", err)
	}
}

func TestSignIn_UnknownEmail_NoEnumeration(t *testing.T) {
	uc := NewSignIn(newMockUserRepo(), newMockRefreshTokenRepo(), newMockHasher(), newMockIssuer(), newMockEventPublisher())
	_, err := uc.Execute(context.Background(), SignInInput{
		ProjectID:       "proj1",
		Email:           "nobody@example.com",
		Password:        "password123",
		RefreshDuration: "7d",
	})
	// Không được lộ ErrUserNotFound — phải là ErrInvalidCredentials.
	if err != domain.ErrInvalidCredentials {
		t.Errorf("got %v, want ErrInvalidCredentials (anti-enumeration)", err)
	}
}
