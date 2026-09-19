package usecase

import (
	"context"
	"testing"
	"time"

	"backify/services/auth/internal/domain"
)

func TestRefreshToken_Success(t *testing.T) {
	users := newMockUserRepo()
	refresh := newMockRefreshTokenRepo()
	user := seedUser(t, users, "proj1", "alice@example.com", "password123")

	token, raw, err := domain.NewRefreshToken("proj1", user.ID, domain.RefreshDuration(domain.RefreshDuration7d))
	if err != nil {
		t.Fatal(err)
	}
	_ = refresh.Create(context.Background(), "proj1", token)

	uc := NewRefreshToken(users, refresh, newMockIssuer())
	out, err := uc.Execute(context.Background(), RefreshTokenInput{
		ProjectID:    "proj1",
		RefreshToken: raw,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Tokens.AccessToken == "" || out.Tokens.RefreshToken == "" {
		t.Error("tokens should be issued")
	}
	if out.Tokens.RefreshToken == raw {
		t.Error("new refresh token should differ from old")
	}

	// Token cũ phải bị mark used.
	old, _ := refresh.GetByTokenHash(context.Background(), "proj1", domain.HashToken(raw))
	if !old.IsUsed() {
		t.Error("old token should be marked used")
	}
}

func TestRefreshToken_Expired(t *testing.T) {
	users := newMockUserRepo()
	refresh := newMockRefreshTokenRepo()
	user := seedUser(t, users, "proj1", "alice@example.com", "password123")

	token, raw, _ := domain.NewRefreshToken("proj1", user.ID, domain.RefreshDuration(domain.RefreshDuration7d))
	token.ExpiresAt = time.Now().UTC().Add(-time.Minute) // hết hạn
	_ = refresh.Create(context.Background(), "proj1", token)

	uc := NewRefreshToken(users, refresh, newMockIssuer())
	_, err := uc.Execute(context.Background(), RefreshTokenInput{
		ProjectID: "proj1", RefreshToken: raw,
	})
	if err != domain.ErrTokenExpired {
		t.Errorf("got %v, want ErrTokenExpired", err)
	}
}

func TestRefreshToken_Revoked(t *testing.T) {
	users := newMockUserRepo()
	refresh := newMockRefreshTokenRepo()
	user := seedUser(t, users, "proj1", "alice@example.com", "password123")

	token, raw, _ := domain.NewRefreshToken("proj1", user.ID, domain.RefreshDuration(domain.RefreshDuration7d))
	now := time.Now().UTC()
	token.Revoke(now)
	_ = refresh.Create(context.Background(), "proj1", token)

	uc := NewRefreshToken(users, refresh, newMockIssuer())
	_, err := uc.Execute(context.Background(), RefreshTokenInput{
		ProjectID: "proj1", RefreshToken: raw,
	})
	if err != domain.ErrTokenRevoked {
		t.Errorf("got %v, want ErrTokenRevoked", err)
	}
}

func TestRefreshToken_ReuseDetection_RevokesAll(t *testing.T) {
	users := newMockUserRepo()
	refresh := newMockRefreshTokenRepo()
	user := seedUser(t, users, "proj1", "alice@example.com", "password123")

	// Token đã used (đã rotate trước đó).
	token, raw, _ := domain.NewRefreshToken("proj1", user.ID, domain.RefreshDuration(domain.RefreshDuration7d))
	now := time.Now().UTC()
	token.MarkUsed(now)
	_ = refresh.Create(context.Background(), "proj1", token)

	// Thêm một token khác cùng user còn sống.
	other, _, _ := domain.NewRefreshToken("proj1", user.ID, domain.RefreshDuration(domain.RefreshDuration7d))
	_ = refresh.Create(context.Background(), "proj1", other)

	uc := NewRefreshToken(users, refresh, newMockIssuer())
	_, err := uc.Execute(context.Background(), RefreshTokenInput{
		ProjectID: "proj1", RefreshToken: raw,
	})
	if err != domain.ErrTokenReuseDetected {
		t.Errorf("got %v, want ErrTokenReuseDetected", err)
	}

	// Toàn bộ token của user phải bị revoke.
	got, _ := refresh.GetByTokenHash(context.Background(), "proj1", other.TokenHash)
	if !got.IsRevoked() {
		t.Error("other token of same user should be revoked on reuse detection")
	}
}
