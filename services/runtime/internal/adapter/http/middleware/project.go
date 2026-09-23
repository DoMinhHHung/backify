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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"missing X-Project-Id header"}`))
			return
		}

		cfg, err := m.configClient.GetProjectConfig(r.Context(), projectID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"project not found or unavailable"}`))
			return
		}

		ctx := domain.WithProjectConfig(r.Context(), cfg)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
