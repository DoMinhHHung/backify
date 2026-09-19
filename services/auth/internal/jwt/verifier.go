package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"

	"backify/services/auth/internal/domain"
)

// Verifier verify JWT access token bằng cùng JWT_SECRET đã dùng để ký ở
// Issuer. Tách riêng khỏi Issuer (dù chung secret) vì đúng theo cấu trúc đã
// chốt (jwt/issuer.go + jwt/verifier.go) — thực tế hai việc ký và verify
// dùng chung key với HS256 nên đơn giản hơn RS256, nhưng vẫn giữ tách file.
type Verifier struct {
	secret []byte
}

// NewVerifier trả lỗi nếu secret rỗng, cùng lý do với NewIssuer.
func NewVerifier(secret string) (*Verifier, error) {
	if secret == "" {
		return nil, errors.New("jwt secret must not be empty")
	}
	return &Verifier{secret: []byte(secret)}, nil
}

// Verify kiểm tra chữ ký + hạn dùng, trả về domain.AccessClaims nếu hợp lệ.
// WithValidMethods chỉ chấp nhận HS256 — chặn algorithm confusion attack
// (kẻ tấn công đổi header "alg" sang "none" hoặc RS256 với public key làm
// HMAC secret để né việc verify chữ ký thật).
func (v *Verifier) Verify(tokenString string) (*domain.AccessClaims, error) {
	var c claims
	_, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (interface{}, error) {
		return v.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrTokenInvalid
	}

	var iat, exp int64
	if c.IssuedAt != nil {
		iat = c.IssuedAt.Unix()
	}
	if c.ExpiresAt != nil {
		exp = c.ExpiresAt.Unix()
	}

	return &domain.AccessClaims{
		Sub:   c.Subject,
		PID:   c.PID,
		Email: c.Email,
		JTI:   c.ID,
		Iat:   iat,
		Exp:   exp,
	}, nil
}
