package internal

import (
	"log"
	"wordpress-go-next/backend/internal/routers"

	"wordpress-go-next/backend/docs"
	"wordpress-go-next/backend/internal/services"
	"wordpress-go-next/backend/pkg/config"
	"wordpress-go-next/backend/pkg/email"
	"wordpress-go-next/backend/pkg/redis"

	"wordpress-go-next/backend/internal/http/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var Emailer *email.EmailService
var RedisClient *redis.RedisService

func InitEmailer() {
	cfg := config.GetConfig()
	Emailer = email.NewEmailService(cfg.SMTP)
}

func InitRedis() {
	cfg := config.GetConfig()
	RedisClient = redis.NewRedisService(cfg.Redis)
}

func RunServer(host, port string) {
	if err := services.InitCasbin(); err != nil {
		panic(err)
	}

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	docs.SwaggerInfo.Title = "WordPress Go Next API"
	docs.SwaggerInfo.Description = "API documentation for the WordPress Go Next backend."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api"

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routers.RegisterRoutes(r)

	if port == "" {
		port = "8080"
	}
	if host == "" {
		host = "0.0.0.0"
	}
	addr := host + ":" + port
	log.Printf("Starting server on %s...", addr)
	r.Run(addr)
}
