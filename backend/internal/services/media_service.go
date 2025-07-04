package services

import (
	"mime/multipart"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaService interface {
	UploadAndSaveMedia(file multipart.File, fileHeader *multipart.FileHeader, createdBy *int64) (*models.Media, error)
	AssociateMedia(mediaID, mediableID uint, mediableType, group string) error
	CreateNested(media *models.Media, parentID *int64) error
	MoveNested(id uint, newParentID *int64) error
	DeleteNested(id uint) error
	GetSiblingMedia(id uint) ([]models.Media, error)
	GetParentMedia(id uint) (*models.Media, error)
	GetDescendantMedia(id uint) ([]models.Media, error)
	GetChildrenMedia(id uint) ([]models.Media, error)
}

type mediaService struct {
	Storage storage.StorageService
}

func NewMediaService(storageService storage.StorageService) MediaService {
	return &mediaService{Storage: storageService}
}

func (s *mediaService) UploadAndSaveMedia(file multipart.File, fileHeader *multipart.FileHeader, createdBy *int64) (*models.Media, error) {
	key := "media/" + uuid.New().String() + "_" + fileHeader.Filename
	url, err := s.Storage.Put(key, file)
	if err != nil {
		return nil, err
	}
	media := &models.Media{
		UUID:      uuid.New(),
		Name:      fileHeader.Filename,
		FileName:  key,
		Disk:      string("custom"),
		MimeType:  fileHeader.Header.Get("Content-Type"),
		Size:      int(fileHeader.Size),
		Url:       url,
		CreatedBy: createdBy,
	}
	if err := database.DB.Create(media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

func (s *mediaService) AssociateMedia(mediaID, mediableID uint, mediableType, group string) error {
	mediable := models.Mediable{
		MediaID:      mediaID,
		MediableID:   mediableID,
		MediableType: mediableType,
		Group:        group,
	}
	return database.DB.Create(&mediable).Error
}

func (s *mediaService) CreateNested(media *models.Media, parentID *int64) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var left int64
		var depth int64 = 0
		if parentID != nil {
			var parent models.Media
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
			tx.Model(&models.Media{}).Where("record_right >= ?", left).Update("record_right", gorm.Expr("record_right + 2"))
			tx.Model(&models.Media{}).Where("record_left > ?", left-1).Update("record_left", gorm.Expr("record_left + 2"))
		} else {
			tx.Model(&models.Media{}).Select("COALESCE(MAX(record_right), 0)").Scan(&left)
			left++
		}
		right := left + 1
		media.RecordLeft = &left
		media.RecordRight = &right
		media.RecordDept = &depth
		media.ParentID = parentID
		return tx.Create(media).Error
	})
}

func (s *mediaService) MoveNested(id uint, newParentID *int64) error {
	// For brevity, not implemented here. (Can be added if needed.)
	return nil
}

func (s *mediaService) DeleteNested(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var node models.Media
		if err := tx.First(&node, id).Error; err != nil {
			return err
		}
		if node.RecordLeft == nil || node.RecordRight == nil {
			return gorm.ErrRecordNotFound
		}
		left := *node.RecordLeft
		right := *node.RecordRight
		width := right - left + 1
		tx.Where("record_left >= ? AND record_right <= ?", left, right).Delete(&models.Media{})
		tx.Model(&models.Media{}).Where("record_left > ?", right).Update("record_left", gorm.Expr("record_left - ?", width))
		tx.Model(&models.Media{}).Where("record_right > ?", right).Update("record_right", gorm.Expr("record_right - ?", width))
		return nil
	})
}

func (s *mediaService) GetSiblingMedia(id uint) ([]models.Media, error) {
	var node models.Media
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	var siblings []models.Media
	if node.ParentID != nil {
		database.DB.Where("parent_id = ? AND id != ?", *node.ParentID, id).Find(&siblings)
	} else {
		database.DB.Where("parent_id IS NULL AND id != ?", id).Find(&siblings)
	}
	return siblings, nil
}

func (s *mediaService) GetParentMedia(id uint) (*models.Media, error) {
	var node models.Media
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	if node.ParentID == nil {
		return nil, nil
	}
	var parent models.Media
	if err := database.DB.First(&parent, *node.ParentID).Error; err != nil {
		return nil, err
	}
	return &parent, nil
}

func (s *mediaService) GetDescendantMedia(id uint) ([]models.Media, error) {
	var node models.Media
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, err
	}
	if node.RecordLeft == nil || node.RecordRight == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var descendants []models.Media
	database.DB.Where("record_left > ? AND record_right < ?", *node.RecordLeft, *node.RecordRight).Find(&descendants)
	return descendants, nil
}

func (s *mediaService) GetChildrenMedia(id uint) ([]models.Media, error) {
	var children []models.Media
	if err := database.DB.Where("parent_id = ?", id).Find(&children).Error; err != nil {
		return nil, err
	}
	return children, nil
}
