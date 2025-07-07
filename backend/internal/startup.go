package startup

import (
	"wordpress-go-next/backend/internal/routers"
	"wordpress-go-next/backend/internal/services"
	"wordpress-go-next/backend/pkg/config"
	"wordpress-go-next/backend/pkg/email"
	"wordpress-go-next/backend/pkg/logger"
	"wordpress-go-next/backend/pkg/redis"
	"wordpress-go-next/backend/pkg/storage"
)

var RedisClient *redis.RedisService

func RunServer(host, port string) {
	if err := services.InitCasbin(); err != nil {
		panic(err)
	}

	// Initialize services with Redis and logging
	cfg := config.GetConfig()

	email.NewEmailService(cfg.SMTP)

	store, _ := storage.NewStorageService(cfg.Storage)

	RedisClient = redis.NewRedisService(cfg.Redis)
	services.InitializeServices(RedisClient, store, logger.LogLevelInfo)

	// Now handled in routers.RunServer
	routers.RunServer(host, port)
}
