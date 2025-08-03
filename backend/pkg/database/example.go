package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ExampleUsage demonstrates how to use the connection pooling features
func ExampleUsage() {
	// Example 1: Basic connection pool usage
	exampleBasicUsage()

	// Example 2: Using connection pool with health checks
	exampleHealthCheckUsage()

	// Example 3: Using batch processing with connection pooling
	exampleBatchProcessing()

	// Example 4: Using retry logic with connection pooling
	exampleRetryLogic()

	// Example 5: Using transaction with connection pooling
	exampleTransactionUsage()
}

func exampleBasicUsage() {
	// Get a database connection with timeout
	db, err := GetConnectionWithTimeout(5 * time.Second)
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return
	}

	// Use the connection
	var result []map[string]interface{}
	if err := db.Raw("SELECT 1 as test").Scan(&result).Error; err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	log.Printf("Query result: %v", result)
}

func exampleHealthCheckUsage() {
	// Perform a health check
	if err := HealthCheck(); err != nil {
		log.Printf("Database health check failed: %v", err)
		return
	}

	log.Println("Database is healthy")

	// Get connection pool statistics
	stats := PoolStats()
	log.Printf("Connection pool stats: %+v", stats)
}

func exampleBatchProcessing() {
	// Create a batch processor
	batchProcessor := NewBatchProcessor(100, 30*time.Second)

	// Example items to process
	items := []interface{}{
		"item1", "item2", "item3", "item4", "item5",
	}

	// Process items in batches
	err := batchProcessor.ProcessBatch(context.Background(), items, func(db *gorm.DB, batch []interface{}) error {
		// Process each batch
		for _, item := range batch {
			log.Printf("Processing item: %v", item)
			// Your database operation here
		}
		return nil
	})

	if err != nil {
		log.Printf("Batch processing failed: %v", err)
		return
	}

	log.Println("Batch processing completed successfully")
}

func exampleRetryLogic() {
	// Use retry logic with connection pooling
	err := WithConnection(context.Background(), func(db *gorm.DB) error {
		// Your database operation here
		var result []map[string]interface{}
		return db.Raw("SELECT 1 as test").Scan(&result).Error
	})

	if err != nil {
		log.Printf("Operation failed after retries: %v", err)
		return
	}

	log.Println("Operation completed successfully")
}

func exampleTransactionUsage() {
	// Use transaction with connection pooling
	err := WithTransaction(func(tx *gorm.DB) error {
		// Your transaction operations here
		if err := tx.Exec("INSERT INTO test_table (name) VALUES (?)", "test").Error; err != nil {
			return err
		}

		if err := tx.Exec("UPDATE test_table SET name = ? WHERE name = ?", "updated", "test").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return
	}

	log.Println("Transaction completed successfully")
}

// ExampleConnectionPool demonstrates advanced connection pool usage
func ExampleConnectionPool() {
	// Create a new connection pool
	pool := NewConnectionPool(GetDB())

	// Set health check interval
	pool.SetHealthCheckInterval(15 * time.Second)

	// Use the connection pool
	ctx := context.Background()
	db, err := pool.GetConnection(ctx)
	if err != nil {
		log.Printf("Failed to get connection from pool: %v", err)
		return
	}

	// Use the connection
	var result []map[string]interface{}
	if err := db.Raw("SELECT 1 as test").Scan(&result).Error; err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	// Check pool health
	if !pool.IsHealthy() {
		log.Println("Connection pool is unhealthy")
		pool.ResetHealth()
	}

	// Get detailed pool statistics
	stats := pool.GetPoolStats()
	log.Printf("Detailed pool stats: %+v", stats)
}

// ExampleConnectionMonitor demonstrates connection monitoring
func ExampleConnectionMonitor() {
	// Create a connection monitor
	monitor := NewConnectionMonitor(100, 5*time.Minute)

	// Start monitoring
	ctx := context.Background()
	StartConnectionMonitoring(ctx, monitor)

	// Track connections
	connectionID := "conn-1"
	if err := monitor.TrackConnection(connectionID); err != nil {
		log.Printf("Failed to track connection: %v", err)
		return
	}

	log.Printf("Active connections: %d", monitor.GetActiveConnections())

	// Release connection
	monitor.ReleaseConnection(connectionID)
}

// ExampleWithCustomConfiguration demonstrates using custom connection pool configuration
func ExampleWithCustomConfiguration() {
	// This would typically be done through environment variables
	// but here we show the concept

	// Custom connection pool settings
	customSettings := map[string]interface{}{
		"max_idle_conns":     20,
		"max_open_conns":     200,
		"conn_max_lifetime":  2 * time.Hour,
		"conn_max_idle_time": 45 * time.Minute,
	}

	log.Printf("Using custom connection pool settings: %+v", customSettings)

	// Use the database with custom settings
	_, err := GetConnectionWithTimeout(10 * time.Second)
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return
	}

	// Get current pool statistics
	stats := PoolStats()
	log.Printf("Current pool statistics: %+v", stats)
}

// ExampleErrorHandling demonstrates error handling with connection pooling
func ExampleErrorHandling() {
	// Example of handling database errors with retry logic
	err := RetryWithBackoff(context.Background(), 3, func() error {
		db, err := GetConnectionWithTimeout(5 * time.Second)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		// Simulate a database operation that might fail
		var result []map[string]interface{}
		if err := db.Raw("SELECT 1 as test").Scan(&result).Error; err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("Operation failed after retries: %v", err)
		return
	}

	log.Println("Operation completed successfully")
}

// ExamplePerformanceMonitoring demonstrates performance monitoring
func ExamplePerformanceMonitoring() {
	// Monitor database performance
	start := time.Now()

	// Perform database operations
	err := WithConnection(context.Background(), func(db *gorm.DB) error {
		// Simulate multiple database operations
		for i := 0; i < 10; i++ {
			var result []map[string]interface{}
			if err := db.Raw("SELECT ? as iteration", i).Scan(&result).Error; err != nil {
				return err
			}
		}
		return nil
	})

	duration := time.Since(start)

	if err != nil {
		log.Printf("Performance test failed: %v", err)
		return
	}

	log.Printf("Performance test completed in %v", duration)

	// Get performance statistics
	stats := PoolStats()
	log.Printf("Performance stats: %+v", stats)
}
