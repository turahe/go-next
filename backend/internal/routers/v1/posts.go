package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterPostRoutes(api *gin.RouterGroup, postHandler controllers.PostHandler, commentHandler controllers.CommentHandler) {
	posts := api.Group("/posts")
	{
		posts.GET("", postHandler.GetPosts)
		posts.POST("", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/posts", "POST"), postHandler.CreatePost)

		// Comments route must come before the :id route to avoid conflicts
		posts.GET(":id/comments", commentHandler.GetCommentsByPost)

		// Post CRUD routes
		posts.GET(":id", postHandler.GetPost)
		posts.PUT(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/posts", "PUT"), postHandler.UpdatePost)
		posts.DELETE(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/posts", "DELETE"), postHandler.DeletePost)
	}
}
