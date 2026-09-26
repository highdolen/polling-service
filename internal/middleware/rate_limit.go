package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]visitor
	limit      int
	windowSize time.Duration
}

func NewRateLimiter(
	limit int,
	windowSize time.Duration,
) *RateLimiter {
	return &RateLimiter{
		visitors:   make(map[string]visitor),
		limit:      limit,
		windowSize: windowSize,
	}
}

func (l *RateLimiter) Middleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		ip := clientIP(r)

		if !l.allow(ip) {
			http.Error(
				w,
				"rate limit exceeded",
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) allow(ip string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	current, ok := l.visitors[ip]

	if !ok || now.Sub(current.windowStart) >= l.windowSize {
		l.visitors[ip] = visitor{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if current.count >= l.limit {
		return false
	}

	current.count++
	l.visitors[ip] = current

	return true
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
