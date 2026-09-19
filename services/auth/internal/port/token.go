package port

import "backify/services/auth/internal/domain"

// TokenIssuer phát hành JWT access token từ AccessClaims.
// *jwt.Issuer (Bước 5) thỏa interface này qua structural typing, không cần wrapper.
type TokenIssuer interface {
	Issue(claims *domain.AccessClaims) (string, error)
}

// TokenVerifier xác minh chữ ký + hạn của JWT access token và trả AccessClaims.
// *jwt.Verifier (Bước 5) thỏa interface này qua structural typing, không cần wrapper.
type TokenVerifier interface {
	Verify(token string) (*domain.AccessClaims, error)
}
