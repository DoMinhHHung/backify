package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

// StructuredLog ghi method, path, status, thời lượng mili giây, request ID,
// project ID và user ID sau khi handler hoàn tất.
func StructuredLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		reqID := middleware.GetReqID(r.Context())
		projectID := ""
		if cfg, ok := domain.ProjectConfigFromContext(r.Context()); ok {
			projectID = cfg.ProjectID
		}
		userID := ""
		if claims, ok := AuthClaimsFromContext(r.Context()); ok {
			userID = claims.UserID
		}

		log.Printf(
			"method=%s path=%s status=%d duration_ms=%d request_id=%s project_id=%s user_id=%s",
			r.Method,
			r.URL.Path,
			ww.Status(),
			time.Since(start).Milliseconds(),
			reqID,
			projectID,
			userID,
		)
	})
}
