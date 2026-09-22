package grpc

import (
	"context"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

type CachedConfigClient struct {
	inner port.ConfigClient
	cache port.ConfigCache
}

func NewCachedConfigClient(inner port.ConfigClient, cache port.ConfigCache) *CachedConfigClient {
	return &CachedConfigClient{
		inner: inner,
		cache: cache,
	}
}

func (c *CachedConfigClient) GetProject(ctx context.Context, projectID string) (*domain.ProjectMeta, error) {
	return c.inner.GetProject(ctx, projectID)
}

func (c *CachedConfigClient) GetProjectConfig(ctx context.Context, projectID string) (*domain.ProjectConfig, error) {
	if cfg, ok := c.cache.Get(ctx, projectID); ok {
		return cfg, nil
	}

	cfg, err := c.inner.GetProjectConfig(ctx, projectID)
	if err != nil {
		return nil, err
	}

	c.cache.Set(ctx, cfg)
	return cfg, nil
}
