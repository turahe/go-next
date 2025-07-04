package requests

type UserRoleUpdateInput struct {
	Role string `json:"role" validate:"required,oneof=admin editor moderator user guest"`
}
