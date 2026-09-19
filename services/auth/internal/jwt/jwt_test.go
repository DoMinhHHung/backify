package jwt

import (
	"testing"
	"time"

	"backify/services/auth/internal/domain"
)

func TestIssueAndVerify(t *testing.T) {
	issuer, err := NewIssuer("test-secret")
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	verifier, err := NewVerifier("test-secret")
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}

	claims, err := domain.NewAccessClaims("user1", "project1", "user@example.com")
	if err != nil {
		t.Fatalf("NewAccessClaims: %v", err)
	}

	tokenString, err := issuer.Issue(claims)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if tokenString == "" {
		t.Fatal("expected non-empty token string")
	}

	got, err := verifier.Verify(tokenString)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if got.Sub != claims.Sub || got.PID != claims.PID || got.Email != claims.Email || got.JTI != claims.JTI {
		t.Fatalf("claims mismatch: got %+v, want %+v", got, claims)
	}
	if got.Iat != claims.Iat || got.Exp != claims.Exp {
		t.Fatalf("iat/exp mismatch: got iat=%d exp=%d, want iat=%d exp=%d", got.Iat, got.Exp, claims.Iat, claims.Exp)
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	issuer, _ := NewIssuer("secret-a")
	verifier, _ := NewVerifier("secret-b")

	claims, _ := domain.NewAccessClaims("user1", "project1", "user@example.com")
	tokenString, _ := issuer.Issue(claims)

	if _, err := verifier.Verify(tokenString); err != domain.ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestVerify_TamperedToken(t *testing.T) {
	issuer, _ := NewIssuer("test-secret")
	verifier, _ := NewVerifier("test-secret")

	claims, _ := domain.NewAccessClaims("user1", "project1", "user@example.com")
	tokenString, _ := issuer.Issue(claims)

	tampered := tokenString[:len(tokenString)-1] + "x"
	if _, err := verifier.Verify(tampered); err != domain.ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid for tampered token, got %v", err)
	}
}

func TestVerify_Expired(t *testing.T) {
	issuer, _ := NewIssuer("test-secret")
	verifier, _ := NewVerifier("test-secret")

	past := time.Now().Add(-2 * time.Hour).Unix()
	expiredClaims := &domain.AccessClaims{
		Sub:   "user1",
		PID:   "project1",
		Email: "user@example.com",
		JTI:   "jti1",
		Iat:   past,
		Exp:   past + 1, // vẫn trước "now", chỉ cần Exp < now để hết hạn
	}

	tokenString, err := issuer.Issue(expiredClaims)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if _, err := verifier.Verify(tokenString); err != domain.ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestNewIssuer_EmptySecret(t *testing.T) {
	if _, err := NewIssuer(""); err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestNewVerifier_EmptySecret(t *testing.T) {
	if _, err := NewVerifier(""); err == nil {
		t.Fatal("expected error for empty secret")
	}
}
