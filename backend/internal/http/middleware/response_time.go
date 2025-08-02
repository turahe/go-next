package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ResponseTimeKey = "X-Response-Time"
)

// ResponseTimeMiddleware adds response time header to responses
func ResponseTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		c.Header(ResponseTimeKey, duration.String())
	}
}
