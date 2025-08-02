package database

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"go-next/internal/models"
	"go-next/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
)

type Database struct {
	*gorm.DB
}

func Setup() error {
	configuration := config.GetConfig()

	db, err := CreateDatabaseConnection(configuration)
	if err != nil {
		return err
	}

	DB = db

	// Auto-migrate all models (disabled for now due to schema conflicts)
	// if err := AutoMigrate(); err != nil {
	// 	return err
	// }

	return nil
}

// AutoMigrate performs database migrations for all models
func AutoMigrate() error {
	// Auto-migrate all models
	err := DB.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Comment{},
		&models.Category{},
		&models.Role{},
		&models.Media{},
		&models.Mediable{},
		&models.Content{},
		&models.Notification{},
	)

	return err
}

func CreateDatabaseConnection(configuration *config.Configuration) (*gorm.DB, error) {
	driver := strings.ToLower(configuration.Database.Driver)
	dsn, err := buildDSN(driver, configuration)
	if err != nil {
		return nil, errors.New("failed to build DSN")
	}

	logmode := configuration.Database.Logmode
	loglevel := logger.Silent
	if logmode {
		loglevel = logger.Info
	}
	newDBLogger := logger.New(
		log.New(getWriter(), "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  loglevel,    // Log level (Silent, Error, Warn, Info)
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
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
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)                  // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)                 // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // Maximum idle time of a connection

	return db, nil
}

func buildDSN(driver string, configuration *config.Configuration) (string, error) {
	switch driver {
	case "mysql":
		// Add connection pool and performance optimizations to MySQL DSN
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s&maxAllowedPacket=0",
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
	file, err := os.OpenFile("log/database.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return os.Stdout
	} else {
		return file
	}
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
