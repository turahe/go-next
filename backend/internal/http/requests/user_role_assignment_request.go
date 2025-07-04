package requests

type UserRoleAssignmentInput struct {
	RoleID uint `json:"role_id" binding:"required"`
}
