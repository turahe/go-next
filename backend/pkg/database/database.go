package database

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"wordpress-go-next/backend/pkg/config"

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

	return nil
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

	var db *gorm.DB
	switch driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newDBLogger})
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newDBLogger})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: newDBLogger})
	case "sqlserver":
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{Logger: newDBLogger})
	}

	if err != nil {
		return nil, errors.New("failed to open database connection")
	}

	return db, nil

}

func buildDSN(driver string, configuration *config.Configuration) (string, error) {
	switch driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", configuration.Database.Username, configuration.Database.Password, configuration.Database.Host, configuration.Database.Port, configuration.Database.Dbname), nil
	case "postgres":
		mode := "disable"
		if configuration.Database.Sslmode {
			mode = "require"
		}
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", configuration.Database.Host, configuration.Database.Username, configuration.Database.Password, configuration.Database.Dbname, configuration.Database.Port, mode), nil
	case "sqlite":
		return "./" + configuration.Database.Dbname + ".db", nil
	case "sqlserver":
		mode := "disable"
		if configuration.Database.Sslmode {
			mode = "true"
		}
		return fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&encrypt=%s", configuration.Database.Username, configuration.Database.Password, configuration.Database.Host, configuration.Database.Port, configuration.Database.Dbname, mode), nil
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
