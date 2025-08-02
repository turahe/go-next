package internal

import (
	"context"
	"fmt"
	"go-next/internal/http/middleware"
	"go-next/internal/routers"
	"go-next/internal/services"
	"go-next/pkg/config"
	"go-next/pkg/database"
	"go-next/pkg/email"
	"go-next/pkg/redis"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var Emailer *email.EmailService
var RedisClient *redis.RedisService

// GetConfig returns the application configuration
func GetConfig() *config.Configuration {
	return config.GetConfig()
}

func InitEmailer() {
	cfg := config.GetConfig()
	Emailer = email.NewEmailService(cfg.SMTP)
}

func InitRedis() {
	cfg := config.GetConfig()
	RedisClient = redis.NewRedisService(cfg.Redis)
	services.GlobalRedisClient = RedisClient
}

func RunServer(host, port string) {
	// Initialize services
	if err := database.Setup(); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	if err := services.InitCasbin(); err != nil {
		log.Fatalf("Failed to initialize Casbin: %v", err)
	}

	// Initialize Redis and Email services
	InitRedis()
	InitEmailer()

	// Set Gin mode based on environment
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = os.Getenv("APP_ENV")
		if ginMode == "development" || ginMode == "dev" {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
	} else {
		gin.SetMode(ginMode)
	}

	r := gin.New()

	// Use custom logger and recovery middleware
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())

	// Add middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.ResponseTimeMiddleware())
	r.Use(middleware.ValidationMiddleware())

	// Register routes
	routers.RegisterRoutes(r)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	if port == "" {
		port = "8080"
	}
	if host == "" {
		host = "0.0.0.0"
	}
	addr := host + ":" + port

	// Create HTTP server with optimized settings
	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s...", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
