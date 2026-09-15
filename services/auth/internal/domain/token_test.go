package domain

import (
	"testing"
	"time"
)

func TestNewAccessClaims(t *testing.T) {
	c, err := NewAccessClaims("user1", "proj1", "a@b.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Sub != "user1" || c.PID != "proj1" || c.Email != "a@b.com" {
		t.Fatalf("unexpected claims: %+v", c)
	}
	if c.JTI == "" {
		t.Fatal("expected jti")
	}
	if c.Exp <= c.Iat {
		t.Fatal("exp must be after iat")
	}
	if c.Expired(time.Unix(c.Iat, 0)) {
		t.Fatal("should not be expired at iat")
	}
	if !c.Expired(time.Unix(c.Exp, 0)) {
		t.Fatal("should be expired at exp")
	}
	if !c.BelongsToProject("proj1") {
		t.Fatal("expected belongs to proj1")
	}
	if c.BelongsToProject("other") {
		t.Fatal("must not belong to other project")
	}
}

func TestParseRefreshDuration(t *testing.T) {
	d, err := ParseRefreshDuration("7d")
	if err != nil || time.Duration(d) != RefreshDuration7d {
		t.Fatalf("expected 7d, got %v %v", d, err)
	}
	d, err = ParseRefreshDuration("30d")
	if err != nil || time.Duration(d) != RefreshDuration30d {
		t.Fatalf("expected 30d, got %v %v", d, err)
	}
	if _, err := ParseRefreshDuration("14d"); err == nil {
		t.Fatal("expected error for 14d")
	}
}

func TestNewRefreshToken_AndRedeem(t *testing.T) {
	token, raw, err := NewRefreshToken("proj1", "user1", RefreshDuration(RefreshDuration7d))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw == "" || token.TokenHash == "" {
		t.Fatal("expected raw token and hash")
	}
	if token.TokenHash != HashToken(raw) {
		t.Fatal("hash mismatch")
	}
	if token.FamilyID == "" {
		t.Fatal("expected family id")
	}

	now := time.Now().UTC()
	if err := token.CanRedeem(now); err != nil {
		t.Fatalf("should be redeemable: %v", err)
	}

	token.MarkUsed(now)
	if err := token.CanRedeem(now); err != ErrTokenReuseDetected {
		t.Fatalf("expected reuse detection, got %v", err)
	}
}

func TestRefreshToken_RotateKeepsFamily(t *testing.T) {
	first, _, err := NewRefreshToken("proj1", "user1", RefreshDuration(RefreshDuration7d))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	next, raw, err := first.Rotate(RefreshDuration(RefreshDuration30d))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw == "" {
		t.Fatal("expected raw token")
	}
	if next.FamilyID != first.FamilyID {
		t.Fatalf("family id must be preserved: %s vs %s", next.FamilyID, first.FamilyID)
	}
	if next.UserID != first.UserID || next.ProjectID != first.ProjectID {
		t.Fatal("user/project must be preserved")
	}
	if next.ID == first.ID {
		t.Fatal("rotated token must have new id")
	}
}

func TestRefreshToken_ExpiredAndRevoked(t *testing.T) {
	token, _, _ := NewRefreshToken("proj1", "user1", RefreshDuration(RefreshDuration7d))
	token.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	if err := token.CanRedeem(time.Now().UTC()); err != ErrTokenExpired {
		t.Fatalf("expected expired, got %v", err)
	}

	token2, _, _ := NewRefreshToken("proj1", "user1", RefreshDuration(RefreshDuration7d))
	token2.Revoke(time.Now().UTC())
	if err := token2.CanRedeem(time.Now().UTC()); err != ErrTokenRevoked {
		t.Fatalf("expected revoked, got %v", err)
	}
}
