package services

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/redis"
	"wordpress-go-next/backend/pkg/storage"

	"github.com/google/uuid"
)

type MediaService interface {
	UploadAndSaveMedia(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, createdBy *int64) (*models.Media, error)
	AssociateMedia(ctx context.Context, mediaID, mediableID uint, mediableType, group string) error
	GetMediaByID(ctx context.Context, id uint) (*models.Media, error)
	GetMediaByUUID(ctx context.Context, uuid string) (*models.Media, error)
	GetMediaByMediable(ctx context.Context, mediableID uint, mediableType string) ([]models.Media, error)
	UpdateMedia(ctx context.Context, media *models.Media) error
	DeleteMedia(ctx context.Context, id uint) error
	GetAllMedia(ctx context.Context, limit, offset int) ([]models.Media, int64, error)
	GetMediaByUser(ctx context.Context, userID uint, limit, offset int) ([]models.Media, int64, error)
	InvalidateMediaCache(ctx context.Context, mediaID uint) error
}

type mediaService struct {
	Storage storage.StorageService
	Redis   *redis.RedisService
}

func NewMediaService(storageService storage.StorageService, redisService *redis.RedisService) MediaService {
	return &mediaService{
		Storage: storageService,
		Redis:   redisService,
	}
}

// Cache keys
const (
	mediaCacheKeyPrefix     = "media:"
	mediaUUIDCacheKeyPrefix = "media:uuid:"
	mediaMediableKeyPrefix  = "media:mediable:"
	mediaUserKeyPrefix      = "media:user:"
	mediaAllKeyPrefix       = "media:all:"
)

func (s *mediaService) getMediaCacheKey(id uint) string {
	return fmt.Sprintf("%s%d", mediaCacheKeyPrefix, id)
}

func (s *mediaService) getMediaUUIDCacheKey(uuid string) string {
	return fmt.Sprintf("%s%s", mediaUUIDCacheKeyPrefix, uuid)
}

func (s *mediaService) getMediaMediableCacheKey(mediableID uint, mediableType string) string {
	return fmt.Sprintf("%s%d:%s", mediaMediableKeyPrefix, mediableID, mediableType)
}

func (s *mediaService) getMediaUserCacheKey(userID uint, limit, offset int) string {
	return fmt.Sprintf("%s%d:%d:%d", mediaUserKeyPrefix, userID, limit, offset)
}

func (s *mediaService) getMediaAllCacheKey(limit, offset int) string {
	return fmt.Sprintf("%s%d:%d", mediaAllKeyPrefix, limit, offset)
}

func (s *mediaService) UploadAndSaveMedia(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, createdBy *int64) (*models.Media, error) {
	key := "media/" + uuid.New().String() + "_" + fileHeader.Filename
	_, err := s.Storage.Put(key, file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	media := &models.Media{
		UUID:     uuid.New(),
		Name:     fileHeader.Filename,
		FileName: key,
		Disk:     "custom",
		MimeType: fileHeader.Header.Get("Content-Type"),
		Size:     fileHeader.Size,
	}
	if createdBy != nil {
		userID := uint(*createdBy)
		media.CreatedBy = &userID
	}

	if err := database.DB.WithContext(ctx).Create(media).Error; err != nil {
		return nil, fmt.Errorf("failed to save media: %w", err)
	}

	// Cache the new media
	if err := s.cacheMedia(ctx, media); err != nil {
		fmt.Printf("Warning: failed to cache media %d: %v\n", media.ID, err)
	}

	// Invalidate related caches
	s.invalidateRelatedCaches(ctx)

	return media, nil
}

func (s *mediaService) AssociateMedia(ctx context.Context, mediaID, mediableID uint, mediableType, group string) error {
	mediable := models.Mediable{
		MediaID:      mediaID,
		MediableID:   mediableID,
		MediableType: mediableType,
		Group:        group,
	}

	if err := database.DB.WithContext(ctx).Create(&mediable).Error; err != nil {
		return fmt.Errorf("failed to associate media: %w", err)
	}

	// Invalidate related caches
	cacheKeys := []string{
		s.getMediaMediableCacheKey(mediableID, mediableType),
		s.getMediaCacheKey(mediaID),
	}

	for _, key := range cacheKeys {
		if err := s.Redis.Delete(ctx, key); err != nil {
			fmt.Printf("Warning: failed to invalidate cache key %s: %v\n", key, err)
		}
	}

	return nil
}

func (s *mediaService) GetMediaByID(ctx context.Context, id uint) (*models.Media, error) {
	cacheKey := s.getMediaCacheKey(id)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var media models.Media
		if err := json.Unmarshal([]byte(cached), &media); err == nil {
			return &media, nil
		}
	}

	var media models.Media
	if err := database.DB.WithContext(ctx).First(&media, id).Error; err != nil {
		return nil, fmt.Errorf("media not found: %w", err)
	}

	// Cache the result
	if err := s.cacheMedia(ctx, &media); err != nil {
		fmt.Printf("Warning: failed to cache media %d: %v\n", media.ID, err)
	}

	return &media, nil
}

func (s *mediaService) GetMediaByUUID(ctx context.Context, uuid string) (*models.Media, error) {
	cacheKey := s.getMediaUUIDCacheKey(uuid)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var media models.Media
		if err := json.Unmarshal([]byte(cached), &media); err == nil {
			return &media, nil
		}
	}

	var media models.Media
	if err := database.DB.WithContext(ctx).Where("uuid = ?", uuid).First(&media).Error; err != nil {
		return nil, fmt.Errorf("media not found: %w", err)
	}

	// Cache the result
	if err := s.cacheMedia(ctx, &media); err != nil {
		fmt.Printf("Warning: failed to cache media %d: %v\n", media.ID, err)
	}

	return &media, nil
}

