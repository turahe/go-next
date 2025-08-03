# Database Connection Pooling

This package provides comprehensive database connection pooling functionality with monitoring, health checks, and performance optimization features.

## Features

- **Configurable Connection Pooling**: Set maximum idle connections, open connections, and connection lifetimes
- **Health Monitoring**: Automatic health checks and connection pool statistics
- **Retry Logic**: Exponential backoff retry mechanism for failed operations
- **Batch Processing**: Efficient batch operations with connection pooling
- **Transaction Support**: Transaction management with connection pooling
- **HTTP Middleware**: Gin middleware for database operations
- **Performance Monitoring**: Real-time connection pool statistics and metrics

## Configuration

### Environment Variables

Add these environment variables to your `.env` file:

```env
# Database Connection Pool Settings
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=30m

# Database Connection Settings
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=wordpress_user
DB_PASSWORD=wordpress_password
DB_NAME=go_next
DB_LOGMODE=true
DB_SSLMODE=require
```

### Configuration Structure

The connection pooling settings are defined in the `DatabaseConfig` struct:

```go
type DatabaseConfig struct {
    Driver   string
    Host     string
    Port     string
    Username string
    Password string
    Dbname   string
    Logmode  bool
    Sslmode  bool
    // Connection pooling configuration
    MaxIdleConns    int
    MaxOpenConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}
```

## Usage

### Basic Connection Pooling

```go
package main

import (
    "context"
    "log"
    "time"
    
    "go-next/pkg/database"
)

func main() {
    // Setup database connection
    if err := database.Setup(); err != nil {
        log.Fatal("Failed to setup database:", err)
    }

    // Get a connection with timeout
    db, err := database.GetConnectionWithTimeout(10 * time.Second)
    if err != nil {
        log.Fatal("Failed to get database connection:", err)
    }

    // Use the connection
    var result []map[string]interface{}
    if err := db.Raw("SELECT 1 as test").Scan(&result).Error; err != nil {
        log.Fatal("Query failed:", err)
    }

    log.Printf("Query result: %v", result)
}
```

### Health Checks

```go
// Perform a health check
if err := database.HealthCheck(); err != nil {
    log.Printf("Database health check failed: %v", err)
    return
}

// Get connection pool statistics
stats := database.PoolStats()
log.Printf("Connection pool stats: %+v", stats)
```

### Transaction Support

```go
// Use transaction with connection pooling
err := database.WithTransaction(func(tx *gorm.DB) error {
    // Your transaction operations here
    if err := tx.Exec("INSERT INTO users (name) VALUES (?)", "John").Error; err != nil {
        return err
    }

    if err := tx.Exec("UPDATE users SET name = ? WHERE name = ?", "Jane", "John").Error; err != nil {
        return err
    }

    return nil
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
    return
}
```

### Retry Logic

```go
// Use retry logic with connection pooling
err := database.WithConnection(context.Background(), func(db *gorm.DB) error {
    // Your database operation here
    var result []map[string]interface{}
    return db.Raw("SELECT 1 as test").Scan(&result).Error
})

if err != nil {
    log.Printf("Operation failed after retries: %v", err)
    return
}
```

### Batch Processing

```go
// Create a batch processor
batchProcessor := database.NewBatchProcessor(100, 30*time.Second)

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
```

## HTTP Middleware

### Basic Database Middleware

```go
import (
    "github.com/gin-gonic/gin"
    "go-next/pkg/database"
)

func setupRoutes(r *gin.Engine) {
    // Add database middleware to all routes
    r.Use(database.DatabaseMiddleware())
    
    // Your routes here
    r.GET("/users", func(c *gin.Context) {
        db := c.MustGet("db").(*gorm.DB)
        // Use db for database operations
    })
}
```

### Database Context with Timeout

```go
// Add database context with timeout to specific routes
r.GET("/users", database.WithDatabaseContext(10*time.Second), func(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)
    ctx := c.MustGet("db_context").(context.Context)
    
    // Use db with context for database operations
})
```

