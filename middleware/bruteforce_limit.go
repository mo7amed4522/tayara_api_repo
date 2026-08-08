package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// In-memory fixed-window IP rate limiter used to protect the unauthenticated
// auth endpoints (login / register / firebase-login) against brute-force and
// credential-stuffing. It deliberately does NOT depend on Redis so the guard is
// always active, even when use-redis is false. A separate, more permissive limit
// applies to the remaining public surface via DefaultLimit()/LimitConfig.
//
// Each unique client IP owns a window; when the window is exhausted the request
// is rejected with HTTP 429. Windows are swept periodically to bound memory.

type ipWindow struct {
	count int
	start time.Time
}

type ipLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	clients  map[string]*ipWindow
	stop     chan struct{}
	stopOnce sync.Once
}

func newIPLimiter(limit int, window time.Duration) *ipLimiter {
	l := &ipLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]*ipWindow),
		stop:    make(chan struct{}),
	}
	go l.sweep()
	return l
}

// Allow reports whether an action from ip may proceed. A false return means the
// caller exhausted its window and should be rejected.
func (l *ipLimiter) Allow(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.clients[ip]
	if !ok || now.Sub(w.start) >= l.window {
		l.clients[ip] = &ipWindow{count: 1, start: now}
		return true
	}
	if w.count >= l.limit {
		return false
	}
	w.count++
	return true
}

// sweep periodically drops expired windows so the map does not grow unbounded.
func (l *ipLimiter) sweep() {
	t := time.NewTicker(l.window)
	defer t.Stop()
	for {
		select {
		case <-l.stop:
			return
		case <-t.C:
			cutoff := time.Now().Add(-l.window)
			l.mu.Lock()
			for ip, w := range l.clients {
				if w.start.Before(cutoff) {
					delete(l.clients, ip)
				}
			}
			l.mu.Unlock()
		}
	}
}

// Close stops the background sweeper.
func (l *ipLimiter) Close() {
	l.stopOnce.Do(func() { close(l.stop) })
}

const (
	// authLimit is the number of allowed auth attempts per IP within authWindow.
	authLimit  = 10
	authWindow = time.Minute
)

var authLimiter = newIPLimiter(authLimit, authWindow)

// AuthRateLimit rate-limits only the unauthenticated credential endpoints
// (login, register, firebase-login) by client IP. All other routes pass through
// untouched, so catalog browsing and public reads are not throttled.
func AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAuthPath(c.Request.URL.Path) {
			if !authLimiter.Allow(c.ClientIP()) {
				c.AbortWithStatusJSON(429, gin.H{
					"code": 429,
					"msg":  "too many authentication attempts, please retry later",
				})
				return
			}
		}
		c.Next()
	}
}

// isAuthPath reports whether the request targets a credential endpoint.
func isAuthPath(path string) bool {
	for _, s := range []string{"/login", "/register", "/firebase-login"} {
		if strings.HasSuffix(path, s) {
			return true
		}
	}
	return false
}
