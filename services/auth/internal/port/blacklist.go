package port

import (
	"context"
	"time"
)

// TokenBlacklist lưu JTI của access token đã sign-out để VerifyToken từ chối
// token còn trong TTL (thời gian còn lại tới exp). Triển khai Redis dùng
// SETEX với key auth:blacklist:<jti>.
type TokenBlacklist interface {
	Add(ctx context.Context, jti string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}
