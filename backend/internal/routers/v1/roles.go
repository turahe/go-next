package v1

import (
	"wordpress-go-next/backend/internal/http/controllers"
	"wordpress-go-next/backend/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(api *gin.RouterGroup, roleHandler controllers.RoleHandler) {
	roles := api.Group("/roles")
	{
		roles.GET("", roleHandler.GetRoles)
		roles.GET(":id", roleHandler.GetRole)
		roles.POST("", middleware.JWTMiddleware(), roleHandler.CreateRole)
		roles.PUT(":id", middleware.JWTMiddleware(), roleHandler.UpdateRole)
		roles.DELETE(":id", middleware.JWTMiddleware(), roleHandler.DeleteRole)
	}
}
