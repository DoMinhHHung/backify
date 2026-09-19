// Package jwt ký và verify JWT access token (HS256) từ domain.AccessClaims.
// Tách khỏi domain vì domain chỉ được import stdlib (theo rule của dự án) —
// domain.AccessClaims mô tả DỮ LIỆU claims, package này lo việc ký/verify
// bằng thư viện ngoài (golang-jwt/jwt).
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"backify/services/auth/internal/domain"
)

// claims là dạng golang-jwt-facing của domain.AccessClaims. Nhúng
// jwt.RegisteredClaims để có sẵn sub/jti/iat/exp đúng tên chuẩn JWT (RFC
// 7519); chỉ pid/email là claim riêng của Backify, theo đúng payload
// {sub, pid, email, jti, iat, exp} đã chốt ở BACKIFY.md §7.9.
type claims struct {
	PID   string `json:"pid"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// Issuer ký domain.AccessClaims thành JWT bằng HS256 với JWT_SECRET.
type Issuer struct {
	secret []byte
}

// NewIssuer trả lỗi nếu secret rỗng — JWT_SECRET rỗng nghĩa là chữ ký HS256
// coi như không có, mọi token ký ra đều giả mạo được bằng secret rỗng; phải
// chặn ngay lúc khởi tạo service thay vì để ký "thành công" với secret rỗng.
func NewIssuer(secret string) (*Issuer, error) {
	if secret == "" {
		return nil, errors.New("jwt secret must not be empty")
	}
	return &Issuer{secret: []byte(secret)}, nil
}

// Issue ký domain.AccessClaims (đã có JTI/Iat/Exp từ domain.NewAccessClaims)
// thành JWT string. Không tự sinh JTI/Exp ở đây — đó là việc của domain,
// package jwt chỉ lo mã hoá/ký, giữ đúng ranh giới domain quyết định
// nghiệp vụ (thời hạn, định danh), infra chỉ thực thi.
func (i *Issuer) Issue(c *domain.AccessClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		PID:   c.PID,
		Email: c.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   c.Sub,
			ID:        c.JTI,
			IssuedAt:  jwt.NewNumericDate(time.Unix(c.Iat, 0).UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Unix(c.Exp, 0).UTC()),
		},
	})
	return token.SignedString(i.secret)
}
