package v1

import (
	"wordpress-go-next/backend/internal/http/controllers"
	"wordpress-go-next/backend/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterTagRoutes(api *gin.RouterGroup, tagHandler *controllers.TagHandler) {
	tags := api.Group("/tags")
	{
		tags.GET("", tagHandler.ListTags)
		tags.GET("/search", tagHandler.SearchTags)
		tags.GET("/count", tagHandler.GetTagCount)
		tags.GET("/slug/:slug", tagHandler.GetTagBySlug)
		tags.GET(":id", tagHandler.GetTagByID)
		tags.POST("", middleware.JWTMiddleware(), tagHandler.CreateTag)
		tags.PUT(":id", middleware.JWTMiddleware(), tagHandler.UpdateTag)
		tags.DELETE(":id", middleware.JWTMiddleware(), tagHandler.DeleteTag)

		// Entity tagging
		tags.GET("/entity", tagHandler.GetTagsByEntity)
		tags.POST("/entity", middleware.JWTMiddleware(), tagHandler.AddTagToEntity)
		tags.DELETE("/entity", middleware.JWTMiddleware(), tagHandler.RemoveTagFromEntity)
		tags.GET("/entities", tagHandler.GetEntitiesByTag)
	}
}
