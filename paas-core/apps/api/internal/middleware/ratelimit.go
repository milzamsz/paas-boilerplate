package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// entry tracks request timestamps for a single client.
type entry struct {
	timestamps []time.Time
	mu         sync.Mutex
}

// RateLimiter stores per-IP rate limit state.
type RateLimiter struct {
	clients sync.Map // map[string]*entry
	limit   int
	window  time.Duration
}

// NewRateLimiter creates a rate limiter with the given limit per window.
// Example: NewRateLimiter(5, 15*time.Minute) = 5 requests per 15 minutes.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limit:  limit,
		window: window,
	}
	// Background cleanup of stale entries every 5 minutes.
	go rl.cleanup()
	return rl
}

// RateLimit returns a Gin middleware that enforces the rate limit.
// Uses client IP as the key. Suitable for auth endpoints.
//
// Goilerplate uses 5 req/15 min for login/register — apply via:
//
//	authLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
//	authGroup.Use(middleware.RateLimit(authLimiter))
func RateLimit(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = c.RemoteIP()
		}

		if !rl.allow(ip) {
			slog.Warn("Rate limit exceeded", "ip", ip, "path", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}

// allow checks whether the given key is within the rate limit.
func (rl *RateLimiter) allow(key string) bool {
	raw, _ := rl.clients.LoadOrStore(key, &entry{})
	e := raw.(*entry)

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove expired timestamps (sliding window).
	valid := e.timestamps[:0]
	for _, t := range e.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	e.timestamps = valid

	if len(e.timestamps) >= rl.limit {
		return false
	}

	e.timestamps = append(e.timestamps, now)
	return true
}

// cleanup periodically removes stale entries from the map.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-rl.window)
		rl.clients.Range(func(key, value any) bool {
			e := value.(*entry)
			e.mu.Lock()
			allExpired := true
			for _, t := range e.timestamps {
				if t.After(cutoff) {
					allExpired = false
					break
				}
			}
			e.mu.Unlock()
			if allExpired {
				rl.clients.Delete(key)
			}
			return true
		})
	}
}
