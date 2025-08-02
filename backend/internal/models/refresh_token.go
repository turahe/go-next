package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshToken represents a refresh token for authentication
type RefreshToken struct {
	BaseModel
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index" validate:"required"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null;size:255" validate:"required,min=1,max=255"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index" validate:"required"`
	IsActive  bool      `json:"is_active" gorm:"default:true;index"`
	IPAddress string    `json:"ip_address" gorm:"size:45"`
	UserAgent string    `json:"user_agent" gorm:"size:500"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// BeforeCreate hook for RefreshToken
func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for RefreshToken
func (r *RefreshToken) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// IsExpired checks if the refresh token is expired
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// IsValid checks if the refresh token is valid
func (r *RefreshToken) IsValid() bool {
	return r.IsActive && !r.IsExpired()
}
