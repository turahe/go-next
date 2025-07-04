package models

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents a stored refresh token (legacy - keeping for backward compatibility)
type RefreshToken struct {
	Token     string    // The refresh token string itself
	UserID    uint      // Who the token belongs to
	ExpiresAt time.Time // When the refresh token expires
	gorm.Model
}

func (m *RefreshToken) BeforeCreate(*gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *RefreshToken) BeforeUpdate(*gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
