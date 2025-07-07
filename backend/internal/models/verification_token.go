package models

import (
	"time"
)

// VerificationToken represents a token for email/phone verification
// Table: verification_tokens

type VerificationToken struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null;size:255" json:"token"`
	Type      string    `gorm:"not null;size:50" json:"type"` // e.g., email, phone
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (VerificationToken) TableName() string {
	return "verification_tokens"
}
