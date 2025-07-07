package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"
)

// TagService interface for tag management
type TagService interface {
	CreateTag(ctx context.Context, tag *models.Tag) error
	GetTagByID(ctx context.Context, id uint64) (*models.Tag, error)
	GetTagBySlug(ctx context.Context, slug string) (*models.Tag, error)
	GetTagByName(ctx context.Context, name string) (*models.Tag, error)
	UpdateTag(ctx context.Context, tag *models.Tag) error
	DeleteTag(ctx context.Context, id uint64) error
	GetAllTags(ctx context.Context, tagType string) ([]models.Tag, error)
	GetActiveTags(ctx context.Context) ([]models.Tag, error)
	GetTagsByEntity(ctx context.Context, entityID uint64, entityType string) ([]models.Tag, error)
	AddTagToEntity(ctx context.Context, tagID, entityID uint64, entityType string) error
	RemoveTagFromEntity(ctx context.Context, tagID, entityID uint64, entityType string) error
	GetEntitiesByTag(ctx context.Context, tagID uint64, entityType string, limit, offset int) ([]map[string]interface{}, int64, error)
	SearchTags(ctx context.Context, query string, limit, offset int) ([]models.Tag, int64, error)
	GetTagCount(ctx context.Context) (int64, error)
	InvalidateTagCache(ctx context.Context, tagID uint64) error
}

type tagService struct {
	Redis *redis.RedisService
}

func NewTagService(redisService *redis.RedisService) TagService {
	return &tagService{
		Redis: redisService,
	}
}

// Cache keys
const (
	tagCacheKeyPrefix     = "tag:"
	tagSlugCacheKeyPrefix = "tag:slug:"
	tagNameCacheKeyPrefix = "tag:name:"
	tagAllKeyPrefix       = "tag:all:"
	tagActiveKeyPrefix    = "tag:active:"
	tagEntityKeyPrefix    = "tag:entity:"
	tagSearchKeyPrefix    = "tag:search:"
	tagCountKeyPrefix     = "tag:count:"
)

func (s *tagService) getTagCacheKey(id uint64) string {
	return fmt.Sprintf("%s%d", tagCacheKeyPrefix, id)
}

func (s *tagService) getTagSlugCacheKey(slug string) string {
	return fmt.Sprintf("%s%s", tagSlugCacheKeyPrefix, slug)
}

func (s *tagService) getTagNameCacheKey(name string) string {
	return fmt.Sprintf("%s%s", tagNameCacheKeyPrefix, name)
}

func (s *tagService) getTagAllCacheKey(tagType string) string {
	if tagType == "" {
		return tagAllKeyPrefix + "all"
	}
	return fmt.Sprintf("%s%s", tagAllKeyPrefix, tagType)
}

func (s *tagService) getTagActiveCacheKey() string {
	return tagActiveKeyPrefix + "list"
}

func (s *tagService) getTagEntityCacheKey(entityID uint64, entityType string) string {
	return fmt.Sprintf("%s%d:%s", tagEntityKeyPrefix, entityID, entityType)
}

func (s *tagService) getTagSearchCacheKey(query string, limit, offset int) string {
	return fmt.Sprintf("%s%s:%d:%d", tagSearchKeyPrefix, query, limit, offset)
}

func (s *tagService) getTagCountCacheKey() string {
	return tagCountKeyPrefix + "total"
}

func (s *tagService) CreateTag(ctx context.Context, tag *models.Tag) error {
	if err := database.DB.WithContext(ctx).Create(tag).Error; err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	// Cache the new tag
	if err := s.cacheTag(ctx, tag); err != nil {
		fmt.Printf("Warning: failed to cache tag %d: %v\n", tag.ID, err)
	}

	// Invalidate related caches
	s.invalidateRelatedCaches(ctx)

	return nil
}

func (s *tagService) GetTagByID(ctx context.Context, id uint64) (*models.Tag, error) {
	cacheKey := s.getTagCacheKey(id)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var tag models.Tag
		if err := json.Unmarshal([]byte(cached), &tag); err == nil {
			return &tag, nil
		}
	}

	var tag models.Tag
	if err := database.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// Cache the result
	if err := s.cacheTag(ctx, &tag); err != nil {
		fmt.Printf("Warning: failed to cache tag %d: %v\n", tag.ID, err)
	}

	return &tag, nil
}

