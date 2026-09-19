package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"backify/services/auth/internal/usecase"
)

type contextKey string

const (
	projectIDKey contextKey = "project_id"
	userIDKey    contextKey = "user_id"
	emailKey     contextKey = "email"
)

// RequireProjectID đọc header X-Project-ID; thiếu → 400, có → gắn vào context.
func RequireProjectID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pid := strings.TrimSpace(r.Header.Get("X-Project-ID"))
		if pid == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = jsonEncode(w, envelope{Error: &errorBody{
				Code:    "INVALID_INPUT",
				Message: "X-Project-ID header is required",
			}})
			return
		}
		ctx := context.WithValue(r.Context(), projectIDKey, pid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth chạy SAU RequireProjectID. Đọc Bearer token, gọi VerifyToken
// usecase; hợp lệ → gắn user_id + email vào context.
func RequireAuth(verify *usecase.VerifyToken) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) || strings.TrimSpace(auth[len(prefix):]) == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = jsonEncode(w, envelope{Error: &errorBody{
					Code:    "TOKEN_INVALID",
					Message: "Authorization Bearer token is required",
				}})
				return
			}
			token := strings.TrimSpace(auth[len(prefix):])
			projectID := projectIDFrom(r.Context())

			out, err := verify.Execute(r.Context(), usecase.VerifyTokenInput{
				Token:             token,
				ExpectedProjectID: projectID,
			})
			if err != nil {
				// Lỗi hạ tầng (Redis, …).
				writeError(w, err)
				return
			}
			if !out.Valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				msg := out.Error
				if msg == "" {
					msg = "token is invalid"
				}
				_ = jsonEncode(w, envelope{Error: &errorBody{
					Code:    "TOKEN_INVALID",
					Message: msg,
				}})
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, out.UserID)
			ctx = context.WithValue(ctx, emailKey, out.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func projectIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(projectIDKey).(string)
	return v
}

func userIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

func jsonEncode(w http.ResponseWriter, v interface{}) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}
