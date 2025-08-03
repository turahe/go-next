package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterCasbinAdapterRoutes(api *gin.RouterGroup, adapterHandler controllers.CasbinAdapterHandler) {
	adapter := api.Group("/casbin/adapter")
	{
		// Database statistics and monitoring
		adapter.GET("/stats", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.GetDatabaseStats)
		adapter.GET("/policies/count", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.GetPolicyCount)
		adapter.GET("/roles/count", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.GetRoleCount)
		adapter.GET("/policies/count/role", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.GetPolicyCountByRole)
		adapter.GET("/users/roles/count", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.GetUserRoleCount)

		// Configuration
		adapter.GET("/config", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.GetAdapterConfig)

		// Backup and restore
		adapter.GET("/backup", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "GET"), adapterHandler.BackupPolicies)
		adapter.POST("/restore", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "POST"), adapterHandler.RestorePolicies)

		// Clear operations (dangerous - admin only)
		adapter.DELETE("/clear/all", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "DELETE"), adapterHandler.ClearAllPolicies)
		adapter.DELETE("/clear/policies", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "DELETE"), adapterHandler.ClearPolicies)
		adapter.DELETE("/clear/roles", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/casbin/adapter", "DELETE"), adapterHandler.ClearRoles)
	}
}