func (s *tagService) GetTagBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	cacheKey := s.getTagSlugCacheKey(slug)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var tag models.Tag
		if err := json.Unmarshal([]byte(cached), &tag); err == nil {
			return &tag, nil
		}
	}

	var tag models.Tag
	if err := database.DB.WithContext(ctx).Where("slug = ?", slug).First(&tag).Error; err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// Cache the result
	if err := s.cacheTag(ctx, &tag); err != nil {
		fmt.Printf("Warning: failed to cache tag %d: %v\n", tag.ID, err)
	}

	return &tag, nil
}

func (s *tagService) GetTagByName(ctx context.Context, name string) (*models.Tag, error) {
	cacheKey := s.getTagNameCacheKey(name)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var tag models.Tag
		if err := json.Unmarshal([]byte(cached), &tag); err == nil {
			return &tag, nil
		}
	}

	var tag models.Tag
	if err := database.DB.WithContext(ctx).Where("name = ?", name).First(&tag).Error; err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// Cache the result
	if err := s.cacheTag(ctx, &tag); err != nil {
		fmt.Printf("Warning: failed to cache tag %d: %v\n", tag.ID, err)
	}

	return &tag, nil
}

func (s *tagService) UpdateTag(ctx context.Context, tag *models.Tag) error {
	if err := database.DB.WithContext(ctx).Save(tag).Error; err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	// Update cache
	if err := s.cacheTag(ctx, tag); err != nil {
		fmt.Printf("Warning: failed to cache tag %d: %v\n", tag.ID, err)
	}

	// Invalidate related caches
	s.invalidateRelatedCaches(ctx)

	return nil
}

func (s *tagService) DeleteTag(ctx context.Context, id uint64) error {
	// Get tag first to invalidate related caches
	var tag models.Tag
	if err := database.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	if err := database.DB.WithContext(ctx).Delete(&tag).Error; err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	// Invalidate caches
	s.invalidateTagCaches(ctx, id)
	s.invalidateRelatedCaches(ctx)

	return nil
}

func (s *tagService) GetAllTags(ctx context.Context, tagType string) ([]models.Tag, error) {
	cacheKey := s.getTagAllCacheKey(tagType)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var tags []models.Tag
		if err := json.Unmarshal([]byte(cached), &tags); err == nil {
			return tags, nil
		}
	}

	var tags []models.Tag
	query := database.DB.WithContext(ctx)
	if tagType != "" {
		query = query.Where("type = ?", tagType)
	}

	if err := query.Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(tags); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return tags, nil
}

func (s *tagService) GetActiveTags(ctx context.Context) ([]models.Tag, error) {
	cacheKey := s.getTagActiveCacheKey()

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var tags []models.Tag
		if err := json.Unmarshal([]byte(cached), &tags); err == nil {
			return tags, nil
		}
	}

	var tags []models.Tag
	if err := database.DB.WithContext(ctx).Where("is_active = ?", true).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get active tags: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(tags); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return tags, nil
}

func (s *tagService) GetTagsByEntity(ctx context.Context, entityID uint64, entityType string) ([]models.Tag, error) {
	cacheKey := s.getTagEntityCacheKey(entityID, entityType)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var tags []models.Tag
		if err := json.Unmarshal([]byte(cached), &tags); err == nil {
			return tags, nil
		}
	}

	var taggedEntities []models.TaggedEntity
	if err := database.DB.WithContext(ctx).
		Preload("Tag").
		Where("entity_id = ? AND entity_type = ?", entityID, entityType).
		Find(&taggedEntities).Error; err != nil {
		return nil, fmt.Errorf("failed to get entity tags: %w", err)
	}

	tags := make([]models.Tag, len(taggedEntities))
	for i, te := range taggedEntities {
		tags[i] = te.Tag
	}

	// Cache the result
	if data, err := json.Marshal(tags); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return tags, nil
}

func (s *tagService) AddTagToEntity(ctx context.Context, tagID, entityID uint64, entityType string) error {
	taggedEntity := models.TaggedEntity{
		TagID:      tagID,
		EntityID:   entityID,
		EntityType: entityType,
	}

	if err := database.DB.WithContext(ctx).Create(&taggedEntity).Error; err != nil {
		return fmt.Errorf("failed to add tag to entity: %w", err)
	}

	// Invalidate entity tags cache
	s.Redis.Delete(ctx, s.getTagEntityCacheKey(entityID, entityType))

	return nil
}

func (s *tagService) RemoveTagFromEntity(ctx context.Context, tagID, entityID uint64, entityType string) error {
	if err := database.DB.WithContext(ctx).
		Where("tag_id = ? AND entity_id = ? AND entity_type = ?", tagID, entityID, entityType).
		Delete(&models.TaggedEntity{}).Error; err != nil {
		return fmt.Errorf("failed to remove tag from entity: %w", err)
	}

	// Invalidate entity tags cache
	s.Redis.Delete(ctx, s.getTagEntityCacheKey(entityID, entityType))

	return nil
}