### Health Check Endpoint

```go
// Add health check endpoint
r.GET("/health", database.HealthCheckMiddleware())
```

### Transaction Middleware

```go
// Add transaction middleware to specific routes
r.POST("/users", database.TransactionMiddleware(), func(c *gin.Context) {
    tx := c.MustGet("tx").(*gorm.DB)
    
    // Use tx for transaction operations
    // Transaction will be automatically committed on success or rolled back on error
})
```

### Connection Pool Monitoring

```go
// Add connection pool monitoring to all routes
r.Use(database.ConnectionPoolMiddleware())

// This middleware adds pool statistics to response headers:
// X-DB-Pool-Max-Open: 100
// X-DB-Pool-Open: 5
// X-DB-Pool-In-Use: 2
// X-DB-Pool-Idle: 3
```

### Retry Middleware

```go
// Add retry middleware to specific routes
r.GET("/users", database.RetryMiddleware(3), func(c *gin.Context) {
    // This route will retry up to 3 times on database errors
})
```

## Advanced Usage

### Custom Connection Pool

```go
// Create a custom connection pool
pool := database.NewConnectionPool(database.GetDB())

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
```

### Connection Monitoring

```go
// Create a connection monitor
monitor := database.NewConnectionMonitor(100, 5*time.Minute)

// Start monitoring
ctx := context.Background()
database.StartConnectionMonitoring(ctx, monitor)

// Track connections
connectionID := "conn-1"
if err := monitor.TrackConnection(connectionID); err != nil {
    log.Printf("Failed to track connection: %v", err)
    return
}

log.Printf("Active connections: %d", monitor.GetActiveConnections())

// Release connection
monitor.ReleaseConnection(connectionID)
```

## Performance Optimization

### Connection Pool Statistics

The connection pool provides detailed statistics:

```go
stats := database.PoolStats()
// Returns:
// {
//   "max_open_connections": 100,
//   "open_connections": 5,
//   "in_use": 2,
//   "idle": 3,
//   "wait_count": 0,
//   "wait_duration": 0,
//   "max_idle_closed": 10,
//   "max_lifetime_closed": 5
// }
```

### Performance Monitoring

```go
// Monitor database performance
start := time.Now()

// Perform database operations
err := database.WithConnection(context.Background(), func(db *gorm.DB) error {
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
stats := database.PoolStats()
log.Printf("Performance stats: %+v", stats)
```

## Best Practices

### 1. Connection Pool Sizing

- **MaxIdleConns**: Set to 10-20% of MaxOpenConns
- **MaxOpenConns**: Set based on your application's concurrency needs
- **ConnMaxLifetime**: Set to 1 hour for most applications
- **ConnMaxIdleTime**: Set to 30 minutes for most applications

### 2. Health Checks

- Perform regular health checks to ensure database connectivity
- Monitor connection pool statistics for performance issues
- Set up alerts for connection pool exhaustion

### 3. Error Handling

- Use retry logic for transient database errors
- Implement proper error handling for connection failures
- Monitor and log database errors for debugging

### 4. Performance Monitoring

- Monitor connection pool statistics regularly
- Set up alerts for slow database operations
- Use connection pool monitoring middleware in production

### 5. Transaction Management

- Use transactions for operations that require atomicity
- Keep transactions as short as possible
- Handle transaction rollbacks properly

## Troubleshooting

### Common Issues

1. **Connection Pool Exhaustion**
   - Increase MaxOpenConns
   - Check for connection leaks
   - Monitor connection usage patterns

2. **Slow Database Operations**
   - Check connection pool statistics
   - Monitor query performance
   - Consider query optimization

3. **Health Check Failures**
   - Verify database connectivity
   - Check network connectivity
   - Review database server logs

### Monitoring

The connection pooling system provides comprehensive monitoring capabilities:

- Real-time connection pool statistics
- Health check status
- Performance metrics
- Error tracking
- Connection usage patterns

Use these monitoring features to identify and resolve performance issues proactively. 