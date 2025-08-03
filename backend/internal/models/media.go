package models

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Media represents a media file (image, video, document, etc.)
type Media struct {
	BaseModelWithOrdering
	UUID            string   `json:"uuid" gorm:"uniqueIndex;not null;size:36"`
	FileName        string   `json:"file_name" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	OriginalName    string   `json:"original_name" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	MimeType        string   `json:"mime_type" gorm:"not null;size:100" validate:"required,min=1,max=100"`
	Size            int64    `json:"size" gorm:"not null;check:size > 0" validate:"required,gt=0"`
	Hash            string   `json:"hash" gorm:"size:64;index"`
	Disk            string   `json:"disk" gorm:"default:'local';size:20" validate:"oneof=local s3 gcs"`
	Path            string   `json:"path" gorm:"not null;size:500" validate:"required,min=1,max=500"`
	URL             string   `json:"url" gorm:"size:1000"`
	Width           *int     `json:"width,omitempty"`
	Height          *int     `json:"height,omitempty"`
	Duration        *float64 `json:"duration,omitempty"`
	CustomAttribute string   `json:"custom_attribute" gorm:"type:json"`
	IsPublic        bool     `json:"is_public" gorm:"default:true;index"`

	// Relationships
	Posts []Post `json:"posts,omitempty" gorm:"many2many:mediables;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for Media
func (Media) TableName() string {
	return "media"
}

// BeforeCreate hook for Media
func (m *Media) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.UUID == "" {
		m.UUID = uuid.New().String()
	}
	return nil
}

// BeforeUpdate hook for Media
func (m *Media) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// IsImage checks if the media is an image
func (m *Media) IsImage() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "image"
}

// IsVideo checks if the media is a video
func (m *Media) IsVideo() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "video"
}

// IsAudio checks if the media is an audio file
func (m *Media) IsAudio() bool {
	return m.MimeType != "" && len(m.MimeType) >= 5 && m.MimeType[:5] == "audio"
}

// IsDocument checks if the media is a document
func (m *Media) IsDocument() bool {
	return m.MimeType != "" && (m.MimeType == "application/pdf" ||
		m.MimeType == "application/msword" ||
		m.MimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
}

// GetFileSize returns the file size in a human-readable format
func (m *Media) GetFileSize() string {
	const unit = 1024
	if m.Size < unit {
		return fmt.Sprintf("%d B", m.Size)
	}
	div, exp := int64(unit), 0
	for n := m.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(m.Size)/float64(div), "KMGTPE"[exp])
}

// GetDimensions returns the dimensions as a string
func (m *Media) GetDimensions() string {
	if m.Width != nil && m.Height != nil {
		return fmt.Sprintf("%dx%d", *m.Width, *m.Height)
	}
	return ""
}

// GetIsPublic returns the public status
func (m *Media) GetIsPublic() bool {
	return m.IsPublic
}

// MakePublic makes the media public
func (m *Media) MakePublic() {
	m.IsPublic = true
}

// MakePrivate makes the media private
func (m *Media) MakePrivate() {
	m.IsPublic = false
}
