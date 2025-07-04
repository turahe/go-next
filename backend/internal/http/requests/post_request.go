package requests

type PostCreateRequest struct {
	Title      string `json:"title" validate:"required,min=3,max=255"`
	Content    string `json:"content" validate:"required,min=1"`
	UserID     uint   `json:"user_id" validate:"required,gt=0"`
	CategoryID uint   `json:"category_id" validate:"required,gt=0"`
}

type PostUpdateRequest struct {
	Title      string `json:"title" validate:"required,min=3,max=255"`
	Content    string `json:"content" validate:"required,min=1"`
	UserID     uint   `json:"user_id" validate:"required,gt=0"`
	CategoryID uint   `json:"category_id" validate:"required,gt=0"`
}
