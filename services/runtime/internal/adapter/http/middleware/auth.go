package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

type contextKey int

const authClaimsKey contextKey = 2

type AuthMiddleware struct {
	tokens port.TokenService
}

func NewAuthMiddleware(tokens port.TokenService) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens}
}

// Handler yêu cầu Bearer access token hợp lệ, đối chiếu project trong token với
// project config nếu có, rồi đưa claims vào context cho handler phía sau.
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		claims, err := m.tokens.ParseAccess(token)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		cfg, ok := domain.ProjectConfigFromContext(r.Context())
		if ok && claims.ProjectID != cfg.ProjectID {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), authClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthClaimsFromContext trả claims do AuthMiddleware lưu và báo false nếu không có.
func AuthClaimsFromContext(ctx context.Context) (*domain.AuthClaims, bool) {
	c, ok := ctx.Value(authClaimsKey).(*domain.AuthClaims)
	return c, ok
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + msg + `"}`))
}
