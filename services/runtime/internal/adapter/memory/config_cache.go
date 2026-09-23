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

func NewConfigCache(ttl time.Duration) *ConfigCache {
	return &ConfigCache{
		data: make(map[string]entry),
		ttl:  ttl,
	}
}

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

func (c *ConfigCache) Set(_ context.Context, cfg *domain.ProjectConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[cfg.ProjectID] = entry{
		cfg:       cfg,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *ConfigCache) Invalidate(_ context.Context, projectID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, projectID)
}
