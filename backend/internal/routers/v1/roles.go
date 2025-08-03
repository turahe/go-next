package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(api *gin.RouterGroup, roleHandler controllers.RoleHandler) {
	roles := api.Group("/roles")
	{
		roles.GET("", middleware.JWTMiddleware(), roleHandler.GetRoles)
		roles.GET(":id", middleware.JWTMiddleware(), roleHandler.GetRole)
		roles.POST("", middleware.JWTMiddleware(), roleHandler.CreateRole)
		roles.PUT(":id", middleware.JWTMiddleware(), roleHandler.UpdateRole)
		roles.DELETE(":id", middleware.JWTMiddleware(), roleHandler.DeleteRole)

		// Menu-related routes
		roles.GET(":id/menus", middleware.JWTMiddleware(), roleHandler.GetRoleMenus)
		roles.POST(":id/menus", middleware.JWTMiddleware(), roleHandler.AssignMenuToRole)
		roles.DELETE(":id/menus", middleware.JWTMiddleware(), roleHandler.RemoveMenuFromRole)
	}
}
