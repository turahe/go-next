package database

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"go-next/internal/models"
	"go-next/pkg/config"
	"go-next/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	dbLogger "gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
	// Connection pool monitoring
	poolStats struct {
		sync.RWMutex
		MaxOpenConnections int
		OpenConnections    int
		InUse              int
		Idle               int
		WaitCount          int64
		WaitDuration       time.Duration
		MaxIdleClosed      int64
		MaxLifetimeClosed  int64
	}
)

type Database struct {
	*gorm.DB
}

// PoolStats returns current connection pool statistics
func PoolStats() map[string]interface{} {
	poolStats.RLock()
	defer poolStats.RUnlock()

	return map[string]interface{}{
		"max_open_connections": poolStats.MaxOpenConnections,
		"open_connections":     poolStats.OpenConnections,
		"in_use":               poolStats.InUse,
		"idle":                 poolStats.Idle,
		"wait_count":           poolStats.WaitCount,
		"wait_duration":        poolStats.WaitDuration,
		"max_idle_closed":      poolStats.MaxIdleClosed,
		"max_lifetime_closed":  poolStats.MaxLifetimeClosed,
	}
}

// StartPoolMonitoring starts a goroutine to monitor connection pool statistics
func StartPoolMonitoring(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updatePoolStats()
			}
		}
	}()
}

func updatePoolStats() {
	if DB == nil {
		return
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return
	}

	stats := sqlDB.Stats()

	poolStats.Lock()
	poolStats.MaxOpenConnections = stats.MaxOpenConnections
	poolStats.OpenConnections = stats.OpenConnections
	poolStats.InUse = stats.InUse
	poolStats.Idle = stats.Idle
	poolStats.WaitCount = stats.WaitCount
	poolStats.WaitDuration = stats.WaitDuration
	poolStats.MaxIdleClosed = stats.MaxIdleClosed
	poolStats.MaxLifetimeClosed = stats.MaxLifetimeClosed
	poolStats.Unlock()
}

func Setup() error {
	configuration := config.GetConfig()

	db, err := CreateDatabaseConnection(configuration)
	if err != nil {
		return err
	}

	DB = db

	// Start connection pool monitoring
	ctx := context.Background()
	StartPoolMonitoring(ctx)

	// Auto-migrate all models
	if err := AutoMigrate(); err != nil {
		return err
	}

	return nil
}

// AutoMigrate performs database migrations for all models
func AutoMigrate() error {
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Comment{},
		&models.Category{},
		&models.Role{},
		&models.UserRole{},
		&models.RoleMenu{},
		&models.CasbinRule{},
		&models.Organization{},
		&models.OrganizationUser{},
		&models.Media{},
		&models.Mediable{},
		&models.Menu{},
		&models.Content{},
		&models.Notification{},
		&models.Setting{},
	); err != nil {
		return err
	}

	// Setup default data after migration
	return SetupDefaultData()
}

// SetupDefaultData creates default roles and casbin rules
func SetupDefaultData() error {
	// Setup default roles
	if err := setupDefaultRoles(); err != nil {
		return err
	}

	// Setup default casbin rules
	if err := setupDefaultCasbinRules(); err != nil {
		return err
	}

	return nil
}

