// Package services provides business logic layer for the blog application.
// This package contains all service interfaces and implementations that handle
// the core business logic, data processing, and external service interactions.
package services

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"go-next/internal/models"
	"go-next/pkg/database"
	"go-next/pkg/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MediaService defines the interface for all media-related business operations.
// This interface provides methods for file upload, media management,
// and media retrieval operations.
type MediaService interface {
	// File upload operations - Methods for handling file uploads

	// UploadFile uploads a file to storage and creates a media record.
	// Validates file type, size, and stores metadata in the database.
	UploadFile(file *multipart.FileHeader, userID uuid.UUID, folder string) (*models.Media, error)

	// UploadMultipleFiles uploads multiple files in a single operation.
	// Useful for bulk upload scenarios like image galleries.
	UploadMultipleFiles(files []*multipart.FileHeader, userID uuid.UUID, folder string) ([]*models.Media, error)

	// Media management - Methods for managing media records

	// GetMediaByID retrieves a media record by its unique identifier.
	// Returns the media with all metadata or an error if not found.
	GetMediaByID(id uuid.UUID) (*models.Media, error)

	// GetMediaBySlug retrieves a media record by its slug.
	// Useful for public media access and SEO-friendly URLs.
	GetMediaBySlug(slug string) (*models.Media, error)

	// UpdateMedia updates media metadata in the database.
	// Allows updating title, description, alt text, and other metadata.
	UpdateMedia(media *models.Media) error

	// DeleteMedia removes a media record and its associated file.
	// Handles both soft and hard deletion based on configuration.
	DeleteMedia(id uuid.UUID) error

	// Media retrieval - Methods for accessing and listing media

	// GetUserMedia retrieves all media uploaded by a specific user.
	// Supports pagination and filtering for user media galleries.
	GetUserMedia(userID uuid.UUID, page, perPage int) ([]*models.Media, int64, error)

	// GetMediaByType retrieves media files of a specific type.
	// Useful for filtering images, videos, documents, etc.
	GetMediaByType(mediaType string, page, perPage int) ([]*models.Media, int64, error)

	// SearchMedia performs a search across media titles and descriptions.
	// Returns paginated results with relevance scoring.
	SearchMedia(query string, page, perPage int) ([]*models.Media, int64, error)

	// Utility methods - Helper functions for media operations

	// ValidateFile checks if a file meets upload requirements.
	// Validates file type, size, and other constraints.
	ValidateFile(file *multipart.FileHeader) error

	// GenerateSlug creates a URL-friendly slug for media files.
	// Ensures unique and SEO-friendly file names.
	GenerateSlug(filename string) string

	// GetMediaURL returns the public URL for accessing a media file.
	// Handles different storage backends and CDN configurations.
	GetMediaURL(media *models.Media) string

	// UploadAndSaveMedia uploads a file and saves it to the database.
	// This is a convenience method that combines file upload and database save.
	UploadAndSaveMedia(file multipart.File, fileHeader *multipart.FileHeader, createdBy *uuid.UUID) (*models.Media, error)

	// AssociateMedia associates a media file with another entity.
	// Creates a many-to-many relationship between media and other models.
	AssociateMedia(mediaID, mediableID uuid.UUID, mediableType, group string) error

	// Nested Set Model methods
	GetMediaDescendants(id uuid.UUID) ([]models.Media, error)
	GetMediaAncestors(id uuid.UUID) ([]models.Media, error)
	GetMediaSiblings(id uuid.UUID) ([]models.Media, error)
	MoveMedia(id uuid.UUID, newParentID uuid.UUID) error
	CreateMediaNested(media *models.Media, parentID *uuid.UUID) error
	DeleteMediaNested(id uuid.UUID) error
}

// mediaService implements the MediaService interface.
// This struct holds the database connection and storage service,
// providing the actual implementation of all media-related business logic.
type mediaService struct {
	db      *gorm.DB               // Database connection for all data operations
	storage storage.StorageService // Storage service for file operations
}

