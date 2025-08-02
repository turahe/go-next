package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterCommentRoutes(api *gin.RouterGroup, commentHandler controllers.CommentHandler) {
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
}
