package requests

type RoleCreateRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=32,unique_role"`
	Description string `json:"description" binding:"omitempty,max=255"`
}
