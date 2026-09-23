package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

func TestConfigCacheSetGetAndInvalidate(t *testing.T) {
	cache := NewConfigCache(time.Minute)
	cfg := &domain.ProjectConfig{ProjectID: "project-1", Version: 2}

	_, ok := cache.Get(context.Background(), cfg.ProjectID)
	require.False(t, ok)

	cache.Set(context.Background(), cfg)
	got, ok := cache.Get(context.Background(), cfg.ProjectID)
	require.True(t, ok)
	require.Same(t, cfg, got)

	cache.Invalidate(context.Background(), cfg.ProjectID)
	_, ok = cache.Get(context.Background(), cfg.ProjectID)
	require.False(t, ok)
}

func TestConfigCacheDoesNotReturnExpiredEntry(t *testing.T) {
	cache := NewConfigCache(time.Minute)
	cache.data["project-1"] = entry{
		cfg:       &domain.ProjectConfig{ProjectID: "project-1"},
		expiresAt: time.Now().Add(-time.Nanosecond),
	}

	got, ok := cache.Get(context.Background(), "project-1")

	require.False(t, ok)
	require.Nil(t, got)
}

func TestConfigCacheSetReplacesExistingProject(t *testing.T) {
	cache := NewConfigCache(time.Minute)
	cache.Set(context.Background(), &domain.ProjectConfig{ProjectID: "project-1", Version: 1})
	cache.Set(context.Background(), &domain.ProjectConfig{ProjectID: "project-1", Version: 2})

	got, ok := cache.Get(context.Background(), "project-1")

	require.True(t, ok)
	require.Equal(t, int32(2), got.Version)
}
