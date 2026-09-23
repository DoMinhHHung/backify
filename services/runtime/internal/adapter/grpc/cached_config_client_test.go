package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type fakeConfigClient struct {
	getProjectCalls       int
	getProjectConfigCalls int
	project               *domain.ProjectMeta
	config                *domain.ProjectConfig
	err                   error
}

func (f *fakeConfigClient) GetProject(context.Context, string) (*domain.ProjectMeta, error) {
	f.getProjectCalls++
	return f.project, f.err
}

func (f *fakeConfigClient) GetProjectConfig(context.Context, string) (*domain.ProjectConfig, error) {
	f.getProjectConfigCalls++
	return f.config, f.err
}

type fakeConfigCache struct {
	values   map[string]*domain.ProjectConfig
	setCalls int
}

func (f *fakeConfigCache) Get(_ context.Context, projectID string) (*domain.ProjectConfig, bool) {
	cfg, ok := f.values[projectID]
	return cfg, ok
}

func (f *fakeConfigCache) Set(_ context.Context, cfg *domain.ProjectConfig) {
	f.setCalls++
	f.values[cfg.ProjectID] = cfg
}

func (f *fakeConfigCache) Invalidate(_ context.Context, projectID string) {
	delete(f.values, projectID)
}

func TestCachedConfigClientReturnsCacheHitWithoutCallingInner(t *testing.T) {
	cached := &domain.ProjectConfig{ProjectID: "project-1", Version: 1}
	inner := &fakeConfigClient{config: &domain.ProjectConfig{ProjectID: "project-1", Version: 2}}
	cache := &fakeConfigCache{values: map[string]*domain.ProjectConfig{"project-1": cached}}

	got, err := NewCachedConfigClient(inner, cache).GetProjectConfig(context.Background(), "project-1")

	require.NoError(t, err)
	require.Same(t, cached, got)
	require.Zero(t, inner.getProjectConfigCalls)
	require.Zero(t, cache.setCalls)
}

func TestCachedConfigClientFetchesAndCachesMiss(t *testing.T) {
	fetched := &domain.ProjectConfig{ProjectID: "project-1", Version: 2}
	inner := &fakeConfigClient{config: fetched}
	cache := &fakeConfigCache{values: make(map[string]*domain.ProjectConfig)}
	client := NewCachedConfigClient(inner, cache)

	first, err := client.GetProjectConfig(context.Background(), "project-1")
	require.NoError(t, err)
	second, err := client.GetProjectConfig(context.Background(), "project-1")

	require.NoError(t, err)
	require.Same(t, fetched, first)
	require.Same(t, fetched, second)
	require.Equal(t, 1, inner.getProjectConfigCalls)
	require.Equal(t, 1, cache.setCalls)
}

func TestCachedConfigClientDoesNotCacheInnerFailure(t *testing.T) {
	wantErr := errors.New("control plane unavailable")
	inner := &fakeConfigClient{err: wantErr}
	cache := &fakeConfigCache{values: make(map[string]*domain.ProjectConfig)}

	got, err := NewCachedConfigClient(inner, cache).GetProjectConfig(context.Background(), "project-1")

	require.Nil(t, got)
	require.ErrorIs(t, err, wantErr)
	require.Zero(t, cache.setCalls)
}

func TestCachedConfigClientAlwaysDelegatesProjectMetadata(t *testing.T) {
	want := &domain.ProjectMeta{ID: "project-1"}
	inner := &fakeConfigClient{project: want}
	cache := &fakeConfigCache{values: make(map[string]*domain.ProjectConfig)}

	got, err := NewCachedConfigClient(inner, cache).GetProject(context.Background(), "project-1")

	require.NoError(t, err)
	require.Same(t, want, got)
	require.Equal(t, 1, inner.getProjectCalls)
}