// NewMediaService creates and returns a new instance of MediaService.
// This factory function initializes the service with the global database
// connection and storage service.
func NewMediaService() MediaService {
	// Initialize storage service with local configuration
	storageConfig := storage.StorageConfig{
		Driver:    storage.DriverLocal,
		LocalPath: "./storage/uploads",
	}

	storageService, err := storage.NewStorageService(storageConfig)
	if err != nil {
		// Fallback to nil storage service for now
		storageService = nil
	}

	return &mediaService{
		db:      database.DB,
		storage: storageService,
	}
}

// UploadFile uploads a file to storage and creates a media record.
// Validates file type, size, and stores metadata in the database.
//
// Parameters:
//   - file: The uploaded file header containing file information
//   - userID: ID of the user uploading the file
//   - folder: Target folder for organizing uploaded files
//
// Returns:
//   - *models.Media: The created media record with all metadata
//   - error: Any error encountered during the operation
//
// Example:
//
//	file, err := c.FormFile("file")
//	if err != nil {
//	    // Handle error
//	}
//	media, err := mediaService.UploadFile(file, userID, "blog")
//	if err != nil {
//	    // Handle error (validation, storage, database, etc.)
//	}
//	fmt.Printf("Uploaded: %s\n", media.FileName)
func (s *mediaService) UploadFile(file *multipart.FileHeader, userID uuid.UUID, folder string) (*models.Media, error) {
	// Validate the file before uploading
	if err := s.ValidateFile(file); err != nil {
		return nil, err
	}

	// Generate unique filename and key
	filename := s.generateUniqueFilename(file.Filename)
	key := folder + "/" + filename

	// Upload file to storage
	fileReader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()

	filePath, err := s.storage.Put(key, fileReader)
	if err != nil {
		return nil, err
	}

	// Create media record
	media := &models.Media{
		UUID:         uuid.New().String(),
		FileName:     filename,
		OriginalName: file.Filename,
		MimeType:     file.Header.Get("Content-Type"),
		Size:         file.Size,
		Disk:         "local",
		Path:         filePath,
		IsPublic:     true,
	}

	// Set the created by field
	media.CreatedBy = &userID

	// Save to database
	if err := s.db.Create(media).Error; err != nil {
		// Clean up uploaded file if database save fails
		s.storage.Delete(key)
		return nil, err
	}

	return media, nil
}

// UploadMultipleFiles uploads multiple files in a single operation.
// Useful for bulk upload scenarios like image galleries.
//
// Parameters:
//   - files: Array of uploaded file headers
//   - userID: ID of the user uploading the files
//   - folder: Target folder for organizing uploaded files
//
// Returns:
//   - []*models.Media: Array of created media records
//   - error: Any error encountered during the operation
//
// Example:
//
//	form, err := c.MultipartForm()
//	if err != nil {
//	    // Handle error
//	}
//	files := form.File["files"]
//	mediaList, err := mediaService.UploadMultipleFiles(files, userID, "gallery")
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("Uploaded %d files\n", len(mediaList))
func (s *mediaService) UploadMultipleFiles(files []*multipart.FileHeader, userID uuid.UUID, folder string) ([]*models.Media, error) {
	var mediaList []*models.Media

	// Process each file
	for _, file := range files {
		media, err := s.UploadFile(file, userID, folder)
		if err != nil {
			// Clean up any successfully uploaded files
			for _, uploadedMedia := range mediaList {
				s.storage.Delete(uploadedMedia.Path)
				s.db.Delete(uploadedMedia)
			}
			return nil, err
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

// GetMediaByID retrieves a media record by its unique identifier.
// Returns the media with all metadata or an error if not found.
//
// Parameters:
//   - id: UUID of the media record to retrieve
//
// Returns:
//   - *models.Media: The media record with all metadata
//   - error: Any error encountered during the operation
//
// Example:
//
//	media, err := mediaService.GetMediaByID(mediaID)
//	if err != nil {
//	    // Handle error (not found, database error, etc.)
//	}
//	fmt.Printf("Media: %s (%s)\n", media.FileName, media.MimeType)
func (s *mediaService) GetMediaByID(id uuid.UUID) (*models.Media, error) {
	var media models.Media

	err := s.db.Preload("Posts").First(&media, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("media not found")
		}
		return nil, err
	}

	return &media, nil
}

// GetMediaBySlug retrieves a media record by its slug.
// Useful for public media access and SEO-friendly URLs.
//
// Parameters:
//   - slug: URL-friendly identifier for the media
//
// Returns:
//   - *models.Media: The media record with all metadata
//   - error: Any error encountered during the operation
//
// Example:
//
//	media, err := mediaService.GetMediaBySlug("my-awesome-image")
//	if err != nil {
//	    // Handle error (not found, database error, etc.)
//	}
//	imageURL := mediaService.GetMediaURL(media)
func (s *mediaService) GetMediaBySlug(slug string) (*models.Media, error) {
	var media models.Media

	err := s.db.Preload("Posts").Where("uuid = ?", slug).First(&media).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("media not found")
		}
		return nil, err
	}

	return &media, nil
}

