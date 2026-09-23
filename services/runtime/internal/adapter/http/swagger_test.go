package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestMountSwaggerServesEmbeddedSpecAndUI(t *testing.T) {
	router := chi.NewRouter()
	MountSwagger(router)

	specRecorder := httptest.NewRecorder()
	router.ServeHTTP(specRecorder, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	require.Equal(t, http.StatusOK, specRecorder.Code)
	require.Equal(t, "application/yaml", specRecorder.Header().Get("Content-Type"))
	require.Contains(t, specRecorder.Body.String(), "openapi:")

	uiRecorder := httptest.NewRecorder()
	router.ServeHTTP(uiRecorder, httptest.NewRequest(http.MethodGet, "/swagger/", nil))
	require.Equal(t, http.StatusOK, uiRecorder.Code)
	require.Contains(t, uiRecorder.Body.String(), "SwaggerUIBundle")
	require.Contains(t, uiRecorder.Body.String(), "url: '/openapi.yaml'")
}

func TestMountSwaggerRedirectsNestedPathToCanonicalUI(t *testing.T) {
	router := chi.NewRouter()
	MountSwagger(router)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/swagger/anything", nil))

	require.Equal(t, http.StatusFound, recorder.Code)
	require.Equal(t, "/swagger/", recorder.Header().Get("Location"))
}