func CreateDatabaseConnection(configuration *config.Configuration) (*gorm.DB, error) {
	// Validate configuration
	if configuration.Database.Host == "" {
		return nil, errors.New("database host is required")
	}
	if configuration.Database.Port == "" {
		return nil, errors.New("database port is required")
	}
	if configuration.Database.Username == "" {
		return nil, errors.New("database username is required")
	}
	if configuration.Database.Dbname == "" {
		return nil, errors.New("database name is required")
	}

	driver := strings.ToLower(configuration.Database.Driver)

	// Log connection attempt
	log.Printf("Attempting to connect to database: %s://%s:%s/%s",
		driver, configuration.Database.Host, configuration.Database.Port, configuration.Database.Dbname)

	dsn, err := buildDSN(driver, configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	logmode := configuration.Database.Logmode
	loglevel := dbLogger.Silent
	if logmode {
		loglevel = dbLogger.Info
	}

	newDBLogger := dbLogger.New(
		log.New(getWriter(), "", log.LstdFlags),
		dbLogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  loglevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Configure GORM with optimized settings
	gormConfig := &gorm.Config{
		Logger: newDBLogger,
		// Disable automatic transaction wrapping for better performance
		DisableAutomaticPing: true,
		// Optimize for read-heavy workloads
		PrepareStmt: true,
		// Disable foreign key constraints for better performance (enable in production)
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	var db *gorm.DB
	switch driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dsn), gormConfig)
	case "sqlserver":
		db, err = gorm.Open(sqlserver.Open(dsn), gormConfig)
	default:
		logger.Errorf("unsupported database driver: %s", driver)
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		logger.Errorf("failed to open database connection: %w", err)
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool with configurable settings
	sqlDB, err := db.DB()
	if err != nil {
		logger.Errorf("failed to get underlying sql.DB: %w", err)
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings from configuration
	sqlDB.SetMaxIdleConns(configuration.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(configuration.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(configuration.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(configuration.Database.ConnMaxIdleTime)

	// Test the connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		logger.Errorf("failed to ping database: %w", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to database: %s://%s:%s/%s",
		driver, configuration.Database.Host, configuration.Database.Port, configuration.Database.Dbname)

	return db, nil
}

// HealthCheck performs a database health check
func HealthCheck() error {
	if DB == nil {
		logger.Errorf("database connection is nil")
		return errors.New("database connection is nil")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.Errorf("failed to get underlying sql.DB: %w", err)
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}

// GetConnectionWithTimeout gets a database connection with timeout
func GetConnectionWithTimeout(timeout time.Duration) (*gorm.DB, error) {
	if DB == nil {
		logger.Errorf("database connection is nil")
		return nil, errors.New("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a new session with timeout
	session := DB.WithContext(ctx)

	// Test the connection
	sqlDB, err := session.DB()
	if err != nil {
		logger.Errorf("failed to get underlying sql.DB: %w", err)
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		logger.Errorf("database ping failed: %w", err)
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return session, nil
}

// WithTransaction executes a function within a database transaction
func WithTransaction(fn func(*gorm.DB) error) error {
	if DB == nil {
		logger.Errorf("database connection is nil")
		return errors.New("database connection is nil")
	}

	return DB.Transaction(fn)
}

// WithTransactionContext executes a function within a database transaction with context
func WithTransactionContext(ctx context.Context, fn func(*gorm.DB) error) error {
	if DB == nil {
		logger.Errorf("database connection is nil")
		return errors.New("database connection is nil")
	}

	return DB.WithContext(ctx).Transaction(fn)
}

func buildDSN(driver string, configuration *config.Configuration) (string, error) {
	switch driver {
	case "mysql":
		// Add connection pool and performance optimizations to MySQL DSN
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s&maxAllowedPacket=0&interpolateParams=true",
			configuration.Database.Username,
			configuration.Database.Password,
			configuration.Database.Host,
			configuration.Database.Port,
			configuration.Database.Dbname), nil
	case "postgres":
		mode := "disable"
		if configuration.Database.Sslmode {
			mode = "require"
		}
		// Add connection pool and performance optimizations to PostgreSQL DSN
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s connect_timeout=10 application_name=wordpress_go_next",
			configuration.Database.Host,
			configuration.Database.Username,
			configuration.Database.Password,
			configuration.Database.Dbname,
			configuration.Database.Port,
			mode), nil
	case "sqlite":
		return "./data/" + configuration.Database.Dbname + ".db", nil
	case "sqlserver":
		mode := "disable"
		if configuration.Database.Sslmode {
			mode = "true"
		}
		return fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&encrypt=%s&connection+timeout=10",
			configuration.Database.Username,
			configuration.Database.Password,
			configuration.Database.Host,
			configuration.Database.Port,
			configuration.Database.Dbname,
			mode), nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", driver)
	}
}

func getWriter() io.Writer {
	if err := os.MkdirAll("log", 0755); err != nil {
		return os.Stdout
	}

	file, err := os.OpenFile("log/database.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return os.Stdout
	}

	return file
}

func GetDB() *gorm.DB {
	return DB
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// setupDefaultRoles creates default roles if they don't exist
func setupDefaultRoles() error {
	defaultRoles := []models.Role{
		{
			Name:        "super_admin",
			Description: "Super Administrator with full system access",
		},
		{
			Name:        "admin",
			Description: "Administrator with administrative privileges",
		},
		{
			Name:        "moderator",
			Description: "Moderator with content management privileges",
		},
		{
			Name:        "user",
			Description: "Regular user with basic access",
		},
		{
			Name:        "guest",
			Description: "Guest user with limited access",
		},
	}

	for _, role := range defaultRoles {
		var existingRole models.Role
		if err := DB.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := DB.Create(&role).Error; err != nil {
					return fmt.Errorf("failed to create role %s: %w", role.Name, err)
				}
			} else {
				return fmt.Errorf("failed to check role %s: %w", role.Name, err)
			}
		}
	}

	return nil
}

// setupDefaultCasbinRules creates default casbin rules
func setupDefaultCasbinRules() error {
	defaultRules := []models.CasbinRule{
		// Super admin can do everything
		{Ptype: "p", V0: "super_admin", V1: "*", V2: "*"},

		// Admin permissions
		{Ptype: "p", V0: "admin", V1: "users", V2: "read"},
		{Ptype: "p", V0: "admin", V1: "users", V2: "write"},
		{Ptype: "p", V0: "admin", V1: "posts", V2: "read"},
		{Ptype: "p", V0: "admin", V1: "posts", V2: "write"},
		{Ptype: "p", V0: "admin", V1: "comments", V2: "read"},
		{Ptype: "p", V0: "admin", V1: "comments", V2: "write"},
		{Ptype: "p", V0: "admin", V1: "categories", V2: "read"},
		{Ptype: "p", V0: "admin", V1: "categories", V2: "write"},
		{Ptype: "p", V0: "admin", V1: "roles", V2: "read"},
		{Ptype: "p", V0: "admin", V1: "roles", V2: "write"},
		{Ptype: "p", V0: "admin", V1: "menus", V2: "read"},
		{Ptype: "p", V0: "admin", V1: "menus", V2: "write"},

		// Moderator permissions
		{Ptype: "p", V0: "moderator", V1: "posts", V2: "read"},
		{Ptype: "p", V0: "moderator", V1: "posts", V2: "write"},
		{Ptype: "p", V0: "moderator", V1: "comments", V2: "read"},
		{Ptype: "p", V0: "moderator", V1: "comments", V2: "write"},
		{Ptype: "p", V0: "moderator", V1: "categories", V2: "read"},

		// User permissions
		{Ptype: "p", V0: "user", V1: "posts", V2: "read"},
		{Ptype: "p", V0: "user", V1: "posts", V2: "write"},
		{Ptype: "p", V0: "user", V1: "comments", V2: "read"},
		{Ptype: "p", V0: "user", V1: "comments", V2: "write"},
		{Ptype: "p", V0: "user", V1: "categories", V2: "read"},
		{Ptype: "p", V0: "user", V1: "profile", V2: "read"},
		{Ptype: "p", V0: "user", V1: "profile", V2: "write"},

		// Guest permissions
		{Ptype: "p", V0: "guest", V1: "posts", V2: "read"},
		{Ptype: "p", V0: "guest", V1: "categories", V2: "read"},
	}

	for _, rule := range defaultRules {
		var existingRule models.CasbinRule
		if err := DB.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
			rule.Ptype, rule.V0, rule.V1, rule.V2).First(&existingRule).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := DB.Create(&rule).Error; err != nil {
					return fmt.Errorf("failed to create casbin rule: %w", err)
				}
			} else {
				return fmt.Errorf("failed to check casbin rule: %w", err)
			}
		}
	}

	return nil
}
