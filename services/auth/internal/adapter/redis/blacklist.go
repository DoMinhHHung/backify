package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"backify/services/auth/internal/port"
)

const blacklistKeyPrefix = "auth:blacklist:"

// Blacklist implement port.TokenBlacklist bằng Redis SETEX.
// Key: auth:blacklist:<jti>, TTL = thời gian còn lại tới exp của access token.
type Blacklist struct {
	client *goredis.Client
}

// NewBlacklist tạo adapter blacklist gắn với Redis client đã mở sẵn.
func NewBlacklist(client *goredis.Client) *Blacklist {
	return &Blacklist{client: client}
}

// Add ghi JTI vào blacklist với TTL cho trước. No-op (không lỗi) nếu ttl <= 0
// vì token đã hết hạn — không cần blacklist thêm.
func (b *Blacklist) Add(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	key := blacklistKeyPrefix + jti
	return b.client.SetEx(ctx, key, "1", ttl).Err()
}

// IsBlacklisted trả true nếu JTI còn trong Redis (token đã sign-out và chưa hết TTL).
func (b *Blacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := blacklistKeyPrefix + jti
	n, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("blacklist exists: %w", err)
	}
	return n > 0, nil
}

// Đảm bảo Blacklist thỏa port.TokenBlacklist lúc compile.
var _ port.TokenBlacklist = (*Blacklist)(nil)
