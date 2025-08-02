package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLoggingMiddleware logs all incoming requests and responses
func RequestLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Log request
		logger.Info("Request started",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		)

		// Process request
		c.Next()

		// Log response
		duration := time.Since(start)
		status := c.Writer.Status()
		bodySize := c.Writer.Size()

		logger.Info("Request completed",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.Int("body_size", bodySize),
		)
	}
}

// ErrorLoggingMiddleware logs errors that occur during request processing
func ErrorLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("Request error",
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("error", err.Error()),
					zap.String("type", string(err.Type)),
				)
			}
		}
	}
}

// PerformanceMiddleware logs slow requests
func PerformanceMiddleware(logger *zap.Logger, threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		if duration > threshold {
			logger.Warn("Slow request detected",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Duration("duration", duration),
				zap.Duration("threshold", threshold),
			)
		}
	}
}
