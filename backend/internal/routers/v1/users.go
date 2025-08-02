package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(api *gin.RouterGroup, userHandler controllers.UserHandler, userRoleHandler controllers.UserRoleHandler, authHandler *controllers.AuthHandler) {
	users := api.Group("/users")
	{
		users.GET(":id", userHandler.GetUserProfile)
		users.PUT(":id", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/users", "PUT"), userHandler.UpdateUserProfile)
		users.PUT(":id/role", middleware.JWTMiddleware(), middleware.CasbinMiddleware("/api/users", "PUT"), userHandler.UpdateUserRole)
		users.POST(":id/roles", middleware.JWTMiddleware(), userRoleHandler.AssignRoleToUser)
		users.DELETE(":id/roles/:role_id", middleware.JWTMiddleware(), userRoleHandler.RemoveRoleFromUser)
		users.GET(":id/roles", middleware.JWTMiddleware(), userRoleHandler.ListUserRoles)
		users.POST(":id/request-email-verification", authHandler.RequestEmailVerification)
		users.POST(":id/verify-email", authHandler.VerifyEmail)
		users.POST(":id/request-phone-verification", authHandler.RequestPhoneVerification)
		users.POST(":id/verify-phone", authHandler.VerifyPhone)
	}
}
