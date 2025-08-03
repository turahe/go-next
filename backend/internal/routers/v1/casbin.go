package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterCasbinRoutes(api *gin.RouterGroup, casbinHandler controllers.CasbinHandler) {
	casbin := api.Group("/casbin")
	{
		// Policy management routes
		casbin.GET("/policies", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/policies", "GET"), casbinHandler.GetPolicies)
		casbin.POST("/policies", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/policies", "POST"), casbinHandler.AddPolicy)
		casbin.DELETE("/policies", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/policies", "DELETE"), casbinHandler.RemovePolicy)
		casbin.GET("/policies/filtered", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/policies", "GET"), casbinHandler.GetFilteredPolicies)

		// User role management routes
		casbin.GET("/users/:user_id/roles", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/users", "GET"), casbinHandler.GetUserRoles)
		casbin.POST("/users/:user_id/roles", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/users", "POST"), casbinHandler.AddRoleForUser)
		casbin.DELETE("/users/:user_id/roles", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/users", "DELETE"), casbinHandler.RemoveRoleForUser)
	}
}
