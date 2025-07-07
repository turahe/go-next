package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"

	"gorm.io/gorm"
)

type CommentService interface {
	GetCommentsByPost(ctx context.Context, postID string) ([]models.Comment, error)
	GetCommentByID(ctx context.Context, id string) (*models.Comment, error)
	CreateComment(ctx context.Context, comment *models.Comment) error
	UpdateComment(ctx context.Context, comment *models.Comment) error
	DeleteComment(ctx context.Context, id string) error
	CreateNested(ctx context.Context, comment *models.Comment, parentID *int64) error
	MoveNested(ctx context.Context, id uint, newParentID *int64) error
	DeleteNested(ctx context.Context, id uint) error
	GetSiblingComments(ctx context.Context, id uint) ([]models.Comment, error)
	GetParentComment(ctx context.Context, id uint) (*models.Comment, error)
	GetDescendantComments(ctx context.Context, id uint) ([]models.Comment, error)
	GetChildrenComments(ctx context.Context, id uint) ([]models.Comment, error)
	GetCommentsByUser(ctx context.Context, userID uint, limit, offset int) ([]models.Comment, int64, error)
	GetAllComments(ctx context.Context, limit, offset int) ([]models.Comment, int64, error)
	InvalidateCommentCache(ctx context.Context, commentID uint) error
}

type commentService struct {
	Redis *redis.RedisService
}

func NewCommentService(redisService *redis.RedisService) CommentService {
	return &commentService{
		Redis: redisService,
	}
}

// Cache keys
const (
	commentCacheKeyPrefix       = "comment:"
	commentPostKeyPrefix        = "comment:post:"
	commentSiblingsKeyPrefix    = "comment:siblings:"
	commentParentKeyPrefix      = "comment:parent:"
	commentChildrenKeyPrefix    = "comment:children:"
	commentDescendantsKeyPrefix = "comment:descendants:"
	commentUserKeyPrefix        = "comment:user:"
	commentAllKeyPrefix         = "comment:all:"
)

func (s *commentService) getCommentCacheKey(id uint) string {
	return fmt.Sprintf("%s%d", commentCacheKeyPrefix, id)
}

func (s *commentService) getCommentPostCacheKey(postID string) string {
	return fmt.Sprintf("%s%s", commentPostKeyPrefix, postID)
}

func (s *commentService) getCommentSiblingsCacheKey(id uint) string {
	return fmt.Sprintf("%s%d", commentSiblingsKeyPrefix, id)
}

func (s *commentService) getCommentParentCacheKey(id uint) string {
	return fmt.Sprintf("%s%d", commentParentKeyPrefix, id)
}

func (s *commentService) getCommentChildrenCacheKey(id uint) string {
	return fmt.Sprintf("%s%d", commentChildrenKeyPrefix, id)
}

func (s *commentService) getCommentDescendantsCacheKey(id uint) string {
	return fmt.Sprintf("%s%d", commentDescendantsKeyPrefix, id)
}

func (s *commentService) getCommentUserCacheKey(userID uint, limit, offset int) string {
	return fmt.Sprintf("%s%d:%d:%d", commentUserKeyPrefix, userID, limit, offset)
}

func (s *commentService) getCommentAllCacheKey(limit, offset int) string {
	return fmt.Sprintf("%s%d:%d", commentAllKeyPrefix, limit, offset)
}

func (s *commentService) GetCommentsByPost(ctx context.Context, postID string) ([]models.Comment, error) {
	cacheKey := s.getCommentPostCacheKey(postID)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var comments []models.Comment
		if err := json.Unmarshal([]byte(cached), &comments); err == nil {
			return comments, nil
		}
	}

	var comments []models.Comment
	err := database.DB.WithContext(ctx).Where("post_id = ?", postID).Find(&comments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by post: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(comments); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return comments, err
}

