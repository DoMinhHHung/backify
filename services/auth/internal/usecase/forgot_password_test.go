package usecase

import (
	"context"
	"testing"
)

func TestForgotPassword_Success(t *testing.T) {
	users := newMockUserRepo()
	resets := newMockPasswordResetRepo()
	pub := newMockEventPublisher()
	seedUser(t, users, "proj1", "alice@example.com", "password123")

	uc := NewForgotPassword(users, resets, pub)
	if err := uc.Execute(context.Background(), ForgotPasswordInput{
		ProjectID: "proj1",
		Email:     "alice@example.com",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pub.count() != 1 {
		t.Errorf("expected 1 event, got %d", pub.count())
	}
	if pub.lastEventName() != "auth.password_reset_requested" {
		t.Errorf("event = %q", pub.lastEventName())
	}
	if len(resets.tokens) != 1 {
		t.Errorf("expected 1 reset token stored, got %d", len(resets.tokens))
	}
}

func TestForgotPassword_UnknownEmail_NoEnumeration(t *testing.T) {
	resets := newMockPasswordResetRepo()
	pub := newMockEventPublisher()
	uc := NewForgotPassword(newMockUserRepo(), resets, pub)

	// Email không tồn tại → thành công giả, không tạo token, không publish.
	if err := uc.Execute(context.Background(), ForgotPasswordInput{
		ProjectID: "proj1",
		Email:     "ghost@example.com",
	}); err != nil {
		t.Fatalf("should return nil for unknown email, got: %v", err)
	}
	if len(resets.tokens) != 0 {
		t.Error("should not create reset token for unknown email")
	}
	if pub.count() != 0 {
		t.Error("should not publish event for unknown email")
	}
}
