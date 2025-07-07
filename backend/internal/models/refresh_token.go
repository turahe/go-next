package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents a stored refresh token (legacy - keeping for backward compatibility)
type RefreshToken struct {
	Base
	Token     string    `gorm:"not null;size:500;uniqueIndex" json:"token" validate:"required"`
	UserID    uint      `gorm:"not null;index" json:"user_id" validate:"required"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at" validate:"required"`
	IsRevoked bool      `gorm:"default:false;index" json:"is_revoked"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName specifies the table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// BeforeCreate validates refresh token data before creation
func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if err := r.Base.BeforeCreate(tx); err != nil {
		return err
	}
	return r.validate()
}

// BeforeUpdate validates refresh token data before update
func (r *RefreshToken) BeforeUpdate(tx *gorm.DB) error {
	if err := r.Base.BeforeUpdate(tx); err != nil {
		return err
	}
	return r.validate()
}

// validate performs validation on refresh token fields
func (r *RefreshToken) validate() error {
	if r.Token == "" {
		return errors.New("refresh token cannot be empty")
	}

	if r.UserID == 0 {
		return errors.New("user ID is required")
	}

	if r.ExpiresAt.IsZero() {
		return errors.New("expiration time is required")
	}

	return nil
}

// IsExpired checks if the refresh token has expired
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// IsValid checks if the refresh token is valid (not expired and not revoked)
func (r *RefreshToken) IsValid() bool {
	return !r.IsExpired() && !r.IsRevoked
}

// Revoke marks the refresh token as revoked
func (r *RefreshToken) Revoke() {
	r.IsRevoked = true
}

// GetTimeUntilExpiry returns the time until the refresh token expires
func (r *RefreshToken) GetTimeUntilExpiry() time.Duration {
	return r.ExpiresAt.Sub(time.Now())
}
