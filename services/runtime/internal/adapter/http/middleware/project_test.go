package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type fakeConfigClient struct {
	getConfigFn func(context.Context, string) (*domain.ProjectConfig, error)
}

func (f *fakeConfigClient) GetProject(context.Context, string) (*domain.ProjectMeta, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeConfigClient) GetProjectConfig(ctx context.Context, projectID string) (*domain.ProjectConfig, error) {
	return f.getConfigFn(ctx, projectID)
}

func TestProjectMiddlewareRequiresProjectHeader(t *testing.T) {
	clientCalled := false
	middleware := NewProjectMiddleware(&fakeConfigClient{getConfigFn: func(context.Context, string) (*domain.ProjectConfig, error) {
		clientCalled = true
		return nil, nil
	}})
	nextCalled := false
	handler := middleware.Handler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true }))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "missing X-Project-Id header")
	require.False(t, clientCalled)
	require.False(t, nextCalled)
}

func TestProjectMiddlewareMapsConfigFailureToNotFound(t *testing.T) {
	middleware := NewProjectMiddleware(&fakeConfigClient{getConfigFn: func(_ context.Context, projectID string) (*domain.ProjectConfig, error) {
		require.Equal(t, "missing-project", projectID)
		return nil, errors.New("not found")
	}})
	handler := middleware.Handler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called")
	}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(HeaderProjectID, "missing-project")

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Contains(t, recorder.Body.String(), "project not found or unavailable")
}

func TestProjectMiddlewareAddsResolvedConfigToContext(t *testing.T) {
	want := &domain.ProjectConfig{ProjectID: "project-1", SchemaName: "project_schema"}
	middleware := NewProjectMiddleware(&fakeConfigClient{getConfigFn: func(_ context.Context, projectID string) (*domain.ProjectConfig, error) {
		require.Equal(t, "project-1", projectID)
		return want, nil
	}})
	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := domain.ProjectConfigFromContext(r.Context())
		require.True(t, ok)
		require.Same(t, want, got)
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(HeaderProjectID, "project-1")

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
}
