package requests

type CategoryCreateRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	ParentID    int64  `json:"parent_id" validate:"omitempty,gt=0"` // Assuming ParentID is optional and should be greater than 0 if provided
}

type CategoryUpdateRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	ParentID    int64  `json:"parent_id" validate:"omitempty,gt=0"` // Assuming ParentID is optional and should be greater than 0 if provided
}
