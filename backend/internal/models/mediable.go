package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Mediable represents a polymorphic relationship between media and other models
type Mediable struct {
	Base
	MediaID      uint   `gorm:"not null;index" json:"media_id" validate:"required"`
	MediableID   uint   `gorm:"not null;index" json:"mediable_id" validate:"required"`
	MediableType string `gorm:"not null;size:50;index" json:"mediable_type" validate:"required"`
	Group        string `gorm:"size:50;index" json:"group,omitempty"` // group: featured, gallery, thumbnail, etc.
	Order        int    `gorm:"default:0;index" json:"order,omitempty"`

	// Relationships
	Media Media `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE" json:"media,omitempty"`
}

// TableName specifies the table name for Mediable
func (Mediable) TableName() string {
	return "mediables"
}

// BeforeCreate sets timestamps and validates mediable data
func (m *Mediable) BeforeCreate(tx *gorm.DB) error {
	if err := m.Base.BeforeCreate(tx); err != nil {
		return err
	}

	// Clean fields
	m.MediableType = strings.ToLower(strings.TrimSpace(m.MediableType))
	m.Group = strings.ToLower(strings.TrimSpace(m.Group))

	return m.validate()
}

// BeforeUpdate validates mediable data before update
func (m *Mediable) BeforeUpdate(tx *gorm.DB) error {
	if err := m.Base.BeforeUpdate(tx); err != nil {
		return err
	}

	// Clean fields
	m.MediableType = strings.ToLower(strings.TrimSpace(m.MediableType))
	m.Group = strings.ToLower(strings.TrimSpace(m.Group))

	return m.validate()
}

// validate performs validation on mediable fields
func (m *Mediable) validate() error {
	if m.MediaID == 0 {
		return errors.New("media ID is required")
	}

	if m.MediableID == 0 {
		return errors.New("mediable ID is required")
	}

	if m.MediableType == "" {
		return errors.New("mediable type is required")
	}

	validTypes := []string{"post", "user", "category", "comment"}
	typeValid := false
	for _, mediableType := range validTypes {
		if m.MediableType == mediableType {
			typeValid = true
			break
		}
	}
	if !typeValid {
		return errors.New("invalid mediable type")
	}

	validGroups := []string{"featured", "gallery", "thumbnail", "avatar", "banner", "logo"}
	if m.Group != "" {
		groupValid := false
		for _, group := range validGroups {
			if m.Group == group {
				groupValid = true
				break
			}
		}
		if !groupValid {
			return errors.New("invalid group")
		}
	}

	return nil
}

// IsFeatured checks if this is a featured media
func (m *Mediable) IsFeatured() bool {
	return m.Group == "featured"
}

// IsGallery checks if this is a gallery media
func (m *Mediable) IsGallery() bool {
	return m.Group == "gallery"
}

// IsThumbnail checks if this is a thumbnail media
func (m *Mediable) IsThumbnail() bool {
	return m.Group == "thumbnail"
}

// IsAvatar checks if this is an avatar media
func (m *Mediable) IsAvatar() bool {
	return m.Group == "avatar"
}

// IsBanner checks if this is a banner media
func (m *Mediable) IsBanner() bool {
	return m.Group == "banner"
}

// IsLogo checks if this is a logo media
func (m *Mediable) IsLogo() bool {
	return m.Group == "logo"
}

// HasGroup checks if the mediable has a specific group
func (m *Mediable) HasGroup(group string) bool {
	return m.Group == group
}
