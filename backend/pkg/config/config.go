package config

import (
	"os"
	"strconv"
	"sync"
	"wordpress-go-next/backend/pkg/email"
	"wordpress-go-next/backend/pkg/redis"
	"wordpress-go-next/backend/pkg/storage"

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
	godotenv.Load()
	config = &Configuration{
		Database: DatabaseConfig{
			Driver:   os.Getenv("DB_TYPE"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Dbname:   os.Getenv("DB_NAME"),
			Logmode:  os.Getenv("DB_LOGMODE") == "true",
			Sslmode:  os.Getenv("DB_SSLMODE") == "true",
		},
		Port:      os.Getenv("PORT"),
		JwtSecret: os.Getenv("JWT_SECRET"),
		SMTP: email.SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: os.Getenv("SMTP_USER"),
			Password: os.Getenv("SMTP_PASS"),
			From:     os.Getenv("SMTP_FROM"),
		},
		Redis: redis.RedisConfig{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
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
