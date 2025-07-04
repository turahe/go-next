package models

import (
	"time"

	"gorm.io/gorm"
)

// Token represents a stored token in the database
type Token struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Token        string    `json:"token" gorm:"not null"`         // The access token string
	UserID       uint      `json:"user_id" gorm:"not null"`       // Who the token belongs to
	ClientSecret string    `json:"client_secret" gorm:"not null"` // Client secret for additional security
	RefreshToken string    `json:"refresh_token" gorm:"not null"` // The refresh token string
	ExpiredAt    time.Time `json:"expired_at"`                    // When the token expires
	gorm.Model
}

// JWTKey stores per-user/client JWT signing keys and expiration
type JWTKey struct {
	ID              uint   `gorm:"primaryKey"`
	UserID          uint   `gorm:"not null;index"`
	ClientKey       string `gorm:"not null;uniqueIndex"`
	SecretKey       string `gorm:"not null"`
	TokenExpiration int    `gorm:"not null;default:3600"` // in seconds
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (m *Token) BeforeCreate(*gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Token) BeforeUpdate(*gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
