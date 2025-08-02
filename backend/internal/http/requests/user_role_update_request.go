package requests

type UserRoleUpdateRequest struct {
	Role string `json:"role" validate:"required,oneof=admin editor moderator user guest"`
}
