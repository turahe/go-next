package models

import "time"

type VerificationToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:UserID"`
	Token     string    `gorm:"unique;not null"`
	Type      string    `gorm:"not null"` // email_verification, phone_verification, password_reset
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
}
