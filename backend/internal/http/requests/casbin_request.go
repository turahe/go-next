package requests

// PolicyRequest represents a Casbin policy request
type PolicyRequest struct {
	Subject string `json:"subject" binding:"required" example:"admin"`
	Domain  string `json:"domain" binding:"required" example:"*"`
	Object  string `json:"object" binding:"required" example:"/api/users"`
	Action  string `json:"action" binding:"required" example:"GET"`
}

// RoleAssignmentRequest represents a role assignment request
type RoleAssignmentRequest struct {
	Role   string `json:"role" binding:"required" example:"admin"`
	Domain string `json:"domain" binding:"required" example:"*"`
}
