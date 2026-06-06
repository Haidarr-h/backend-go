package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count    int
	windowAt time.Time
}

var (
	mu       sync.Mutex
	visitors = make(map[string]*visitor)
)

func RateLimit(maxReq int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		v, ok := visitors[ip]

		if !ok || time.Since(v.windowAt) > window {
			visitors[ip] = &visitor{count: 1, windowAt: time.Now()}
			mu.Unlock()
			c.Next()
			return
		}
		mu.Unlock()
		c.Next()
	}
}
