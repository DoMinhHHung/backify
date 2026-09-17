package domain

import (
	"testing"
	"time"
)

func TestNewPasswordResetToken(t *testing.T) {
	token, raw, err := NewPasswordResetToken("proj1", "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw == "" || token.TokenHash == "" {
		t.Fatal("expected raw token and hash")
	}
	if token.TokenHash != HashToken(raw) {
		t.Fatal("hash mismatch")
	}
	if token.ExpiresAt.Sub(token.CreatedAt) != PasswordResetTokenTTL {
		t.Fatalf("expected 1h TTL, got %v", token.ExpiresAt.Sub(token.CreatedAt))
	}

	now := time.Now().UTC()
	if err := token.CanRedeem(now); err != nil {
		t.Fatalf("should be redeemable: %v", err)
	}
}

func TestNewPasswordResetToken_MissingFields(t *testing.T) {
	if _, _, err := NewPasswordResetToken("", "user1"); err == nil {
		t.Fatal("expected error for missing project_id")
	}
	if _, _, err := NewPasswordResetToken("proj1", ""); err == nil {
		t.Fatal("expected error for missing user_id")
	}
}

func TestPasswordResetToken_UsedCannotRedeem(t *testing.T) {
	token, _, _ := NewPasswordResetToken("proj1", "user1")
	token.MarkUsed(time.Now().UTC())
	if err := token.CanRedeem(time.Now().UTC()); err != ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid for used token, got %v", err)
	}
}

func TestPasswordResetToken_Expired(t *testing.T) {
	token, _, _ := NewPasswordResetToken("proj1", "user1")
	token.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	if err := token.CanRedeem(time.Now().UTC()); err != ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}
