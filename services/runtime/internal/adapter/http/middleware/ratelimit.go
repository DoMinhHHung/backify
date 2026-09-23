package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		hits:   make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(HeaderProjectID)
		if key == "" {
			key = r.RemoteAddr
		}
		if !rl.allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)
	arr := rl.hits[key]
	fresh := arr[:0]
	for _, t := range arr {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}
	if len(fresh) >= rl.limit {
		rl.hits[key] = fresh
		return false
	}
	rl.hits[key] = append(fresh, now)
	return true
}