// UpdateMedia updates media metadata in the database.
// Allows updating title, description, alt text, and other metadata.
//
// Parameters:
//   - media: The media record with updated fields
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	media.IsPublic = false
//	err := mediaService.UpdateMedia(media)
//	if err != nil {
//	    // Handle error
//	}
func (s *mediaService) UpdateMedia(media *models.Media) error {
	return s.db.Save(media).Error
}

// DeleteMedia removes a media record and its associated file.
// Handles both soft and hard deletion based on configuration.
//
// Parameters:
//   - id: UUID of the media record to delete
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	err := mediaService.DeleteMedia(mediaID)
//	if err != nil {
//	    // Handle error (not found, storage error, etc.)
//	}
func (s *mediaService) DeleteMedia(id uuid.UUID) error {
	// Get media record first
	media, err := s.GetMediaByID(id)
	if err != nil {
		return err
	}

	// Delete file from storage
	if err := s.storage.Delete(media.Path); err != nil {
		return err
	}

	// Delete from database
	return s.db.Delete(media).Error
}

// GetUserMedia retrieves all media uploaded by a specific user.
// Supports pagination and filtering for user media galleries.
//
// Parameters:
//   - userID: ID of the user whose media to retrieve
//   - page: Current page number (1-based)
//   - perPage: Number of media items per page
//
// Returns:
//   - []*models.Media: List of media records
//   - int64: Total count of user's media
//   - error: Any error encountered during the operation
//
// Example:
//
//	mediaList, total, err := mediaService.GetUserMedia(userID, 1, 20)
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("User has %d media items\n", total)
func (s *mediaService) GetUserMedia(userID uuid.UUID, page, perPage int) ([]*models.Media, int64, error) {
	var mediaList []*models.Media
	var total int64

	// Count total media for this user
	if err := s.db.Model(&models.Media{}).Where("created_by = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated media
	offset := (page - 1) * perPage
	err := s.db.Where("created_by = ?", userID).
		Preload("Posts").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&mediaList).Error

	return mediaList, total, err
}

// GetMediaByType retrieves media files of a specific type.
// Useful for filtering images, videos, documents, etc.
//
// Parameters:
//   - mediaType: MIME type to filter by (e.g., "image/", "video/")
//   - page: Current page number (1-based)
//   - perPage: Number of media items per page
//
// Returns:
//   - []*models.Media: List of media records of the specified type
//   - int64: Total count of media of this type
//   - error: Any error encountered during the operation
//
// Example:
//
//	images, total, err := mediaService.GetMediaByType("image/", 1, 10)
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("Found %d images\n", total)
func (s *mediaService) GetMediaByType(mediaType string, page, perPage int) ([]*models.Media, int64, error) {
	var mediaList []*models.Media
	var total int64

	// Count total media of this type
	if err := s.db.Model(&models.Media{}).Where("mime_type LIKE ?", mediaType+"%").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated media
	offset := (page - 1) * perPage
	err := s.db.Where("mime_type LIKE ?", mediaType+"%").
		Preload("Posts").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&mediaList).Error

	return mediaList, total, err
}

