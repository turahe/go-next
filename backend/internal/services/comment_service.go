package services

import (
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/redis"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentService interface {
	GetCommentsByPost(postID string) ([]models.Comment, error)
	GetCommentByID(id string) (*models.Comment, error)
	CreateComment(comment *models.Comment) error
	UpdateComment(comment *models.Comment) error
	DeleteComment(id string) error
	CreateNested(comment *models.Comment, parentID *uuid.UUID) error
	MoveNested(id uuid.UUID, newParentID *uuid.UUID) error
	DeleteNested(id uuid.UUID) error
	GetSiblingComments(id uuid.UUID) ([]models.Comment, error)
	GetParentComment(id uuid.UUID) (*models.Comment, error)
	GetDescendantComments(id uuid.UUID) ([]models.Comment, error)
	GetChildrenComments(id uuid.UUID) ([]models.Comment, error)
}

type commentService struct {
	redisService *redis.RedisService
}

func NewCommentService(redisService *redis.RedisService) CommentService {
	return &commentService{
		redisService: redisService,
	}
}

func (s *commentService) GetCommentsByPost(postID string) ([]models.Comment, error) {
	postUUID, err := uuid.Parse(postID)
	if err != nil {
		return nil, err
	}

	var comments []models.Comment
	err = database.DB.Where("post_id = ?", postUUID).Find(&comments).Error
	return comments, err
}

func (s *commentService) GetCommentByID(id string) (*models.Comment, error) {
	commentID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var comment models.Comment
	err = database.DB.First(&comment, commentID).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (s *commentService) CreateComment(comment *models.Comment) error {
	return database.DB.Create(comment).Error
}

func (s *commentService) UpdateComment(comment *models.Comment) error {
	return database.DB.Save(comment).Error
}

func (s *commentService) DeleteComment(id string) error {
	commentID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return database.DB.Delete(&models.Comment{}, commentID).Error
}

func (s *commentService) CreateNested(comment *models.Comment, parentID *uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if parentID != nil {
			var parent models.Comment
			if err := tx.First(&parent, parentID).Error; err != nil {
				return err
			}
			if parent.RecordRight == 0 {
				return gorm.ErrRecordNotFound
			}

			// Update parent's right value
			tx.Model(&models.Comment{}).
				Where("record_right >= ?", parent.RecordRight).
				Update("record_right", gorm.Expr("record_right + 2"))
			tx.Model(&models.Comment{}).
				Where("record_left > ?", parent.RecordRight).
				Update("record_left", gorm.Expr("record_left + 2"))

			comment.RecordLeft = parent.RecordRight
			comment.RecordRight = parent.RecordRight + 1
			comment.RecordDept = parent.RecordDept + 1
			comment.ParentID = parentID
		} else {
			// Create as root
			var maxRight int
			tx.Model(&models.Comment{}).Select("COALESCE(MAX(record_right), 0)").Scan(&maxRight)
			comment.RecordLeft = maxRight + 1
			comment.RecordRight = maxRight + 2
			comment.RecordDept = 0
			comment.ParentID = nil
		}

		return tx.Create(comment).Error
	})
}

func (s *commentService) MoveNested(id uuid.UUID, newParentID *uuid.UUID) error {
	// For brevity, not implemented here. (Can be added if needed.)
	return nil
}

func (s *commentService) DeleteNested(id uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var comment models.Comment
		if err := tx.First(&comment, id).Error; err != nil {
			return err
		}
		if comment.RecordLeft == 0 || comment.RecordRight == 0 {
			return gorm.ErrRecordNotFound
		}

		left := comment.RecordLeft
		right := comment.RecordRight
		width := right - left + 1

		// Delete the comment and all its descendants
		tx.Where("record_left >= ? AND record_right <= ?", left, right).Delete(&models.Comment{})

		// Update the remaining nodes
		tx.Model(&models.Comment{}).
			Where("record_left > ?", right).
			Update("record_left", gorm.Expr("record_left - ?", width))
		tx.Model(&models.Comment{}).
			Where("record_right > ?", right).
			Update("record_right", gorm.Expr("record_right - ?", width))

		return nil
	})
}

func (s *commentService) GetSiblingComments(id uuid.UUID) ([]models.Comment, error) {
	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		return nil, err
	}

	var siblings []models.Comment
	err := database.DB.Where("parent_id = ? AND id != ?", comment.ParentID, id).Find(&siblings).Error
	return siblings, err
}

func (s *commentService) GetParentComment(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		return nil, err
	}

	if comment.ParentID == nil {
		return nil, nil // No parent
	}

	var parent models.Comment
	err := database.DB.First(&parent, comment.ParentID).Error
	if err != nil {
		return nil, err
	}
	return &parent, nil
}

func (s *commentService) GetDescendantComments(id uuid.UUID) ([]models.Comment, error) {
	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		return nil, err
	}

	var descendants []models.Comment
	err := database.DB.Where("record_left > ? AND record_right < ?", comment.RecordLeft, comment.RecordRight).Find(&descendants).Error
	return descendants, err
}

func (s *commentService) GetChildrenComments(id uuid.UUID) ([]models.Comment, error) {
	var children []models.Comment
	err := database.DB.Where("parent_id = ?", id).Find(&children).Error
	return children, err
}

var CommentSvc CommentService = &commentService{}
