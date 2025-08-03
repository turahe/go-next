package v1

import (
	"go-next/internal/http/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(api *gin.RouterGroup, authHandler controllers.AuthHandler) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/reset-password", authHandler.ResetPassword)
		auth.POST("/refresh", authHandler.RefreshToken)
		// auth.POST("/logout", authHandler.Logout)
		// auth.POST("/request-password-reset", authHandler.RequestPasswordReset)
		// auth.POST("/change-password", authHandler.ChangePassword)
		// auth.POST("/verify-email", authHandler.VerifyEmail)
		// auth.POST("/resend-verification-email", authHandler.ResendVerificationEmail)
	}
}