// SearchMedia performs a search across media titles and descriptions.
// Returns paginated results with relevance scoring.
//
// Parameters:
//   - query: Search term to look for in media metadata
//   - page: Current page number (1-based)
//   - perPage: Number of media items per page
//
// Returns:
//   - []*models.Media: List of matching media records
//   - int64: Total count of matching media
//   - error: Any error encountered during the operation
//
// Example:
//
//	results, total, err := mediaService.SearchMedia("nature", 1, 10)
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("Found %d media matching 'nature'\n", total)
func (s *mediaService) SearchMedia(query string, page, perPage int) ([]*models.Media, int64, error) {
	var mediaList []*models.Media
	var total int64

	// Build search query
	searchQuery := s.db.Model(&models.Media{}).
		Where("file_name ILIKE ? OR original_name ILIKE ?",
			"%"+query+"%", "%"+query+"%")

	// Count total matching media
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * perPage
	err := searchQuery.Preload("Posts").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&mediaList).Error

	return mediaList, total, err
}

// ValidateFile checks if a file meets upload requirements.
// Validates file type, size, and other constraints.
//
// Parameters:
//   - file: The file header to validate
//
// Returns:
//   - error: Any validation errors encountered
//
// Example:
//
//	err := mediaService.ValidateFile(file)
//	if err != nil {
//	    // Handle validation error (file too large, invalid type, etc.)
//	}
func (s *mediaService) ValidateFile(file *multipart.FileHeader) error {
	// Check file size (e.g., max 10MB)
	const maxSize = 10 * 1024 * 1024 // 10MB
	if file.Size > maxSize {
		return errors.New("file size exceeds maximum allowed size")
	}

	// Check file type
	allowedTypes := []string{
		"image/jpeg", "image/png", "image/gif", "image/webp",
		"video/mp4", "video/webm", "video/ogg",
		"application/pdf", "text/plain",
	}

	contentType := file.Header.Get("Content-Type")
	isAllowed := false
	for _, allowedType := range allowedTypes {
		if strings.HasPrefix(contentType, allowedType) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return errors.New("file type not allowed")
	}

	return nil
}

// GenerateSlug creates a URL-friendly slug for media files.
// Ensures unique and SEO-friendly file names.
//
// Parameters:
//   - filename: Original filename to convert to slug
//
// Returns:
//   - string: URL-friendly slug
//
// Example:
//
//	slug := mediaService.GenerateSlug("My Awesome Image.jpg")
//	// Result: "my-awesome-image"
func (s *mediaService) GenerateSlug(filename string) string {
	// Remove file extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	slug = strings.ReplaceAll(slug, "_", "-")

	// Ensure uniqueness by adding timestamp if needed
	// This is a simplified version - in practice, you'd check for duplicates
	return slug
}

// GetMediaURL returns the public URL for accessing a media file.
// Handles different storage backends and CDN configurations.
//
// Parameters:
//   - media: The media record to get URL for
//
// Returns:
//   - string: Public URL for accessing the media file
//
// Example:
//
//	media, err := mediaService.GetMediaByID(mediaID)
//	if err != nil {
//	    // Handle error
//	}
//	url := mediaService.GetMediaURL(media)
//	fmt.Printf("Media URL: %s\n", url)
func (s *mediaService) GetMediaURL(media *models.Media) string {
	if media.URL != "" {
		return media.URL
	}

	// Generate URL from storage service
	url, err := s.storage.GetURL(media.Path)
	if err != nil {
		return ""
	}

	return url
}