func (s *mediaService) GetMediaByMediable(ctx context.Context, mediableID uint, mediableType string) ([]models.Media, error) {
	cacheKey := s.getMediaMediableCacheKey(mediableID, mediableType)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var media []models.Media
		if err := json.Unmarshal([]byte(cached), &media); err == nil {
			return media, nil
		}
	}

	var media []models.Media
	err := database.DB.WithContext(ctx).
		Joins("JOIN mediables ON media.id = mediables.media_id").
		Where("mediables.mediable_id = ? AND mediables.mediable_type = ?", mediableID, mediableType).
		Find(&media).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get media by mediable: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(media); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return media, nil
}

func (s *mediaService) UpdateMedia(ctx context.Context, media *models.Media) error {
	if err := database.DB.WithContext(ctx).Save(media).Error; err != nil {
		return fmt.Errorf("failed to update media: %w", err)
	}

	// Update cache
	if err := s.cacheMedia(ctx, media); err != nil {
		fmt.Printf("Warning: failed to cache media %d: %v\n", media.ID, err)
	}

	// Invalidate related caches
	s.invalidateRelatedCaches(ctx)

	return nil
}

func (s *mediaService) DeleteMedia(ctx context.Context, id uint) error {
	// Get media first to invalidate related caches
	var media models.Media
	if err := database.DB.WithContext(ctx).First(&media, id).Error; err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	if err := database.DB.WithContext(ctx).Delete(&media).Error; err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}

	// Invalidate caches
	s.invalidateMediaCaches(ctx, id)
	s.invalidateRelatedCaches(ctx)

	return nil
}

func (s *mediaService) GetAllMedia(ctx context.Context, limit, offset int) ([]models.Media, int64, error) {
	cacheKey := s.getMediaAllCacheKey(limit, offset)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var result struct {
			Media []models.Media `json:"media"`
			Total int64          `json:"total"`
		}
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result.Media, result.Total, nil
		}
	}

	var media []models.Media
	var total int64

	// Get total count
	if err := database.DB.WithContext(ctx).Model(&models.Media{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count media: %w", err)
	}

	// Get media with pagination
	if err := database.DB.WithContext(ctx).Limit(limit).Offset(offset).Find(&media).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get media: %w", err)
	}

	// Cache the result
	result := struct {
		Media []models.Media `json:"media"`
		Total int64          `json:"total"`
	}{
		Media: media,
		Total: total,
	}

	if data, err := json.Marshal(result); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 15*time.Minute)
	}

	return media, total, nil
}

func (s *mediaService) GetMediaByUser(ctx context.Context, userID uint, limit, offset int) ([]models.Media, int64, error) {
	cacheKey := s.getMediaUserCacheKey(userID, limit, offset)

	// Try to get from cache first
	if cached, err := s.Redis.Get(ctx, cacheKey); err == nil {
		var result struct {
			Media []models.Media `json:"media"`
			Total int64          `json:"total"`
		}
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result.Media, result.Total, nil
		}
	}

	var media []models.Media
	var total int64

	// Get total count
	if err := database.DB.WithContext(ctx).Model(&models.Media{}).Where("created_by = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user media: %w", err)
	}

	// Get media with pagination
	if err := database.DB.WithContext(ctx).Where("created_by = ?", userID).Limit(limit).Offset(offset).Find(&media).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user media: %w", err)
	}

	// Cache the result
	result := struct {
		Media []models.Media `json:"media"`
		Total int64          `json:"total"`
	}{
		Media: media,
		Total: total,
	}

	if data, err := json.Marshal(result); err == nil {
		s.Redis.SetWithTTL(ctx, cacheKey, string(data), 15*time.Minute)
	}

	return media, total, nil
}

func (s *mediaService) InvalidateMediaCache(ctx context.Context, mediaID uint) error {
	s.invalidateMediaCaches(ctx, mediaID)
	return nil
}

// Helper methods
func (s *mediaService) cacheMedia(ctx context.Context, media *models.Media) error {
	data, err := json.Marshal(media)
	if err != nil {
		return err
	}

	// Cache by ID
	cacheKey := s.getMediaCacheKey(media.ID)
	if err := s.Redis.SetWithTTL(ctx, cacheKey, string(data), 30*time.Minute); err != nil {
		return err
	}

	// Cache by UUID
	uuidCacheKey := s.getMediaUUIDCacheKey(media.UUID.String())
	return s.Redis.SetWithTTL(ctx, uuidCacheKey, string(data), 30*time.Minute)
}

func (s *mediaService) invalidateMediaCaches(ctx context.Context, mediaID uint) {
	cacheKeys := []string{
		s.getMediaCacheKey(mediaID),
	}

	for _, key := range cacheKeys {
		s.Redis.Delete(ctx, key)
	}
}

func (s *mediaService) invalidateRelatedCaches(ctx context.Context) {
	// Invalidate pagination caches
	patterns := []string{
		mediaAllKeyPrefix + "*",
		mediaUserKeyPrefix + "*",
	}

	for _, pattern := range patterns {
		s.Redis.DeletePattern(ctx, pattern)
	}
}
