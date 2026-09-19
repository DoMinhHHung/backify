package usecase

import (
	"context"
	"testing"
	"time"

	"backify/services/auth/internal/domain"
)

func TestResetPassword_Success(t *testing.T) {
	users := newMockUserRepo()
	resets := newMockPasswordResetRepo()
	refresh := newMockRefreshTokenRepo()
	user := seedUser(t, users, "proj1", "alice@example.com", "oldpassword")

	// Refresh token còn sống — phải bị revoke sau reset.
	rt, _, _ := domain.NewRefreshToken("proj1", user.ID, domain.RefreshDuration(domain.RefreshDuration7d))
	_ = refresh.Create(context.Background(), "proj1", rt)

	prt, raw, err := domain.NewPasswordResetToken("proj1", user.ID)
	if err != nil {
		t.Fatal(err)
	}
	_ = resets.Create(context.Background(), "proj1", prt)

	uc := NewResetPassword(users, resets, refresh, newMockHasher())
	if err := uc.Execute(context.Background(), ResetPasswordInput{
		ProjectID:   "proj1",
		ResetToken:  raw,
		NewPassword: "newpassword99",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Password đã đổi.
	got, _ := users.GetByID(context.Background(), "proj1", user.ID)
	if got.PasswordHash != "hashed:newpassword99" {
		t.Errorf("password hash = %q", got.PasswordHash)
	}

	// Reset token đã used.
	stored, _ := resets.GetByTokenHash(context.Background(), "proj1", domain.HashToken(raw))
	if !stored.IsUsed() {
		t.Error("reset token should be marked used")
	}

	// Mọi refresh token của user bị revoke.
	rtGot, _ := refresh.GetByTokenHash(context.Background(), "proj1", rt.TokenHash)
	if !rtGot.IsRevoked() {
		t.Error("refresh tokens should be revoked after password reset")
	}
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	users := newMockUserRepo()
	resets := newMockPasswordResetRepo()
	user := seedUser(t, users, "proj1", "alice@example.com", "password123")

	prt, raw, _ := domain.NewPasswordResetToken("proj1", user.ID)
	prt.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	_ = resets.Create(context.Background(), "proj1", prt)

	uc := NewResetPassword(users, resets, newMockRefreshTokenRepo(), newMockHasher())
	err := uc.Execute(context.Background(), ResetPasswordInput{
		ProjectID: "proj1", ResetToken: raw, NewPassword: "newpassword99",
	})
	if err != domain.ErrTokenExpired {
		t.Errorf("got %v, want ErrTokenExpired", err)
	}
}

func TestResetPassword_AlreadyUsed(t *testing.T) {
	users := newMockUserRepo()
	resets := newMockPasswordResetRepo()
	user := seedUser(t, users, "proj1", "alice@example.com", "password123")

	prt, raw, _ := domain.NewPasswordResetToken("proj1", user.ID)
	now := time.Now().UTC()
	prt.MarkUsed(now)
	_ = resets.Create(context.Background(), "proj1", prt)

	uc := NewResetPassword(users, resets, newMockRefreshTokenRepo(), newMockHasher())
	err := uc.Execute(context.Background(), ResetPasswordInput{
		ProjectID: "proj1", ResetToken: raw, NewPassword: "newpassword99",
	})
	if err != domain.ErrTokenInvalid {
		t.Errorf("got %v, want ErrTokenInvalid", err)
	}
}

func TestResetPassword_WeakPassword(t *testing.T) {
	users := newMockUserRepo()
	resets := newMockPasswordResetRepo()
	user := seedUser(t, users, "proj1", "alice@example.com", "password123")

	prt, raw, _ := domain.NewPasswordResetToken("proj1", user.ID)
	_ = resets.Create(context.Background(), "proj1", prt)

	uc := NewResetPassword(users, resets, newMockRefreshTokenRepo(), newMockHasher())
	err := uc.Execute(context.Background(), ResetPasswordInput{
		ProjectID: "proj1", ResetToken: raw, NewPassword: "short",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