func (s *commentService) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	cacheKey := s.getCommentCacheKey(uint(0)) // We'll need to parse the ID properly

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var comment models.Comment
		if err := json.Unmarshal([]byte(cached), &comment); err == nil {
			return &comment, nil
		}
	}

	var comment models.Comment
	err := database.DB.WithContext(ctx).First(&comment, id).Error
	if err != nil {
		return nil, fmt.Errorf("comment not found: %w", err)
	}

	// Cache the result
	if err := s.cacheComment(ctx, &comment); err != nil {
		fmt.Printf("Warning: failed to cache comment %d: %v\n", comment.ID, err)
	}

	return &comment, err
}

func (s *commentService) CreateComment(ctx context.Context, comment *models.Comment) error {
	if err := database.DB.WithContext(ctx).Create(comment).Error; err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	// Cache the new comment
	if err := s.cacheComment(ctx, comment); err != nil {
		fmt.Printf("Warning: failed to cache comment %d: %v\n", comment.ID, err)
	}

	// Invalidate related caches
	s.invalidateRelatedCaches(ctx, comment)

	return nil
}

func (s *commentService) UpdateComment(ctx context.Context, comment *models.Comment) error {
	if err := database.DB.WithContext(ctx).Save(comment).Error; err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	// Update cache
	if err := s.cacheComment(ctx, comment); err != nil {
		fmt.Printf("Warning: failed to cache comment %d: %v\n", comment.ID, err)
	}

	// Invalidate related caches
	s.invalidateRelatedCaches(ctx, comment)

	return nil
}

func (s *commentService) DeleteComment(ctx context.Context, id string) error {
	// Get comment first to invalidate related caches
	var comment models.Comment
	if err := database.DB.WithContext(ctx).First(&comment, id).Error; err != nil {
		return fmt.Errorf("comment not found: %w", err)
	}

	if err := database.DB.WithContext(ctx).Delete(&models.Comment{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	// Invalidate caches
	s.invalidateCommentCaches(ctx, comment.ID)
	s.invalidateRelatedCaches(ctx, &comment)

	return nil
}

func (s *commentService) CreateNested(ctx context.Context, comment *models.Comment, parentID *int64) error {
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var left int64
		var depth int64 = 0
		if parentID != nil {
			var parent models.Comment
			if err := tx.First(&parent, *parentID).Error; err != nil {
				return fmt.Errorf("parent comment not found: %w", err)
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

		if err := tx.Create(comment).Error; err != nil {
			return fmt.Errorf("failed to create nested comment: %w", err)
		}

		// Cache the new comment
		if err := s.cacheComment(ctx, comment); err != nil {
			fmt.Printf("Warning: failed to cache comment %d: %v\n", comment.ID, err)
		}

		// Invalidate parent and sibling caches
		if parentID != nil {
			s.invalidateParentCaches(ctx, uint(*parentID))
		}

		return nil
	})
}

func (s *commentService) MoveNested(ctx context.Context, id uint, newParentID *int64) error {
	// Implementation for moving nested comments
	// This is a complex operation that requires careful handling of the nested set model
	// For now, we'll invalidate all related caches
	s.invalidateAllCommentCaches(ctx)
	return nil
}

func (s *commentService) DeleteNested(ctx context.Context, id uint) error {
	return database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var node models.Comment
		if err := tx.First(&node, id).Error; err != nil {
			return fmt.Errorf("comment not found: %w", err)
		}
		if node.RecordLeft == nil || node.RecordRight == nil {
			return gorm.ErrRecordNotFound
		}
		left := *node.RecordLeft
		right := *node.RecordRight
		width := right - left + 1

		// Get all descendants to invalidate their caches
		var descendants []models.Comment
		tx.Where("record_left >= ? AND record_right <= ?", left, right).Find(&descendants)

		tx.Where("record_left >= ? AND record_right <= ?", left, right).Delete(&models.Comment{})
		tx.Model(&models.Comment{}).Where("record_left > ?", right).Update("record_left", gorm.Expr("record_left - ?", width))
		tx.Model(&models.Comment{}).Where("record_right > ?", right).Update("record_right", gorm.Expr("record_right - ?", width))

		// Invalidate caches for deleted comments and descendants
		for _, descendant := range descendants {
			s.invalidateCommentCaches(ctx, descendant.ID)
		}
		s.invalidateCommentCaches(ctx, id)

		// Invalidate parent caches if exists
		if node.ParentID != nil {
			s.invalidateParentCaches(ctx, uint(*node.ParentID))
		}

		return nil
	})
}

