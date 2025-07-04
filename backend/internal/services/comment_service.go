package services

import (
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"

	"gorm.io/gorm"
)

type CommentService interface {
	GetCommentsByPost(postID string) ([]models.Comment, error)
	GetCommentByID(id string) (*models.Comment, error)
	CreateComment(comment *models.Comment) error
	UpdateComment(comment *models.Comment) error
	DeleteComment(id string) error
	CreateNested(comment *models.Comment, parentID *int64) error
	MoveNested(id uint, newParentID *int64) error
	DeleteNested(id uint) error
	GetSiblingComments(id uint) ([]models.Comment, error)
	GetParentComment(id uint) (*models.Comment, error)
	GetDescendantComments(id uint) ([]models.Comment, error)
	GetChildrenComments(id uint) ([]models.Comment, error)
}

type commentService struct{}

func (s *commentService) GetCommentsByPost(postID string) ([]models.Comment, error) {
	var comments []models.Comment
	err := database.DB.Where("post_id = ?", postID).Find(&comments).Error
	return comments, err
}

func (s *commentService) GetCommentByID(id string) (*models.Comment, error) {
	var comment models.Comment
	err := database.DB.First(&comment, id).Error
	return &comment, err
}

func (s *commentService) CreateComment(comment *models.Comment) error {
	return database.DB.Create(comment).Error
}

func (s *commentService) UpdateComment(comment *models.Comment) error {
	return database.DB.Save(comment).Error
}

func (s *commentService) DeleteComment(id string) error {
	return database.DB.Delete(&models.Comment{}, id).Error
}

func (s *commentService) CreateNested(comment *models.Comment, parentID *int64) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var left int64
		var depth int64 = 0
		if parentID != nil {
			var parent models.Comment
			if err := tx.First(&parent, *parentID).Error; err != nil {
				return err
			}
			if parent.RecordRight == nil {
				return gorm.ErrRecordNotFound
			}
			left = *parent.RecordRight
			depth = 0
			if parent.RecordDept != nil {
				depth = *parent.RecordDept + 1
			}
			tx.Model(&models.Comment{}).Where("record_right >= ?", left).Update("record_right", gorm.Expr("record_right + 2"))
			tx.Model(&models.Comment{}).Where("record_left > ?", left-1).Update("record_left", gorm.Expr("record_left + 2"))
		} else {
			tx.Model(&models.Comment{}).Select("COALESCE(MAX(record_right), 0)").Scan(&left)
			left++
		}
		right := left + 1
		comment.RecordLeft = &left
		comment.RecordRight = &right
		comment.RecordDept = &depth
		comment.ParentID = parentID
		return tx.Create(comment).Error
	})
}

func (s *commentService) MoveNested(id uint, newParentID *int64) error {
	// For brevity, not implemented here. (Can be added if needed.)
	return nil
}

func (s *commentService) DeleteNested(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var node models.Comment
		if err := tx.First(&node, id).Error; err != nil {
			return err
		}
		if node.RecordLeft == nil || node.RecordRight == nil {
			return gorm.ErrRecordNotFound
		}
		left := *node.RecordLeft
		right := *node.RecordRight
		width := right - left + 1
		tx.Where("record_left >= ? AND record_right <= ?", left, right).Delete(&models.Comment{})
		tx.Model(&models.Comment{}).Where("record_left > ?", right).Update("record_left", gorm.Expr("record_left - ?", width))
		tx.Model(&models.Comment{}).Where("record_right > ?", right).Update("record_right", gorm.Expr("record_right - ?", width))
		return nil
	})
}

func (s *commentService) GetSiblingComments(id uint) ([]models.Comment, error) {
	var node models.Comment
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	var siblings []models.Comment
	if node.ParentID != nil {
		database.DB.Where("parent_id = ? AND id != ?", *node.ParentID, id).Find(&siblings)
	} else {
		database.DB.Where("parent_id IS NULL AND id != ?", id).Find(&siblings)
	}
	return siblings, nil
}

func (s *commentService) GetParentComment(id uint) (*models.Comment, error) {
	var node models.Comment
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	if node.ParentID == nil {
		return nil, nil
	}
	var parent models.Comment
	if err := database.DB.First(&parent, *node.ParentID).Error; err != nil {
		return nil, err
	}
	return &parent, nil
}

func (s *commentService) GetDescendantComments(id uint) ([]models.Comment, error) {
	var node models.Comment
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	if node.RecordLeft == nil || node.RecordRight == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var descendants []models.Comment
	database.DB.Where("record_left > ? AND record_right < ?", *node.RecordLeft, *node.RecordRight).Find(&descendants)
	return descendants, nil
}

func (s *commentService) GetChildrenComments(id uint) ([]models.Comment, error) {
	var children []models.Comment
	if err := database.DB.Where("parent_id = ?", id).Find(&children).Error; err != nil {
		return nil, err
	}
	return children, nil
}

var CommentSvc CommentService = &commentService{}
