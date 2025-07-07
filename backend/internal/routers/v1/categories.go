package v1

import (
	"wordpress-go-next/backend/internal/http/controllers"
	"wordpress-go-next/backend/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterCategoryRoutes(api *gin.RouterGroup, categoryHandler controllers.CategoryHandler) {
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
}