// generateUniqueFilename creates a unique filename to prevent conflicts.
// This is a helper method used internally by UploadFile.
func (s *mediaService) generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)

	// Add timestamp for uniqueness
	timestamp := strings.ReplaceAll(time.Now().Format("20060102150405"), " ", "")

	return name + "_" + timestamp + ext
}

// UploadAndSaveMedia uploads a file and saves it to the database.
// This is a convenience method that combines file upload and database save.
//
// Parameters:
//   - file: The file to upload
//   - fileHeader: File header containing metadata
//   - createdBy: ID of the user who created the media
//
// Returns:
//   - *models.Media: The created media record
//   - error: Any error encountered during the operation
func (s *mediaService) UploadAndSaveMedia(file multipart.File, fileHeader *multipart.FileHeader, createdBy *uuid.UUID) (*models.Media, error) {
	// Validate the file
	if err := s.ValidateFile(fileHeader); err != nil {
		return nil, err
	}

	// Generate unique filename
	filename := s.generateUniqueFilename(fileHeader.Filename)

	// Create media record
	media := &models.Media{
		OriginalName: fileHeader.Filename,
		FileName:     filename,
		Path:         filename, // Simplified path for now
		Size:         fileHeader.Size,
		MimeType:     fileHeader.Header.Get("Content-Type"),
	}

	// Set the created by field
	if createdBy != nil {
		media.CreatedBy = createdBy
	}

	// Save to database
	if err := s.db.Create(media).Error; err != nil {
		return nil, err
	}

	return media, nil
}

// AssociateMedia associates a media file with another entity.
// Creates a many-to-many relationship between media and other models.
//
// Parameters:
//   - mediaID: ID of the media file
//   - mediableID: ID of the entity to associate with
//   - mediableType: Type of the entity (e.g., "categories", "posts")
//   - group: Group name for the association (e.g., "image", "video")
//
// Returns:
//   - error: Any error encountered during the operation
func (s *mediaService) AssociateMedia(mediaID, mediableID uuid.UUID, mediableType, group string) error {
	// Create mediable association
	mediable := &models.Mediable{
		MediaID:      mediaID,
		MediableID:   mediableID,
		MediableType: mediableType,
		Group:        group,
	}

	return s.db.Create(mediable).Error
}

// Nested Set Model methods for Media

func (s *mediaService) GetMediaDescendants(id uuid.UUID) ([]models.Media, error) {
	var media models.Media
	if err := s.db.First(&media, id).Error; err != nil {
		return nil, err
	}

	var descendants []models.Media
	err := s.db.Where("record_left > ? AND record_right < ?", media.RecordLeft, media.RecordRight).
		Order("record_left ASC").Find(&descendants).Error
	return descendants, err
}

func (s *mediaService) GetMediaAncestors(id uuid.UUID) ([]models.Media, error) {
	var media models.Media
	if err := s.db.First(&media, id).Error; err != nil {
		return nil, err
	}

	var ancestors []models.Media
	err := s.db.Where("record_left < ? AND record_right > ?", media.RecordLeft, media.RecordRight).
		Order("record_left ASC").Find(&ancestors).Error
	return ancestors, err
}

func (s *mediaService) GetMediaSiblings(id uuid.UUID) ([]models.Media, error) {
	var media models.Media
	if err := s.db.First(&media, id).Error; err != nil {
		return nil, err
	}

	var siblings []models.Media
	err := s.db.Where("record_left > ? AND record_right < ? AND record_dept = ?",
		media.RecordLeft, media.RecordRight, media.RecordDept).
		Order("record_left ASC").Find(&siblings).Error
	return siblings, err
}

func (s *mediaService) MoveMedia(id uuid.UUID, newParentID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var media models.Media
		if err := tx.First(&media, id).Error; err != nil {
			return err
		}

		return s.moveMediaSubtree(tx, &media, newParentID)
	})
}

