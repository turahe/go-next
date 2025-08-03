package models

import (
	"testing"
)

func TestCommentStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   CommentStatus
		expected bool
	}{
		{"pending", CommentStatusPending, true},
		{"approved", CommentStatusApproved, true},
		{"rejected", CommentStatusRejected, true},
		{"invalid", CommentStatus("invalid"), false},
		{"empty", CommentStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.expected {
				t.Errorf("CommentStatus.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPostStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   PostStatus
		expected bool
	}{
		{"draft", PostStatusDraft, true},
		{"published", PostStatusPublished, true},
		{"archived", PostStatusArchived, true},
		{"invalid", PostStatus("invalid"), false},
		{"empty", PostStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.expected {
				t.Errorf("PostStatus.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestModelType_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		modelType ModelType
		expected  bool
	}{
		{"post", ModelTypePost, true},
		{"comment", ModelTypeComment, true},
		{"invalid", ModelType("invalid"), false},
		{"empty", ModelType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.modelType.IsValid(); got != tt.expected {
				t.Errorf("ModelType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCommentStatus_String(t *testing.T) {
	if CommentStatusPending.String() != "pending" {
		t.Errorf("CommentStatusPending.String() = %v, want %v", CommentStatusPending.String(), "pending")
	}
	if CommentStatusApproved.String() != "approved" {
		t.Errorf("CommentStatusApproved.String() = %v, want %v", CommentStatusApproved.String(), "approved")
	}
	if CommentStatusRejected.String() != "rejected" {
		t.Errorf("CommentStatusRejected.String() = %v, want %v", CommentStatusRejected.String(), "rejected")
	}
}

func TestPostStatus_String(t *testing.T) {
	if PostStatusDraft.String() != "draft" {
		t.Errorf("PostStatusDraft.String() = %v, want %v", PostStatusDraft.String(), "draft")
	}
	if PostStatusPublished.String() != "published" {
		t.Errorf("PostStatusPublished.String() = %v, want %v", PostStatusPublished.String(), "published")
	}
	if PostStatusArchived.String() != "archived" {
		t.Errorf("PostStatusArchived.String() = %v, want %v", PostStatusArchived.String(), "archived")
	}
}
