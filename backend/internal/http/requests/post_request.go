package requests

type ContentInput struct {
	Content string `json:"content" validate:"required,min=1"`
	Type    string `json:"type,omitempty"`
}

type PostCreateRequest struct {
	Title      string         `json:"title" validate:"required,min=3,max=255"`
	Contents   []ContentInput `json:"contents" validate:"required,dive"`
	UserID     uint64         `json:"user_id" validate:"required,gt=0"`
	CategoryID uint64         `json:"category_id" validate:"required,gt=0"`
}

type PostUpdateRequest struct {
	Title      string         `json:"title" validate:"required,min=3,max=255"`
	Contents   []ContentInput `json:"contents" validate:"required,dive"`
	UserID     uint64         `json:"user_id" validate:"required,gt=0"`
	CategoryID uint64         `json:"category_id" validate:"required,gt=0"`
}
