package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"backify/services/auth/internal/domain"
)

func TestVerifyToken_Valid(t *testing.T) {
	verifier := newMockVerifier()
	blacklist := newMockBlacklist()
	claims := &domain.AccessClaims{
		Sub: "user1", PID: "proj1", Email: "a@b.com",
		JTI: "jti-ok", Iat: time.Now().Unix(), Exp: time.Now().Add(time.Hour).Unix(),
	}
	verifier.verifyFn = func(tok string) (*domain.AccessClaims, error) {
		return claims, nil
	}

	uc := NewVerifyToken(verifier, blacklist)
	out, err := uc.Execute(context.Background(), VerifyTokenInput{
		Token: "good", ExpectedProjectID: "proj1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Valid {
		t.Errorf("Valid=false, Error=%q", out.Error)
	}
	if out.UserID != "user1" || out.Email != "a@b.com" {
		t.Errorf("claims mismatch: %+v", out)
	}
}

func TestVerifyToken_InvalidSignature(t *testing.T) {
	verifier := newMockVerifier() // mặc định invalid
	uc := NewVerifyToken(verifier, newMockBlacklist())
	out, err := uc.Execute(context.Background(), VerifyTokenInput{
		Token: "bad", ExpectedProjectID: "proj1",
	})
	if err != nil {
		t.Fatalf("infra error not expected: %v", err)
	}
	if out.Valid {
		t.Error("should be invalid")
	}
	if out.Error == "" {
		t.Error("Error message should be set")
	}
}

func TestVerifyToken_WrongProject(t *testing.T) {
	verifier := newMockVerifier()
	verifier.verifyFn = func(tok string) (*domain.AccessClaims, error) {
		return &domain.AccessClaims{
			Sub: "u", PID: "other-proj", Email: "a@b.com",
			JTI: "j", Exp: time.Now().Add(time.Hour).Unix(),
		}, nil
	}
	uc := NewVerifyToken(verifier, newMockBlacklist())
	out, err := uc.Execute(context.Background(), VerifyTokenInput{
		Token: "x", ExpectedProjectID: "proj1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Valid {
		t.Error("should be invalid for wrong project")
	}
}

func TestVerifyToken_Blacklisted(t *testing.T) {
	verifier := newMockVerifier()
	blacklist := newMockBlacklist()
	claims := &domain.AccessClaims{
		Sub: "user1", PID: "proj1", Email: "a@b.com",
		JTI: "jti-bl", Exp: time.Now().Add(time.Hour).Unix(),
	}
	verifier.verifyFn = func(tok string) (*domain.AccessClaims, error) {
		return claims, nil
	}
	_ = blacklist.Add(context.Background(), "jti-bl", time.Hour)

	uc := NewVerifyToken(verifier, blacklist)
	out, err := uc.Execute(context.Background(), VerifyTokenInput{
		Token: "x", ExpectedProjectID: "proj1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Valid {
		t.Error("blacklisted token should be invalid")
	}
}

func TestVerifyToken_BlacklistInfraError(t *testing.T) {
	verifier := newMockVerifier()
	verifier.verifyFn = func(tok string) (*domain.AccessClaims, error) {
		return &domain.AccessClaims{
			Sub: "u", PID: "proj1", Email: "a@b.com",
			JTI: "j", Exp: time.Now().Add(time.Hour).Unix(),
		}, nil
	}
	// Blacklist trả lỗi hạ tầng.
	bl := &failingBlacklist{err: errors.New("redis down")}
	uc := NewVerifyToken(verifier, bl)
	_, err := uc.Execute(context.Background(), VerifyTokenInput{
		Token: "x", ExpectedProjectID: "proj1",
	})
	if err == nil {
		t.Fatal("expected infra error to propagate")
	}
}

type failingBlacklist struct {
	err error
}

func (f *failingBlacklist) Add(ctx context.Context, jti string, ttl time.Duration) error {
	return f.err
}

func (f *failingBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	return false, f.err
}
