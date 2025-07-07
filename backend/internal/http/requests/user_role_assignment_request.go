package requests

type UserRoleAssignmentRequest struct {
	RoleID uint64 `json:"role_id" binding:"required"`
}
