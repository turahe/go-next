package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterBlogRoutes(api *gin.RouterGroup, blogHandler controllers.BlogHandler) {
	blog := api.Group("/blog")
	{
		// Blog statistics (static routes first)
		blog.GET("/stats", blogHandler.GetBlogStats)
		blog.GET("/stats/categories", blogHandler.GetCategoryStats)
		blog.GET("/archive", blogHandler.GetMonthlyArchive)

		// Categories and tags (static routes first)
		blog.GET("/categories", blogHandler.GetPublicCategories)
		blog.GET("/tags", blogHandler.GetPublicTags)

		// Public blog endpoints (static routes first)
		blog.GET("/posts", blogHandler.GetPublicPosts)
		blog.GET("/posts/featured", blogHandler.GetFeaturedPosts)
		blog.GET("/posts/popular", blogHandler.GetPopularPosts)
		blog.GET("/search", blogHandler.SearchPosts)

		// Admin blog endpoints (static routes first)
		blog.POST("/posts", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "POST"), blogHandler.CreatePost)

		// Dynamic routes with parameters (must come after static routes)
		blog.GET("/posts/:slug", blogHandler.GetPublicPost)
		blog.GET("/posts/:slug/related", blogHandler.GetRelatedPosts)
		blog.POST("/posts/:id/view", blogHandler.IncrementViewCount)
		blog.PUT("/posts/:id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "PUT"), blogHandler.UpdatePost)
		blog.DELETE("/posts/:id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "DELETE"), blogHandler.DeletePost)
		blog.POST("/posts/:id/publish", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "POST"), blogHandler.PublishPost)
		blog.POST("/posts/:id/unpublish", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "POST"), blogHandler.UnpublishPost)
		blog.POST("/posts/:id/archive", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/blog/posts", "POST"), blogHandler.ArchivePost)

		// Category and tag dynamic routes
		blog.GET("/categories/:slug", blogHandler.GetCategoryBySlug)
		blog.GET("/categories/:slug/posts", blogHandler.GetPostsByCategory)
		blog.GET("/tags/:slug", blogHandler.GetTagBySlug)
		blog.GET("/tags/:slug/posts", blogHandler.GetPostsByTag)
	}
}
