package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

// supportLimiter is a per-identity fixed-window limiter. It is deliberately in-process: the
// support chat budget is a UX guard, not a security boundary, and it must never make a request
// fail because Redis is down.
type supportLimiter struct {
	mu      sync.Mutex
	buckets map[string]*window
}

type window struct {
	count int
	reset time.Time
}

var supportLimits = &supportLimiter{buckets: map[string]*window{}}

// SupportRateLimit limits an authenticated actor to `limit` requests per window per bucket.
func SupportRateLimit(bucket string, limit int, duration time.Duration) gin.HandlerFunc {
	if limit <= 0 || duration <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		actor := strconv.FormatUint(uint64(utils.GetBxUserID(c)), 10)
		if actor == "0" {
			actor = "s" + strconv.FormatUint(uint64(utils.GetUserID(c)), 10)
		}
		if actor == "s0" {
			// Unauthenticated callers share no identity, so fall back to the client IP instead of
			// letting every anonymous request exhaust one global bucket.
			actor = "ip:" + c.ClientIP()
		}
		key := bucket + ":" + actor
		if !supportLimits.allow(key, limit, duration) {
			response.FailWithError(response.WithCode(429, "too many requests"), c)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (l *supportLimiter) allow(key string, limit int, duration time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	// Opportunistic cleanup keeps the map bounded without a goroutine.
	if len(l.buckets) > 4096 {
		for k, w := range l.buckets {
			if now.After(w.reset) {
				delete(l.buckets, k)
			}
		}
	}
	w, ok := l.buckets[key]
	if !ok || now.After(w.reset) {
		l.buckets[key] = &window{count: 1, reset: now.Add(duration)}
		return true
	}
	if w.count >= limit {
		return false
	}
	w.count++
	return true
}
