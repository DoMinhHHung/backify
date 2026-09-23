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

type fakeTokenService struct {
	parseAccessFn func(string) (*domain.AuthClaims, error)
}

func (f *fakeTokenService) Issue(string, string, string) (*domain.TokenPair, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeTokenService) ParseAccess(token string) (*domain.AuthClaims, error) {
	return f.parseAccessFn(token)
}

func (f *fakeTokenService) ParseRefresh(string) (*domain.AuthClaims, error) {
	return nil, errors.New("not implemented")
}

func TestAuthMiddlewareRejectsMissingMalformedAndInvalidBearerToken(t *testing.T) {
	parseCalls := 0
	middleware := NewAuthMiddleware(&fakeTokenService{parseAccessFn: func(string) (*domain.AuthClaims, error) {
		parseCalls++
		return nil, errors.New("invalid token")
	}})
	nextCalls := 0
	handler := middleware.Handler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalls++ }))

	for _, header := range []string{"", "Basic abc", "Bearer invalid"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Authorization", header)

		handler.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.JSONEq(t, `{"error":"unauthorized"}`, recorder.Body.String())
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	}
	require.Equal(t, 1, parseCalls)
	require.Zero(t, nextCalls)
}

func TestAuthMiddlewareAddsClaimsToContext(t *testing.T) {
	wantClaims := &domain.AuthClaims{UserID: "user-1", ProjectID: "project-1", Role: "user", TokenType: "access"}
	middleware := NewAuthMiddleware(&fakeTokenService{parseAccessFn: func(token string) (*domain.AuthClaims, error) {
		require.Equal(t, "token-value", token)
		return wantClaims, nil
	}})

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := AuthClaimsFromContext(r.Context())
		require.True(t, ok)
		require.Same(t, wantClaims, claims)
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer token-value")
	request = request.WithContext(domain.WithProjectConfig(request.Context(), &domain.ProjectConfig{ProjectID: "project-1"}))

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestAuthMiddlewareRejectsTokenForAnotherProject(t *testing.T) {
	middleware := NewAuthMiddleware(&fakeTokenService{parseAccessFn: func(string) (*domain.AuthClaims, error) {
		return &domain.AuthClaims{UserID: "user-1", ProjectID: "other-project"}, nil
	}})
	nextCalled := false
	handler := middleware.Handler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true }))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer token")
	request = request.WithContext(domain.WithProjectConfig(request.Context(), &domain.ProjectConfig{ProjectID: "project-1"}))

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.False(t, nextCalled)
}

func TestAuthClaimsFromContextRejectsWrongValueType(t *testing.T) {
	claims, ok := AuthClaimsFromContext(context.WithValue(context.Background(), authClaimsKey, "not claims"))

	require.False(t, ok)
	require.Nil(t, claims)
}