func (s *commentService) GetSiblingComments(ctx context.Context, id uint) ([]models.Comment, error) {
	cacheKey := s.getCommentSiblingsCacheKey(id)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var siblings []models.Comment
		if err := json.Unmarshal([]byte(cached), &siblings); err == nil {
			return siblings, nil
		}
	}

	var node models.Comment
	if err := database.DB.WithContext(ctx).First(&node, id).Error; err != nil {
		return nil, fmt.Errorf("comment not found: %w", err)
	}
	var siblings []models.Comment
	if node.ParentID != nil {
		database.DB.WithContext(ctx).Where("parent_id = ? AND id != ?", *node.ParentID, id).Find(&siblings)
	} else {
		database.DB.WithContext(ctx).Where("parent_id IS NULL AND id != ?", id).Find(&siblings)
	}

	// Cache the result
	if data, err := json.Marshal(siblings); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return siblings, nil
}

func (s *commentService) GetParentComment(ctx context.Context, id uint) (*models.Comment, error) {
	cacheKey := s.getCommentParentCacheKey(id)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var parent models.Comment
		if err := json.Unmarshal([]byte(cached), &parent); err == nil {
			return &parent, nil
		}
	}

	var node models.Comment
	if err := database.DB.WithContext(ctx).First(&node, id).Error; err != nil {
		return nil, fmt.Errorf("comment not found: %w", err)
	}
	if node.ParentID == nil {
		return nil, nil
	}
	var parent models.Comment
	if err := database.DB.WithContext(ctx).First(&parent, *node.ParentID).Error; err != nil {
		return nil, fmt.Errorf("parent comment not found: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(parent); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return &parent, nil
}

func (s *commentService) GetDescendantComments(ctx context.Context, id uint) ([]models.Comment, error) {
	cacheKey := s.getCommentDescendantsCacheKey(id)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var descendants []models.Comment
		if err := json.Unmarshal([]byte(cached), &descendants); err == nil {
			return descendants, nil
		}
	}

	var node models.Comment
	if err := database.DB.WithContext(ctx).First(&node, id).Error; err != nil {
		return nil, fmt.Errorf("comment not found: %w", err)
	}
	if node.RecordLeft == nil || node.RecordRight == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var descendants []models.Comment
	database.DB.WithContext(ctx).Where("record_left > ? AND record_right < ?", *node.RecordLeft, *node.RecordRight).Find(&descendants)

	// Cache the result
	if data, err := json.Marshal(descendants); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return descendants, nil
}

func (s *commentService) GetChildrenComments(ctx context.Context, id uint) ([]models.Comment, error) {
	cacheKey := s.getCommentChildrenCacheKey(id)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var children []models.Comment
		if err := json.Unmarshal([]byte(cached), &children); err == nil {
			return children, nil
		}
	}

	var children []models.Comment
	if err := database.DB.WithContext(ctx).Where("parent_id = ?", id).Find(&children).Error; err != nil {
		return nil, fmt.Errorf("failed to get children comments: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(children); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return children, nil
}

func (s *commentService) GetCommentsByUser(ctx context.Context, userID uint, limit, offset int) ([]models.Comment, int64, error) {
	cacheKey := s.getCommentUserCacheKey(userID, limit, offset)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var result struct {
			Comments []models.Comment `json:"comments"`
			Total    int64            `json:"total"`
		}
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result.Comments, result.Total, nil
		}
	}

	var comments []models.Comment
	var total int64

	// Get total count
	if err := database.DB.WithContext(ctx).Model(&models.Comment{}).Where("created_by = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user comments: %w", err)
	}

	// Get comments with pagination
	if err := database.DB.WithContext(ctx).Where("created_by = ?", userID).Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user comments: %w", err)
	}

	// Cache the result
	result := struct {
		Comments []models.Comment `json:"comments"`
		Total    int64            `json:"total"`
	}{
		Comments: comments,
		Total:    total,
	}

	if data, err := json.Marshal(result); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 15*time.Minute)
	}

	return comments, total, nil
}

