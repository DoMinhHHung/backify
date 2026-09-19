package usecase

import (
	"context"
	"testing"
	"time"

	"backify/services/auth/internal/domain"
)

func TestSignOut_RevokesRefreshAndBlacklistsAccess(t *testing.T) {
	refresh := newMockRefreshTokenRepo()
	blacklist := newMockBlacklist()
	verifier := newMockVerifier()

	token, raw, err := domain.NewRefreshToken("proj1", "user1", domain.RefreshDuration(domain.RefreshDuration7d))
	if err != nil {
		t.Fatal(err)
	}
	_ = refresh.Create(context.Background(), "proj1", token)

	claims := &domain.AccessClaims{
		Sub: "user1", PID: "proj1", Email: "a@b.com",
		JTI: "jti-1", Iat: time.Now().Unix(), Exp: time.Now().Add(time.Hour).Unix(),
	}
	verifier.verifyFn = func(tok string) (*domain.AccessClaims, error) {
		if tok == "good-access" {
			return claims, nil
		}
		return nil, domain.ErrTokenInvalid
	}

	uc := NewSignOut(refresh, verifier, blacklist)
	if err := uc.Execute(context.Background(), SignOutInput{
		ProjectID:    "proj1",
		AccessToken:  "good-access",
		RefreshToken: raw,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Refresh token phải bị revoke.
	got, err := refresh.GetByTokenHash(context.Background(), "proj1", domain.HashToken(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsRevoked() {
		t.Error("refresh token should be revoked")
	}

	// Access token JTI phải trong blacklist.
	ok, err := blacklist.IsBlacklisted(context.Background(), "jti-1")
	if err != nil || !ok {
		t.Errorf("jti should be blacklisted, ok=%v err=%v", ok, err)
	}
}

func TestSignOut_IgnoresInvalidAccessToken(t *testing.T) {
	refresh := newMockRefreshTokenRepo()
	token, raw, _ := domain.NewRefreshToken("proj1", "user1", domain.RefreshDuration(domain.RefreshDuration7d))
	_ = refresh.Create(context.Background(), "proj1", token)

	verifier := newMockVerifier() // mặc định trả ErrTokenInvalid
	uc := NewSignOut(refresh, verifier, newMockBlacklist())

	// Access token hỏng vẫn signout thành công (chỉ cần revoke refresh).
	if err := uc.Execute(context.Background(), SignOutInput{
		ProjectID:    "proj1",
		AccessToken:  "expired-or-garbage",
		RefreshToken: raw,
	}); err != nil {
		t.Fatalf("should ignore access token errors, got: %v", err)
	}
}

func TestSignOut_RefreshNotFound(t *testing.T) {
	uc := NewSignOut(newMockRefreshTokenRepo(), newMockVerifier(), newMockBlacklist())
	err := uc.Execute(context.Background(), SignOutInput{
		ProjectID:    "proj1",
		AccessToken:  "x",
		RefreshToken: "nonexistent",
	})
	if err != domain.ErrTokenInvalid {
		t.Errorf("got %v, want ErrTokenInvalid", err)
	}
}
