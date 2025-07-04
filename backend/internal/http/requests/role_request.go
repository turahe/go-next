package requests

type RoleCreateRequest struct {
	Name string `json:"name" binding:"required,min=2,max=32"`
}
