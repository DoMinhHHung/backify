package memory

import (
	"context"
	"sync"
	"time"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type entry struct {
	cfg       *domain.ProjectConfig
	expiresAt time.Time
}

type ConfigCache struct {
	mu   sync.RWMutex
	data map[string]entry
	ttl  time.Duration
}

// NewConfigCache tạo cache trong bộ nhớ với thời hạn ttl cho từng cấu hình.
func NewConfigCache(ttl time.Duration) *ConfigCache {
	return &ConfigCache{
		data: make(map[string]entry),
		ttl:  ttl,
	}
}

// Get trả cấu hình chưa hết hạn theo projectID; entry hết hạn được xem là cache miss.
func (c *ConfigCache) Get(_ context.Context, projectID string) (*domain.ProjectConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.data[projectID]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.cfg, true
}

// Set lưu cấu hình theo ProjectID và đặt thời điểm hết hạn tính từ lúc gọi.
func (c *ConfigCache) Set(_ context.Context, cfg *domain.ProjectConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[cfg.ProjectID] = entry{
		cfg:       cfg,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate xóa cấu hình của projectID khỏi cache nếu có.
func (c *ConfigCache) Invalidate(_ context.Context, projectID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, projectID)
}
