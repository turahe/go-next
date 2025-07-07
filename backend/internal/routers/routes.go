package routers

import (
	"log"
	"wordpress-go-next/backend/docs"
	"wordpress-go-next/backend/internal/http/controllers"
	"wordpress-go-next/backend/internal/http/middleware"
	v1 "wordpress-go-next/backend/internal/routers/v1"
	"wordpress-go-next/backend/internal/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterRoutes registers all API routes to the given Gin engine.
func RegisterRoutes(r *gin.Engine) {
	authHandler := controllers.NewAuthHandler(services.AuthSvc)
	adminAuthHandler := controllers.NewAuthHandler(services.AuthSvc)
	// Initialize handlers for posts, comments, categories, users, roles, media
	postHandler := controllers.NewPostHandler(services.PostSvc)
	commentHandler := controllers.NewCommentHandler(services.CommentSvc)
	mediaSvc := services.ServiceMgr.MediaService
	categoryHandler := controllers.NewCategoryHandler(services.CategorySvc, mediaSvc)
	userHandler := controllers.NewUserHandler(services.UserSvc)
	userRoleHandler := controllers.NewUserRoleHandler(services.UserRoleSvc)
	roleHandler := controllers.NewRoleHandler(services.RoleSvc)
	mediaHandler := controllers.NewMediaHandler(mediaSvc)
	tagHandler := controllers.NewTagHandler(services.TagSvc, nil)

	r.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})

	r.POST("/api/register", authHandler.Register)
	r.POST("/api/admin/register", adminAuthHandler.Register)
	r.POST("/api/login", authHandler.Login)

	api := r.Group("/api/v1")

	v1.RegisterTagRoutes(api, tagHandler)
	v1.RegisterPostRoutes(api, postHandler, commentHandler)
	v1.RegisterCategoryRoutes(api, categoryHandler)
	v1.RegisterCommentRoutes(api, commentHandler)
	v1.RegisterUserRoutes(api, userHandler, userRoleHandler, authHandler)
	v1.RegisterRoleRoutes(api, roleHandler)
	v1.RegisterMediaRoutes(api, mediaHandler)

	// Password reset (not grouped under users for simplicity)
	api.POST("/request-password-reset", authHandler.RequestPasswordReset)
	api.POST("/reset-password", authHandler.ResetPassword)
	// Refresh token endpoint
	api.POST("/auth/refresh", authHandler.RefreshToken)
}

// RunServer sets up the Gin engine, middleware, Swagger docs, and starts the server.
func RunServer(host, port string) {
	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	docs.SwaggerInfo.Title = "WordPress Go Next API"
	docs.SwaggerInfo.Description = "API documentation for the WordPress Go Next backend."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api"

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	RegisterRoutes(r)

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
