package config

import (
	"go-next/pkg/email"
	"go-next/pkg/redis"
	"go-next/pkg/storage"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     string
	Username string
	Password string
	Dbname   string
	Logmode  bool
	Sslmode  bool
}

type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     string
	Password string
	DB       int
}
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type WhatsAppConfig struct {
	BaseURL string
	Session string
}

type Configuration struct {
	Database  DatabaseConfig
	Port      string
	JwtSecret string
	SMTP      email.SMTPConfig
	Redis     redis.RedisConfig
	Storage   storage.StorageConfig
	WhatsApp  WhatsAppConfig
}

var (
	config *Configuration
	once   sync.Once
)

func LoadConfig() {
	// Load environment variables from .env files
	// Try to load from multiple possible locations
	envFiles := []string{
		".env",
		"../.env",
		"../../.env",
		".env.dev",
		"../.env.dev",
		"../../.env.dev",
		".env.prod",
		"../.env.prod",
		"../../.env.prod",
	}

	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			break
		}
	}

	config = &Configuration{
		Database: DatabaseConfig{
			Driver:   getEnvWithDefault("DB_TYPE", "sqlite"),
			Host:     getEnvWithDefault("DB_HOST", "localhost"),
			Port:     getEnvWithDefault("DB_PORT", "5432"),
			Username: getEnvWithDefault("DB_USER", "wordpress_user"),
			Password: getEnvWithDefault("DB_PASSWORD", "wordpress_password"),
			Dbname:   getEnvWithDefault("DB_NAME", "go_next"),
			Logmode:  os.Getenv("DB_LOGMODE") == "true",
			Sslmode:  os.Getenv("DB_SSLMODE") == "require",
		},
		Port:      getEnvWithDefault("PORT", "8080"),
		JwtSecret: getEnvWithDefault("JWT_SECRET", "your-super-secret-jwt-key-here"),
		SMTP: email.SMTPConfig{
			Host:     getEnvWithDefault("MAIL_HOST", "localhost"),
			Port:     getEnvAsInt("MAIL_PORT", 1025),
			Username: getEnvWithDefault("MAIL_USERNAME", ""),
			Password: getEnvWithDefault("MAIL_PASSWORD", ""),
			From:     getEnvWithDefault("MAIL_FROM", "noreply@example.com"),
		},
		Redis: redis.RedisConfig{
			Addr:     getEnvWithDefault("REDIS_HOST", "localhost") + ":" + getEnvWithDefault("REDIS_PORT", "6379"),
			Password: getEnvWithDefault("REDIS_PASSWORD", "redis_password"),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Storage: storage.StorageConfig{
			Driver:    storage.StorageDriver(os.Getenv("STORAGE_DRIVER")),
			Bucket:    os.Getenv("STORAGE_BUCKET"),
			Region:    os.Getenv("STORAGE_REGION"),
			Endpoint:  os.Getenv("STORAGE_ENDPOINT"),
			AccessKey: os.Getenv("STORAGE_ACCESS_KEY"),
			SecretKey: os.Getenv("STORAGE_SECRET_KEY"),
			LocalPath: os.Getenv("STORAGE_LOCAL_PATH"),
			CDNPrefix: os.Getenv("STORAGE_CDN_PREFIX"),
		},
		WhatsApp: WhatsAppConfig{
			BaseURL: getEnvOrDefault("WHATSAPP_BASE_URL", "http://localhost:3000"),
			Session: getEnvOrDefault("WHATSAPP_SESSION", "default"),
		},
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultVal int) int {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getEnvOrDefault(name, defaultVal string) string {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	return val
}

func GetConfig() *Configuration {
	once.Do(LoadConfig)
	return config
}
