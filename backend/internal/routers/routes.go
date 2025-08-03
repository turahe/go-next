package routers

import (
	"go-next/internal/http/controllers"
	v1 "go-next/internal/routers/v1"
	"go-next/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	// Initialize handlers
	authHandler := controllers.NewAuthHandler()
	postHandler := controllers.NewPostHandler(services.PostSvc)
	categoryHandler := controllers.NewCategoryHandler(services.CategorySvc)
	commentHandler := controllers.NewCommentHandler(services.CommentSvc)
	userHandler := controllers.NewUserHandler(services.UserSvc)
	roleHandler := controllers.NewRoleHandler(services.RoleSvc)
	userRoleHandler := controllers.NewUserRoleHandler(services.UserRoleSvc)
	menuHandler := controllers.NewMenuHandler(services.MenuSvc)
	dashboardHandler := controllers.NewDashboardHandler()

	// Initialize media handler
	mediaHandler := controllers.NewMediaHandler(services.MediaSvc)

	// Initialize tag handler
	tagHandler := controllers.NewTagHandler(services.TagSvc, nil) // TODO: Add logger

	// Initialize blog service and handler
	blogSvc := services.NewBlogService()
	blogHandler := controllers.NewBlogHandler(blogSvc)

	// Initialize WebSocket hub and handlers
	wsHub := services.NewHub()
	go wsHub.Run()
	notificationHandler := controllers.NewNotificationHandler()
	wsHandler := controllers.NewWebSocketHandler(wsHub)

	// Create API v1 group
	api := r.Group("/api/v1")

	// Register all route modules
	v1.RegisterAuthRoutes(api, authHandler)
	v1.RegisterBlogRoutes(api, blogHandler)
	v1.RegisterDashboardRoutes(api, dashboardHandler)
	v1.RegisterNotificationRoutes(api, notificationHandler)
	v1.RegisterAdminNotificationRoutes(api, notificationHandler)
	v1.RegisterWebSocketRoutes(api, wsHandler)

	// Register existing route modules
	v1.RegisterPostRoutes(api, postHandler, commentHandler)
	v1.RegisterUserRoutes(api, userHandler, userRoleHandler, authHandler)
	v1.RegisterCategoryRoutes(api, categoryHandler)
	v1.RegisterCommentRoutes(api, commentHandler)
	v1.RegisterRoleRoutes(api, roleHandler)
	v1.RegisterMediaRoutes(api, mediaHandler)
	v1.RegisterTagRoutes(api, tagHandler)
	v1.RegisterMenuRoutes(api, menuHandler)

	// Initialize Casbin handler and register routes
	casbinService := services.NewCasbinService()
	casbinHandler := controllers.NewCasbinHandler(casbinService)
	v1.RegisterCasbinRoutes(api, casbinHandler)

	// Register organization routes
	v1.RegisterOrganizationRoutes(api)
}
