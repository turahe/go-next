package models

// CommentStatus represents the possible statuses of a comment
type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"
	CommentStatusApproved CommentStatus = "approved"
	CommentStatusRejected CommentStatus = "rejected"
)

// String returns the string representation of the status
func (s CommentStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s CommentStatus) IsValid() bool {
	switch s {
	case CommentStatusPending, CommentStatusApproved, CommentStatusRejected:
		return true
	default:
		return false
	}
}

// PostStatus represents the possible statuses of a post
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
)

// String returns the string representation of the status
func (s PostStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s PostStatus) IsValid() bool {
	switch s {
	case PostStatusDraft, PostStatusPublished, PostStatusArchived:
		return true
	default:
		return false
	}
}

// ModelType represents the possible types of models that can have comments
type ModelType string

const (
	ModelTypePost    ModelType = "post"
	ModelTypeComment ModelType = "comment"
)

// String returns the string representation of the model type
func (t ModelType) String() string {
	return string(t)
}

// IsValid checks if the model type is valid
func (t ModelType) IsValid() bool {
	switch t {
	case ModelTypePost, ModelTypeComment:
		return true
	default:
		return false
	}
}
