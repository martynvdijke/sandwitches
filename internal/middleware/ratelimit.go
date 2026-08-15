package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipBucket struct {
	tokens float64
	last   time.Time
}

// RateLimit is a per-IP token-bucket rate limiter. rate is tokens added per
// second, burst is the maximum bucket capacity (and initial allowance). When a
// request exceeds the limit it receives 429 with a Retry-After header and a
// standard error envelope.
func RateLimit(rate, burst float64) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]*ipBucket)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		b, ok := buckets[ip]
		if !ok {
			b = &ipBucket{tokens: burst, last: now}
			buckets[ip] = b
		}
		b.tokens += now.Sub(b.last).Seconds() * rate
		if b.tokens > burst {
			b.tokens = burst
		}
		b.last = now
		if b.tokens < 1 {
			retryAfter := 1
			if rate > 0 {
				retryAfter = int((1 - b.tokens) / rate)
				if retryAfter < 1 {
					retryAfter = 1
				}
			}
			mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message":    "Rate limit exceeded, please retry later",
				"code":       "rate_limited",
				"request_id": GetRequestID(c),
			})
			c.Abort()
			return
		}
		b.tokens--
		mu.Unlock()
		c.Next()
	}
}
