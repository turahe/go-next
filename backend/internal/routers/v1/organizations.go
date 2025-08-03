package v1

import (
	"go-next/internal/http/controllers"
	"go-next/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterOrganizationRoutes registers organization-related routes
func RegisterOrganizationRoutes(r *gin.RouterGroup) {
	organizationHandler := controllers.NewOrganizationHandler()

	// Organization CRUD routes
	organizations := r.Group("/organizations")
	{
		// Public routes (if any)
		organizations.GET("", organizationHandler.GetAllOrganizations)

		// Protected routes
		organizations.Use(middleware.JWTMiddleware())
		{
			organizations.POST("", organizationHandler.CreateOrganization)
			organizations.GET("/:id", organizationHandler.GetOrganization)
			organizations.PUT("/:id", organizationHandler.UpdateOrganization)
			organizations.DELETE("/:id", organizationHandler.DeleteOrganization)

			// Organization user management
			organizations.POST("/:id/users/:user_id", organizationHandler.AddUserToOrganization)
			organizations.DELETE("/:id/users/:user_id", organizationHandler.RemoveUserFromOrganization)

			// Organization policies
			organizations.GET("/:id/policies", organizationHandler.GetOrganizationPolicies)

			// Organization user roles
			organizations.GET("/:id/users/:user_id/role", organizationHandler.GetUserRoleInOrganization)
			organizations.PUT("/:id/users/:user_id/role", organizationHandler.UpdateUserRoleInOrganization)
			organizations.GET("/:id/users-with-roles", organizationHandler.GetOrganizationUsersWithRoles)
		}
	}
}