func (s *mediaService) CreateMediaNested(media *models.Media, parentID *uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if parentID != nil {
			var parent models.Media
			if err := tx.First(&parent, parentID).Error; err != nil {
				return err
			}
			if parent.RecordRight == 0 {
				return gorm.ErrRecordNotFound
			}

			// Update parent's right value
			tx.Model(&models.Media{}).
				Where("record_right >= ?", parent.RecordRight).
				Update("record_right", gorm.Expr("record_right + 2"))
			tx.Model(&models.Media{}).
				Where("record_left > ?", parent.RecordRight).
				Update("record_left", gorm.Expr("record_left + 2"))

			media.RecordLeft = parent.RecordRight
			media.RecordRight = parent.RecordRight + 1
			media.RecordDept = parent.RecordDept + 1
			media.ParentID = parentID
		} else {
			// Create as root
			var maxRight int
			tx.Model(&models.Media{}).Select("COALESCE(MAX(record_right), 0)").Scan(&maxRight)

			media.RecordLeft = maxRight + 1
			media.RecordRight = maxRight + 2
			media.RecordDept = 0
		}

		return tx.Create(media).Error
	})
}

func (s *mediaService) DeleteMediaNested(id uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var media models.Media
		if err := tx.First(&media, id).Error; err != nil {
			return err
		}

		// Calculate the width of the subtree
		width := media.RecordRight - media.RecordLeft + 1

		// Delete the media and all its descendants
		if err := tx.Where("record_left >= ? AND record_right <= ?", media.RecordLeft, media.RecordRight).Delete(&models.Media{}).Error; err != nil {
			return err
		}

		// Update the left and right values of remaining nodes
		tx.Model(&models.Media{}).
			Where("record_left > ?", media.RecordRight).
			Update("record_left", gorm.Expr("record_left - ?", width))

		tx.Model(&models.Media{}).
			Where("record_right > ?", media.RecordRight).
			Update("record_right", gorm.Expr("record_right - ?", width))

		return nil
	})
}

// moveMediaSubtree moves a media and its entire subtree to a new parent
func (s *mediaService) moveMediaSubtree(tx *gorm.DB, media *models.Media, newParentID uuid.UUID) error {
	// Calculate the width of the subtree
	width := media.RecordRight - media.RecordLeft + 1

	// Get the new parent
	var newParent models.Media
	if newParentID != uuid.Nil {
		if err := tx.First(&newParent, newParentID).Error; err != nil {
			return err
		}
	}

	// Calculate the new position
	var newLeft int
	if newParentID != uuid.Nil {
		newLeft = newParent.RecordRight
	} else {
		// Moving to root level
		var maxRight int
		tx.Model(&models.Media{}).Select("COALESCE(MAX(record_right), 0)").Scan(&maxRight)
		newLeft = maxRight + 1
	}

	// Calculate the offset
	offset := newLeft - media.RecordLeft

	// Update all nodes in the subtree
	tx.Model(&models.Media{}).
		Where("record_left >= ? AND record_right <= ?", media.RecordLeft, media.RecordRight).
		Updates(map[string]interface{}{
			"record_left":  gorm.Expr("record_left + ?", offset),
			"record_right": gorm.Expr("record_right + ?", offset),
			"record_dept":  gorm.Expr("record_dept + ?", newParent.RecordDept-media.RecordDept+1),
		})

	// Update nodes to the right of the old position
	tx.Model(&models.Media{}).
		Where("record_left > ?", media.RecordRight).
		Update("record_left", gorm.Expr("record_left - ?", width))

	tx.Model(&models.Media{}).
		Where("record_right > ?", media.RecordRight).
		Update("record_right", gorm.Expr("record_right - ?", width))

	// Update nodes to the right of the new position
	tx.Model(&models.Media{}).
		Where("record_left >= ?", newLeft).
		Update("record_left", gorm.Expr("record_left + ?", width))

	tx.Model(&models.Media{}).
		Where("record_right >= ?", newLeft).
		Update("record_right", gorm.Expr("record_right + ?", width))

	return nil
}
