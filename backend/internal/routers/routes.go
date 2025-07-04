package routers

import (
	"wordpress-go-next/backend/internal/http/controllers"
	"wordpress-go-next/backend/internal/http/middleware"
	"wordpress-go-next/backend/internal/services"
	"wordpress-go-next/backend/pkg/config"
	"wordpress-go-next/backend/pkg/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	authHandler := controllers.NewAuthHandler()
	adminAuthHandler := controllers.NewAdminAuthHandler()
	store, _ := storage.NewStorageService(config.GetConfig().Storage)
	mediaSvc := services.NewMediaService(store)
	postHandler := controllers.NewPostHandler(services.PostSvc)
	categoryHandler := controllers.NewCategoryHandler(services.CategorySvc, mediaSvc)
	commentHandler := controllers.NewCommentHandler(services.CommentSvc)
	userHandler := controllers.NewUserHandler(services.UserSvc)
	roleHandler := controllers.NewRoleHandler(services.RoleSvc)
	mediaHandler := controllers.NewMediaHandler(mediaSvc)
	userRoleHandler := controllers.NewUserRoleHandler(services.UserRoleSvc)

	r.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})

	r.POST("/api/register", authHandler.Register)
	r.POST("/api/admin/register", adminAuthHandler.Register)
	r.POST("/api/login", authHandler.Login)

	api := r.Group("/api/v1")

	// Posts
	posts := api.Group("/posts")
	{
		posts.GET("", postHandler.GetPosts)
		posts.GET(":id", postHandler.GetPost)
		posts.POST("", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/posts", "POST"), postHandler.CreatePost)
		posts.PUT(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/posts", "PUT"), postHandler.UpdatePost)
		posts.DELETE(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/posts", "DELETE"), postHandler.DeletePost)
		posts.GET(":post_id/comments", commentHandler.GetCommentsByPost)
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
		users.GET(":id", userHandler.GetUserProfile)
		users.PUT(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/users", "PUT"), userHandler.UpdateUserProfile)
		users.PUT(":id/role", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/users", "PUT"), userHandler.UpdateUserRole)
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
		media.GET(":id/siblings", mediaHandler.GetSiblingMedia)
		media.GET(":id/parent", mediaHandler.GetParentMedia)
		media.GET(":id/descendants", mediaHandler.GetDescendantMedia)
		media.GET(":id/children", mediaHandler.GetChildrenMedia)
		media.POST("/nested", middleware.JWTMiddleware(), mediaHandler.CreateMediaNested)
		media.POST(":id/move", middleware.JWTMiddleware(), mediaHandler.MoveMediaNested)
		media.DELETE(":id/nested", middleware.JWTMiddleware(), mediaHandler.DeleteMediaNested)
	}

	// Password reset (not grouped under users for simplicity)
	api.POST("/request-password-reset", authHandler.RequestPasswordReset)
	api.POST("/reset-password", authHandler.ResetPassword)
	// Refresh token endpoint
	api.POST("/auth/refresh", authHandler.RefreshToken)
}
