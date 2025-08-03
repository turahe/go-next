package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterMenuRoutes(api *gin.RouterGroup, menuHandler controllers.MenuHandler) {
	menus := api.Group("/menus")
	{
		menus.GET("", menuHandler.GetMenus)
		menus.GET("tree", menuHandler.GetMenuTree)
		menus.GET("parent/:parent_id", menuHandler.GetMenuByParent)
		menus.GET(":id", menuHandler.GetMenu)
		menus.GET(":id/descendants", menuHandler.GetMenuDescendants)
		menus.GET(":id/ancestors", menuHandler.GetMenuAncestors)
		menus.GET(":id/siblings", menuHandler.GetMenuSiblings)
		menus.POST("", middleware.JWTMiddleware(), menuHandler.CreateMenu)
		menus.PUT(":id", middleware.JWTMiddleware(), menuHandler.UpdateMenu)
		menus.DELETE(":id", middleware.JWTMiddleware(), menuHandler.DeleteMenu)
		menus.POST(":id/move", middleware.JWTMiddleware(), menuHandler.MoveMenu)

		// Role-related routes
		menus.GET(":id/roles", menuHandler.GetMenuRoles)
		menus.POST(":id/roles", middleware.JWTMiddleware(), menuHandler.AssignRoleToMenu)
		menus.DELETE(":id/roles", middleware.JWTMiddleware(), menuHandler.RemoveRoleFromMenu)
	}
}
