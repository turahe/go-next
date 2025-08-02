package routers

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"
	"go-next/internal/services"
	"go-next/pkg/config"
	"go-next/pkg/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	authHandler := controllers.NewAuthHandler()
	store, _ := storage.NewStorageService(config.GetConfig().Storage)
	mediaSvc := services.NewMediaService(store, services.GlobalRedisClient)
	postHandler := controllers.NewPostHandler(services.PostSvc)
	categoryHandler := controllers.NewCategoryHandler(services.CategorySvc, mediaSvc)
	commentHandler := controllers.NewCommentHandler(services.CommentSvc)
	userHandler := controllers.NewUserHandler(services.UserSvc)
	roleHandler := controllers.NewRoleHandler(services.RoleSvc)
	mediaHandler := controllers.NewMediaHandler(mediaSvc)
	userRoleHandler := controllers.NewUserRoleHandler(services.UserRoleSvc)
	dashboardHandler := controllers.NewDashboardHandler()

	// Initialize blog service and handler
	blogSvc := services.NewBlogService()
	blogHandler := controllers.NewBlogHandler(blogSvc)

	// Initialize WebSocket hub and handlers
	wsHub := services.NewHub()
	go wsHub.Run()
	notificationHandler := controllers.NewNotificationHandler()
	wsHandler := controllers.NewWebSocketHandler(wsHub)

	api := r.Group("/api/v1")
	api.POST("/register", authHandler.Register)
	api.POST("/login", authHandler.Login)
	api.POST("/request-password-reset", authHandler.RequestPasswordReset)
	api.POST("/reset-password", authHandler.ResetPassword)
	api.POST("/auth/refresh", authHandler.RefreshToken)

	// Posts
	posts := api.Group("/posts")
	{
		posts.GET("", postHandler.GetPosts)
		posts.GET(":id", postHandler.GetPost)
		posts.POST("", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/posts", "POST"), postHandler.CreatePost)
		posts.PUT(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/posts", "PUT"), postHandler.UpdatePost)
		posts.DELETE(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/posts", "DELETE"), postHandler.DeletePost)
	}

	// Blog endpoints (public)
	blog := api.Group("/blog")
	{
		// Public blog endpoints
		blog.GET("/posts", blogHandler.GetPublicPosts)
		blog.GET("/posts/featured", blogHandler.GetFeaturedPosts)
		blog.GET("/posts/popular", blogHandler.GetPopularPosts)
		blog.GET("/posts/:slug", blogHandler.GetPublicPost)
		blog.GET("/posts/:slug/related", blogHandler.GetRelatedPosts)
		blog.GET("/search", blogHandler.SearchPosts)

		// Blog statistics
		blog.GET("/stats", blogHandler.GetBlogStats)
		blog.GET("/stats/categories", blogHandler.GetCategoryStats)
		blog.GET("/archive", blogHandler.GetMonthlyArchive)

		// Categories and tags
		blog.GET("/categories", blogHandler.GetPublicCategories)
		blog.GET("/categories/:slug", blogHandler.GetCategoryBySlug)
		blog.GET("/categories/:slug/posts", blogHandler.GetPostsByCategory)
		blog.GET("/tags", blogHandler.GetPublicTags)
		blog.GET("/tags/:slug", blogHandler.GetTagBySlug)
		blog.GET("/tags/:slug/posts", blogHandler.GetPostsByTag)

		// View count tracking
		blog.POST("/posts/:id/view", blogHandler.IncrementViewCount)

		// Admin blog endpoints (require authentication)
		blog.POST("/posts", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "POST"), blogHandler.CreatePost)
		blog.PUT("/posts/:id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "PUT"), blogHandler.UpdatePost)
		blog.DELETE("/posts/:id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "DELETE"), blogHandler.DeletePost)
		blog.POST("/posts/:id/publish", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "POST"), blogHandler.PublishPost)
		blog.POST("/posts/:id/unpublish", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "POST"), blogHandler.UnpublishPost)
		blog.POST("/posts/:id/archive", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "POST"), blogHandler.ArchivePost)
	}

	// Categories
	categories := api.Group("/categories")
	{
		categories.GET("", categoryHandler.GetCategories)
		categories.GET(":id", categoryHandler.GetCategory)
		categories.GET(":id/children", categoryHandler.GetChildrenCategories)
		categories.POST("", middleware.JWTMiddleware(), categoryHandler.CreateCategory)
		categories.PUT(":id", middleware.JWTMiddleware(), categoryHandler.UpdateCategory)
		categories.DELETE(":id", middleware.JWTMiddleware(), categoryHandler.DeleteCategory)
		categories.POST("/nested", middleware.JWTMiddleware(), categoryHandler.CreateCategoryNested)
		categories.POST(":id/move", middleware.JWTMiddleware(), categoryHandler.MoveCategoryNested)
		categories.DELETE(":id/nested", middleware.JWTMiddleware(), categoryHandler.DeleteCategoryNested)
	}

	// Comments
	comments := api.Group("/comments")
	{
		comments.GET(":id", commentHandler.GetComment)
		comments.GET(":id/siblings", commentHandler.GetSiblingComments)
		comments.GET(":id/parent", commentHandler.GetParentComment)
		comments.GET(":id/descendants", commentHandler.GetDescendantComments)
		comments.GET(":id/children", commentHandler.GetChildrenComments)
		comments.POST("", middleware.JWTMiddleware(), commentHandler.CreateComment)
		comments.PUT(":id", middleware.JWTMiddleware(), commentHandler.UpdateComment)
		comments.DELETE(":id", middleware.JWTMiddleware(), commentHandler.DeleteComment)
		comments.POST("/nested", middleware.JWTMiddleware(), commentHandler.CreateCommentNested)
		comments.POST(":id/move", middleware.JWTMiddleware(), commentHandler.MoveCommentNested)
		comments.DELETE(":id/nested", middleware.JWTMiddleware(), commentHandler.DeleteCommentNested)
	}

	// Users
	users := api.Group("/users")
	{
		users.GET("", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/users", "GET"), userHandler.GetUsers)
		users.GET(":id", userHandler.GetUserProfile)
		users.POST("", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/users", "POST"), userHandler.UserCreate)
		users.PUT(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/users", "PUT"), userHandler.UpdateUserProfile)
		users.PUT(":id/role", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/users", "PUT"), userHandler.UpdateUserRole)
		users.DELETE(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/users", "DELETE"), userHandler.DeleteUser)
		users.POST(":id/roles", middleware.JWTMiddleware(), userRoleHandler.AssignRoleToUser)
		users.DELETE(":id/roles/:role_id", middleware.JWTMiddleware(), userRoleHandler.RemoveRoleFromUser)
		users.GET(":id/roles", middleware.JWTMiddleware(), userRoleHandler.ListUserRoles)
		users.POST(":id/request-email-verification", authHandler.RequestEmailVerification)
		users.POST(":id/verify-email", authHandler.VerifyEmail)
		users.POST(":id/request-phone-verification", authHandler.RequestPhoneVerification)
		users.POST(":id/verify-phone", authHandler.VerifyPhone)
	}

	// Roles
	roles := api.Group("/roles")
	{
		roles.GET("", roleHandler.GetRoles)
		roles.GET(":id", roleHandler.GetRole)
		roles.POST("", middleware.JWTMiddleware(), roleHandler.CreateRole)
		roles.PUT(":id", middleware.JWTMiddleware(), roleHandler.UpdateRole)
		roles.DELETE(":id", middleware.JWTMiddleware(), roleHandler.DeleteRole)
	}

	// Media
	media := api.Group("/media")
	{
		media.POST("/upload", middleware.JWTMiddleware(), mediaHandler.UploadMedia)
		media.POST(":id/associate", middleware.JWTMiddleware(), mediaHandler.AssociateMedia)
	}

	// Dashboard
	dashboard := api.Group("/dashboard")
	{
		dashboard.GET("/stats", middleware.JWTMiddleware(), dashboardHandler.GetDashboardStats)
	}

	// Notifications
	notifications := api.Group("/notifications")
	{
		notifications.GET("", middleware.JWTMiddleware(), notificationHandler.GetUserNotifications)
		notifications.GET("/unread-count", middleware.JWTMiddleware(), notificationHandler.GetUnreadCount)
		notifications.PUT(":id/read", middleware.JWTMiddleware(), notificationHandler.MarkAsRead)
		notifications.PUT("/mark-all-read", middleware.JWTMiddleware(), notificationHandler.MarkAllAsRead)
		notifications.DELETE(":id", middleware.JWTMiddleware(), notificationHandler.DeleteNotification)
		notifications.DELETE("", middleware.JWTMiddleware(), notificationHandler.DeleteAllNotifications)
	}

	// WebSocket
	ws := api.Group("/ws")
	{
		ws.GET("/status", wsHandler.GetWebSocketStatus)
		ws.GET("/connect", middleware.JWTMiddleware(), wsHandler.HandleWebSocket)
	}

	// Admin notifications
	admin := api.Group("/admin")
	{
		admin.POST("/notifications", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/admin/notifications", "POST"), notificationHandler.CreateNotification)
	}

}
