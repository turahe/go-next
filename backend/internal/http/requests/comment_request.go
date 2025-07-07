package requests

type CommentCreateRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
	UserID  uint64 `json:"user_id" validate:"required,gt=0"`
	PostID  uint64 `json:"post_id" validate:"required,gt=0"`
}

type CommentUpdateRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
	UserID  uint64 `json:"user_id" validate:"required,gt=0"`
	PostID  uint64 `json:"post_id" validate:"required,gt=0"`
}
