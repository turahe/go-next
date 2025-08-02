package services

import (
	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/redis"
	"go-next/pkg/storage"
	"mime/multipart"

	"github.com/google/uuid"
)

type MediaService interface {
	UploadAndSaveMedia(file multipart.File, fileHeader *multipart.FileHeader, createdBy *uuid.UUID) (*models.Media, error)
	AssociateMedia(mediaID, mediableID uuid.UUID, mediableType, group string) error
	GetMediaByID(id string) (*models.Media, error)
	DeleteMedia(id string) error
	GetMediaByType(mediaType string) ([]models.Media, error)
}

type mediaService struct {
	Storage      storage.StorageService
	redisService *redis.RedisService
}

func NewMediaService(storageService storage.StorageService, redisService *redis.RedisService) MediaService {
	return &mediaService{
		Storage:      storageService,
		redisService: redisService,
	}
}

func (s *mediaService) UploadAndSaveMedia(file multipart.File, fileHeader *multipart.FileHeader, createdBy *uuid.UUID) (*models.Media, error) {
	key := "media/" + uuid.New().String() + "_" + fileHeader.Filename
	_, err := s.Storage.Put(key, file)
	if err != nil {
		return nil, err
	}
	media := &models.Media{
		UUID:         uuid.New().String(),
		FileName:     key,
		OriginalName: fileHeader.Filename,
		Disk:         "local",
		MimeType:     fileHeader.Header.Get("Content-Type"),
		Size:         int64(fileHeader.Size),
		Path:         key,
		CreatedBy:    createdBy,
	}
	if err := database.DB.Create(media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

func (s *mediaService) AssociateMedia(mediaID, mediableID uuid.UUID, mediableType, group string) error {
	mediable := models.Mediable{
		MediaID:      mediaID,
		MediableID:   mediableID,
		MediableType: mediableType,
		Group:        group,
	}
	return database.DB.Create(&mediable).Error
}

func (s *mediaService) GetMediaByID(id string) (*models.Media, error) {
	mediaID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var media models.Media
	err = database.DB.First(&media, mediaID).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (s *mediaService) DeleteMedia(id string) error {
	mediaID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return database.DB.Delete(&models.Media{}, mediaID).Error
}

func (s *mediaService) GetMediaByType(mediaType string) ([]models.Media, error) {
	var media []models.Media
	err := database.DB.Where("mime_type LIKE ?", mediaType+"%").Find(&media).Error
	return media, err
}

var MediaSvc MediaService = NewMediaService(nil, nil) // Will be initialized properly in startup
