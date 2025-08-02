package requests

import (
	"github.com/google/uuid"
)

type UserRoleAssignmentInput struct {
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}