func (s *tagService) GetEntitiesByTag(ctx context.Context, tagID uint64, entityType string, limit, offset int) ([]map[string]interface{}, int64, error) {
	var taggedEntities []models.TaggedEntity
	var total int64

	// Get total count
	if err := database.DB.WithContext(ctx).
		Model(&models.TaggedEntity{}).
		Where("tag_id = ? AND entity_type = ?", tagID, entityType).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tagged entities: %w", err)
	}

	// Get tagged entities with pagination
	if err := database.DB.WithContext(ctx).
		Where("tag_id = ? AND entity_type = ?", tagID, entityType).
		Limit(limit).Offset(offset).
		Find(&taggedEntities).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get tagged entities: %w", err)
	}

	// Convert to map for flexibility
	entities := make([]map[string]interface{}, len(taggedEntities))
	for i, te := range taggedEntities {
		entities[i] = map[string]interface{}{
			"entity_id":   te.EntityID,
			"entity_type": te.EntityType,
			"created_at":  te.CreatedAt,
		}
	}

	return entities, total, nil
}

func (s *tagService) SearchTags(ctx context.Context, query string, limit, offset int) ([]models.Tag, int64, error) {
	cacheKey := s.getTagSearchCacheKey(query, limit, offset)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var result struct {
			Tags  []models.Tag `json:"tags"`
			Total int64        `json:"total"`
		}
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result.Tags, result.Total, nil
		}
	}

	var tags []models.Tag
	var total int64

	// Get total count
	if err := database.DB.WithContext(ctx).
		Model(&models.Tag{}).
		Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%").
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tags: %w", err)
	}

	// Get tags with pagination
	if err := database.DB.WithContext(ctx).
		Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%").
		Limit(limit).Offset(offset).
		Find(&tags).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search tags: %w", err)
	}

	// Cache the result
	result := struct {
		Tags  []models.Tag `json:"tags"`
		Total int64        `json:"total"`
	}{
		Tags:  tags,
		Total: total,
	}

	if data, err := json.Marshal(result); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 15*time.Minute)
	}

	return tags, total, nil
}

func (s *tagService) GetTagCount(ctx context.Context) (int64, error) {
	cacheKey := s.getTagCountCacheKey()

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var count int64
		if _, err := fmt.Sscanf(cached, "%d", &count); err == nil {
			return count, nil
		}
	}

	var count int64
	if err := database.DB.WithContext(ctx).Model(&models.Tag{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count tags: %w", err)
	}

	// Cache the result
	s.Redis.SetWithTTL(ctx, cacheKey, fmt.Sprintf("%d", count), 10*time.Minute)

	return count, nil
}

func (s *tagService) InvalidateTagCache(ctx context.Context, tagID uint64) error {
	s.invalidateTagCaches(ctx, tagID)
	return nil
}

// Helper methods
func (s *tagService) cacheTag(ctx context.Context, tag *models.Tag) error {
	data, err := json.Marshal(tag)
	if err != nil {
		return err
	}

	// Cache by ID
	cacheKey := s.getTagCacheKey(uint64(tag.ID))
	if err := s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute); err != nil {
		return err
	}

	// Cache by slug
	slugCacheKey := s.getTagSlugCacheKey(tag.Slug)
	if err := s.Redis.SetWithTTL(ctx, slugCacheKey, string(data), 30*time.Minute); err != nil {
		return err
	}

	// Cache by name
	nameCacheKey := s.getTagNameCacheKey(tag.Name)
	return s.Redis.SetWithTTL(ctx, nameCacheKey, string(data), 30*time.Minute)
}

func (s *tagService) invalidateTagCaches(ctx context.Context, tagID uint64) {
	cacheKeys := []string{
		s.getTagCacheKey(tagID),
	}

	for _, key := range cacheKeys {
		s.Redis.Delete(ctx, key)
	}
}

func (s *tagService) invalidateRelatedCaches(ctx context.Context) {
	// Invalidate all tags cache and search caches
	patterns := []string{
		tagAllKeyPrefix + "*",
		tagActiveKeyPrefix + "*",
		tagSearchKeyPrefix + "*",
		tagCountKeyPrefix + "*",
	}

	for _, pattern := range patterns {
		s.Redis.DeletePattern(ctx, pattern)
	}
}

// Global tag service instance
var TagSvc TagService
