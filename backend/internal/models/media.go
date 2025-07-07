package models

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Media represents a media file in the system
type Media struct {
	BaseWithUser
	UUID            uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"uuid"`
	Hash            string    `gorm:"size:64;index" json:"hash,omitempty"`
	Name            string    `gorm:"size:255;not null" json:"name" validate:"required,min=1,max=255"`
	FileName        string    `gorm:"size:255;not null" json:"file_name" validate:"required,min=1,max=255"`
	Disk            string    `gorm:"size:50;not null;default:'local'" json:"disk" validate:"required,oneof=local s3 gcs"`
	MimeType        string    `gorm:"size:100;not null" json:"mime_type" validate:"required"`
	Size            int64     `gorm:"not null;check:size > 0" json:"size" validate:"required,gt=0"`
	Width           *int      `gorm:"index" json:"width,omitempty"`
	Height          *int      `gorm:"index" json:"height,omitempty"`
	Duration        *int      `gorm:"index" json:"duration,omitempty"` // in seconds
	CustomAttribute string    `gorm:"size:500" json:"custom_attribute,omitempty"`

	// Relationships
	Mediables []Mediable `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE" json:"mediables,omitempty"`
}

// TableName specifies the table name for Media
func (Media) TableName() string {
	return "media"
}

// BeforeCreate sets timestamps and validates media data
func (m *Media) BeforeCreate(tx *gorm.DB) error {
	if err := m.BaseWithUser.BeforeCreate(tx); err != nil {
		return err
	}

	// Generate UUID if not provided
	if m.UUID == uuid.Nil {
		m.UUID = uuid.New()
	}

	// Set default disk
	if m.Disk == "" {
		m.Disk = "local"
	}

	// Clean file name
	m.FileName = strings.TrimSpace(m.FileName)
	m.Name = strings.TrimSpace(m.Name)

	return m.validate()
}

// BeforeUpdate validates media data before update
func (m *Media) BeforeUpdate(tx *gorm.DB) error {
	if err := m.BaseWithUser.BeforeUpdate(tx); err != nil {
		return err
	}

	// Clean file name
	m.FileName = strings.TrimSpace(m.FileName)
	m.Name = strings.TrimSpace(m.Name)

	return m.validate()
}

// validate performs validation on media fields
func (m *Media) validate() error {
	if len(m.Name) < 1 || len(m.Name) > 255 {
		return errors.New("media name must be between 1 and 255 characters")
	}

	if len(m.FileName) < 1 || len(m.FileName) > 255 {
		return errors.New("file name must be between 1 and 255 characters")
	}

	if m.Size <= 0 {
		return errors.New("file size must be greater than 0")
	}

	if m.MimeType == "" {
		return errors.New("mime type is required")
	}

	validDisks := []string{"local", "s3", "gcs"}
	diskValid := false
	for _, disk := range validDisks {
		if m.Disk == disk {
			diskValid = true
			break
		}
	}
	if !diskValid {
		return errors.New("invalid disk type")
	}

	return nil
}

// GetFileExtension returns the file extension
func (m *Media) GetFileExtension() string {
	return strings.ToLower(filepath.Ext(m.FileName))
}

// IsImage checks if the media is an image
func (m *Media) IsImage() bool {
	return strings.HasPrefix(m.MimeType, "image/")
}

// IsVideo checks if the media is a video
func (m *Media) IsVideo() bool {
	return strings.HasPrefix(m.MimeType, "video/")
}

// IsAudio checks if the media is an audio file
func (m *Media) IsAudio() bool {
	return strings.HasPrefix(m.MimeType, "audio/")
}

// IsDocument checks if the media is a document
func (m *Media) IsDocument() bool {
	documentTypes := []string{"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}
	for _, docType := range documentTypes {
		if m.MimeType == docType {
			return true
		}
	}
	return false
}

// GetFileSizeInMB returns the file size in megabytes
func (m *Media) GetFileSizeInMB() float64 {
	return float64(m.Size) / (1024 * 1024)
}

// GetFileSizeInKB returns the file size in kilobytes
func (m *Media) GetFileSizeInKB() float64 {
	return float64(m.Size) / 1024
}

// GetAspectRatio returns the aspect ratio if width and height are available
func (m *Media) GetAspectRatio() float64 {
	if m.Width != nil && m.Height != nil && *m.Height > 0 {
		return float64(*m.Width) / float64(*m.Height)
	}
	return 0
}

// GetDurationInMinutes returns the duration in minutes
func (m *Media) GetDurationInMinutes() float64 {
	if m.Duration != nil {
		return float64(*m.Duration) / 60
	}
	return 0
}

// IsPublic checks if the media is publicly accessible
func (m *Media) IsPublic() bool {
	// Add logic to determine if media is public based on your requirements
	return true
}

// GetStoragePath returns the storage path for this media
func (m *Media) GetStoragePath() string {
	return filepath.Join(m.Disk, m.FileName)
}
