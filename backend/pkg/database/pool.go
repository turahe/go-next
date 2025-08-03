package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ConnectionPool represents a managed database connection pool
type ConnectionPool struct {
	db          *gorm.DB
	mu          sync.RWMutex
	healthCheck time.Duration
	lastHealth  time.Time
	isHealthy   bool
	errorCount  int
	maxErrors   int
}

// NewConnectionPool creates a new connection pool instance
func NewConnectionPool(db *gorm.DB) *ConnectionPool {
	return &ConnectionPool{
		db:          db,
		healthCheck: 30 * time.Second,
		maxErrors:   5,
		isHealthy:   true,
	}
}

// GetConnection returns a database connection from the pool
func (cp *ConnectionPool) GetConnection(ctx context.Context) (*gorm.DB, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	if !cp.isHealthy {
		return nil, fmt.Errorf("connection pool is unhealthy")
	}

	// Check if we need to perform a health check
	if time.Since(cp.lastHealth) > cp.healthCheck {
		if err := cp.performHealthCheck(ctx); err != nil {
			cp.errorCount++
			if cp.errorCount >= cp.maxErrors {
				cp.isHealthy = false
				return nil, fmt.Errorf("connection pool marked as unhealthy after %d errors", cp.maxErrors)
			}
			return nil, fmt.Errorf("health check failed: %w", err)
		}
		cp.lastHealth = time.Now()
		cp.errorCount = 0
	}

	return cp.db.WithContext(ctx), nil
}

// performHealthCheck performs a health check on the database connection
func (cp *ConnectionPool) performHealthCheck(ctx context.Context) error {
	sqlDB, err := cp.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return sqlDB.PingContext(pingCtx)
}

// IsHealthy returns the current health status of the connection pool
func (cp *ConnectionPool) IsHealthy() bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.isHealthy
}

// ResetHealth resets the health status of the connection pool
func (cp *ConnectionPool) ResetHealth() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.isHealthy = true
	cp.errorCount = 0
}

// SetHealthCheckInterval sets the health check interval
func (cp *ConnectionPool) SetHealthCheckInterval(interval time.Duration) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.healthCheck = interval
}

// ConnectionPoolStats provides detailed statistics about the connection pool
type ConnectionPoolStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
	IsHealthy          bool
	LastHealthCheck    time.Time
	ErrorCount         int
}

// GetPoolStats returns detailed connection pool statistics
func (cp *ConnectionPool) GetPoolStats() ConnectionPoolStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	stats := ConnectionPoolStats{
		IsHealthy:       cp.isHealthy,
		LastHealthCheck: cp.lastHealth,
		ErrorCount:      cp.errorCount,
	}

	if cp.db != nil {
		if sqlDB, err := cp.db.DB(); err == nil {
			dbStats := sqlDB.Stats()
			stats.MaxOpenConnections = dbStats.MaxOpenConnections
			stats.OpenConnections = dbStats.OpenConnections
			stats.InUse = dbStats.InUse
			stats.Idle = dbStats.Idle
			stats.WaitCount = dbStats.WaitCount
			stats.WaitDuration = dbStats.WaitDuration
			stats.MaxIdleClosed = dbStats.MaxIdleClosed
			stats.MaxLifetimeClosed = dbStats.MaxLifetimeClosed
		}
	}

	return stats
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// Calculate backoff delay
		delay := time.Duration(1<<uint(i)) * time.Second
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}
	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// WithConnection executes a function with a database connection, handling retries
func WithConnection(ctx context.Context, fn func(*gorm.DB) error) error {
	return RetryWithBackoff(ctx, 3, func() error {
		db, err := GetConnectionWithTimeout(10 * time.Second)
		if err != nil {
			return err
		}
		return fn(db)
	})
}

// BatchProcessor handles batch database operations with connection pooling
type BatchProcessor struct {
	batchSize int
	timeout   time.Duration
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize int, timeout time.Duration) *BatchProcessor {
	return &BatchProcessor{
		batchSize: batchSize,
		timeout:   timeout,
	}
}

// ProcessBatch processes items in batches with connection pooling
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, items []interface{}, processor func(*gorm.DB, []interface{}) error) error {
	for i := 0; i < len(items); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]

		ctx, cancel := context.WithTimeout(ctx, bp.timeout)
		err := WithConnection(ctx, func(db *gorm.DB) error {
			return processor(db, batch)
		})
		cancel()

		if err != nil {
			return fmt.Errorf("batch processing failed at index %d: %w", i, err)
		}
	}
	return nil
}

// ConnectionMonitor provides monitoring capabilities for database connections
type ConnectionMonitor struct {
	mu             sync.RWMutex
	connections    map[string]time.Time
	maxConnections int
	timeout        time.Duration
}

// NewConnectionMonitor creates a new connection monitor
func NewConnectionMonitor(maxConnections int, timeout time.Duration) *ConnectionMonitor {
	return &ConnectionMonitor{
		connections:    make(map[string]time.Time),
		maxConnections: maxConnections,
		timeout:        timeout,
	}
}

// TrackConnection tracks a database connection
func (cm *ConnectionMonitor) TrackConnection(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.connections) >= cm.maxConnections {
		return fmt.Errorf("maximum connections reached: %d", cm.maxConnections)
	}

	cm.connections[id] = time.Now()
	return nil
}

// ReleaseConnection releases a tracked connection
func (cm *ConnectionMonitor) ReleaseConnection(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.connections, id)
}

// GetActiveConnections returns the number of active connections
func (cm *ConnectionMonitor) GetActiveConnections() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.connections)
}

// CleanupStaleConnections removes stale connections
func (cm *ConnectionMonitor) CleanupStaleConnections() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for id, timestamp := range cm.connections {
		if now.Sub(timestamp) > cm.timeout {
			delete(cm.connections, id)
			log.Printf("Removed stale connection: %s", id)
		}
	}
}

// StartConnectionMonitoring starts monitoring database connections
func StartConnectionMonitoring(ctx context.Context, monitor *ConnectionMonitor) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				monitor.CleanupStaleConnections()
			}
		}
	}()
}
