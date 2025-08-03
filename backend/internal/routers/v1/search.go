package v1

import (
	"github.com/gin-gonic/gin"

	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"
)

// RegisterSearchRoutes registers search-related routes
func RegisterSearchRoutes(router *gin.RouterGroup, searchHandler *controllers.SearchHandler) {
	search := router.Group("/search")
	{
		// Public search endpoints
		search.GET("", searchHandler.Search)
		search.GET("/posts", searchHandler.SearchPosts)
		search.GET("/users", searchHandler.SearchUsers)
		search.GET("/categories", searchHandler.SearchCategories)
		search.GET("/media", searchHandler.SearchMedia)
		search.GET("/suggestions", searchHandler.GetSuggestions)
		search.GET("/stats", searchHandler.GetSearchStats)
		search.GET("/health", searchHandler.HealthCheck)

		// Admin-only endpoints (require authentication)
		admin := search.Group("")
		admin.Use(middleware.JWTMiddleware())
		admin.Use(middleware.CasbinMiddleware("search", "manage"))
		{
			admin.POST("/reindex", searchHandler.ReindexAll)
			admin.POST("/init", searchHandler.InitializeIndexes)
		}
	}
}