func (s *commentService) GetAllComments(ctx context.Context, limit, offset int) ([]models.Comment, int64, error) {
	cacheKey := s.getCommentAllCacheKey(limit, offset)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var result struct {
			Comments []models.Comment `json:"comments"`
			Total    int64            `json:"total"`
		}
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result.Comments, result.Total, nil
		}
	}

	var comments []models.Comment
	var total int64

	// Get total count
	if err := database.DB.WithContext(ctx).Model(&models.Comment{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	// Get comments with pagination
	if err := database.DB.WithContext(ctx).Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get comments: %w", err)
	}

	// Cache the result
	result := struct {
		Comments []models.Comment `json:"comments"`
		Total    int64            `json:"total"`
	}{
		Comments: comments,
		Total:    total,
	}

	if data, err := json.Marshal(result); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 15*time.Minute)
	}

	return comments, total, nil
}

func (s *commentService) InvalidateCommentCache(ctx context.Context, commentID uint) error {
	s.invalidateCommentCaches(ctx, commentID)
	return nil
}

// Helper methods
func (s *commentService) cacheComment(ctx context.Context, comment *models.Comment) error {
	data, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	cacheKey := s.getCommentCacheKey(comment.ID)
	return s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
}

func (s *commentService) invalidateCommentCaches(ctx context.Context, commentID uint) {
	cacheKeys := []string{
		s.getCommentCacheKey(commentID),
		s.getCommentSiblingsCacheKey(commentID),
		s.getCommentParentCacheKey(commentID),
		s.getCommentChildrenCacheKey(commentID),
		s.getCommentDescendantsCacheKey(commentID),
	}

	for _, key := range cacheKeys {
		s.Redis.Delete(ctx, key)
	}
}

func (s *commentService) invalidateParentCaches(ctx context.Context, parentID uint) {
	cacheKeys := []string{
		s.getCommentCacheKey(parentID),
		s.getCommentChildrenCacheKey(parentID),
		s.getCommentDescendantsCacheKey(parentID),
	}

	for _, key := range cacheKeys {
		s.Redis.Delete(ctx, key)
	}
}

func (s *commentService) invalidateRelatedCaches(ctx context.Context, comment *models.Comment) {
	// Invalidate post comments cache
	if comment.PostID != 0 {
		s.Redis.Delete(ctx, s.getCommentPostCacheKey(fmt.Sprintf("%d", comment.PostID)))
	}

	// Invalidate pagination caches
	patterns := []string{
		commentAllKeyPrefix + "*",
		commentUserKeyPrefix + "*",
	}

	for _, pattern := range patterns {
		s.Redis.DeletePattern(ctx, pattern)
	}
}

func (s *commentService) invalidateAllCommentCaches(ctx context.Context) {
	patterns := []string{
		commentCacheKeyPrefix + "*",
		commentPostKeyPrefix + "*",
		commentSiblingsKeyPrefix + "*",
		commentParentKeyPrefix + "*",
		commentChildrenKeyPrefix + "*",
		commentDescendantsKeyPrefix + "*",
		commentUserKeyPrefix + "*",
		commentAllKeyPrefix + "*",
	}

	for _, pattern := range patterns {
		s.Redis.DeletePattern(ctx, pattern)
	}
}

var CommentSvc CommentService = &commentService{}
