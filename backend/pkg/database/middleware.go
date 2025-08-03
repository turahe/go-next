package database

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DatabaseMiddleware provides database connection context to HTTP handlers
func DatabaseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add database context to gin context
		c.Set("db", GetDB())

		// Add connection pool stats to context
		c.Set("pool_stats", PoolStats())

		c.Next()
	}
}

// WithDatabaseContext middleware that provides a database connection with timeout
func WithDatabaseContext(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		db, err := GetConnectionWithTimeout(timeout)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Database connection unavailable",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Add database connection to context
		c.Set("db", db)
		c.Set("db_context", ctx)

		c.Next()
	}
}

// HealthCheckMiddleware provides database health check endpoint
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":     "unhealthy",
				"error":      err.Error(),
				"pool_stats": PoolStats(),
			})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"pool_stats": PoolStats(),
		})
	}
}

// TransactionMiddleware provides database transaction context
func TransactionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start a database transaction
		tx := GetDB().Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to start database transaction",
				"details": tx.Error.Error(),
			})
			c.Abort()
			return
		}

		// Add transaction to context
		c.Set("tx", tx)

		// Handle the request
		c.Next()

		// Check if there was an error during request processing
		if len(c.Errors) > 0 {
			// Rollback transaction on error
			tx.Rollback()
		} else {
			// Commit transaction on success
			if err := tx.Commit().Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to commit database transaction",
					"details": err.Error(),
				})
			}
		}
	}
}

// ConnectionPoolMiddleware provides connection pool monitoring
func ConnectionPoolMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Add pool stats to response headers
		stats := PoolStats()
		c.Header("X-DB-Pool-Max-Open", fmt.Sprintf("%d", stats["max_open_connections"]))
		c.Header("X-DB-Pool-Open", fmt.Sprintf("%d", stats["open_connections"]))
		c.Header("X-DB-Pool-In-Use", fmt.Sprintf("%d", stats["in_use"]))
		c.Header("X-DB-Pool-Idle", fmt.Sprintf("%d", stats["idle"]))

		c.Next()

		// Log slow database operations
		duration := time.Since(start)
		if duration > 1*time.Second {
			log.Printf("Slow database operation: %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
		}
	}
}

// RetryMiddleware provides retry logic for database operations
func RetryMiddleware(maxRetries int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var lastErr error

		for i := 0; i < maxRetries; i++ {
			// Process the request
			c.Next()

			// Check if there were any errors
			if len(c.Errors) == 0 {
				return // Success, exit retry loop
			}

			lastErr = c.Errors.Last().Err

			// If it's not a database-related error, don't retry
			if !isDatabaseError(lastErr) {
				return
			}

			// Wait before retry (exponential backoff)
			if i < maxRetries-1 {
				delay := time.Duration(1<<uint(i)) * time.Millisecond * 100
				if delay > 2*time.Second {
					delay = 2 * time.Second
				}
				time.Sleep(delay)
			}
		}

		// If we get here, all retries failed
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database operation failed after retries",
			"details": lastErr.Error(),
		})
	}
}

// isDatabaseError checks if an error is database-related
func isDatabaseError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common database error patterns
	errStr := err.Error()
	return contains(errStr, "connection") ||
		contains(errStr, "timeout") ||
		contains(errStr, "deadlock") ||
		contains(errStr, "lock") ||
		contains(errStr, "database") ||
		contains(errStr, "sql")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

// DatabaseMetricsMiddleware provides database metrics for monitoring
func DatabaseMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record metrics before request
		stats := PoolStats()

		// Add metrics to context for potential export
		c.Set("db_metrics_start", stats)

		c.Next()

		// Record metrics after request
		endStats := PoolStats()
		c.Set("db_metrics_end", endStats)

		// Calculate metrics delta
		delta := calculateMetricsDelta(stats, endStats)
		c.Set("db_metrics_delta", delta)
	}
}

// calculateMetricsDelta calculates the difference between two metric sets
func calculateMetricsDelta(start, end map[string]interface{}) map[string]interface{} {
	delta := make(map[string]interface{})

	// Calculate differences for numeric metrics
	if startWait, ok := start["wait_count"].(int64); ok {
		if endWait, ok := end["wait_count"].(int64); ok {
			delta["wait_count_delta"] = endWait - startWait
		}
	}

	if startInUse, ok := start["in_use"].(int); ok {
		if endInUse, ok := end["in_use"].(int); ok {
			delta["in_use_delta"] = endInUse - startInUse
		}
	}

	return delta
}
