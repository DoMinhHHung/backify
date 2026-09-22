package middleware

import (
	"net/http"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

const HeaderProjectID = "X-Project-Id"

type ProjectMiddleware struct {
	configClient port.ConfigClient
}

func NewProjectMiddleware(configClient port.ConfigClient) *ProjectMiddleware {
	return &ProjectMiddleware{configClient: configClient}
}

func (m *ProjectMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.Header.Get(HeaderProjectID)
		if projectID == "" {
			http.Error(w, `{"error":"missing X-Project-Id header"}`, http.StatusBadRequest)
			return
		}

		cfg, err := m.configClient.GetProjectConfig(r.Context(), projectID)
		if err != nil {
			http.Error(w, `{"error":"project not found or unavailable"}`, http.StatusNotFound)
			return
		}

		ctx := domain.WithProjectConfig(r.Context(), cfg)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
