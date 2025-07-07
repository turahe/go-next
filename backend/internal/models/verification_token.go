package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// VerificationToken represents a verification token for email, phone, or password reset
type VerificationToken struct {
	Base
	UserID    uint      `gorm:"not null;index" json:"user_id" validate:"required"`
	Token     string    `gorm:"uniqueIndex;not null;size:255" json:"token" validate:"required"`
	Type      string    `gorm:"not null;size:50;index" json:"type" validate:"required,oneof=email_verification phone_verification password_reset"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at" validate:"required"`
	Used      bool      `gorm:"default:false;index" json:"used"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName specifies the table name for VerificationToken
func (VerificationToken) TableName() string {
	return "verification_tokens"
}

// BeforeCreate validates verification token data before creation
func (v *VerificationToken) BeforeCreate(tx *gorm.DB) error {
	if err := v.Base.BeforeCreate(tx); err != nil {
		return err
	}
	return v.validate()
}

// BeforeUpdate validates verification token data before update
func (v *VerificationToken) BeforeUpdate(tx *gorm.DB) error {
	if err := v.Base.BeforeUpdate(tx); err != nil {
		return err
	}
	return v.validate()
}

// validate performs validation on verification token fields
func (v *VerificationToken) validate() error {
	if v.UserID == 0 {
		return errors.New("user ID is required")
	}

	if v.Token == "" {
		return errors.New("token cannot be empty")
	}

	if v.Type == "" {
		return errors.New("token type is required")
	}

	validTypes := []string{"email_verification", "phone_verification", "password_reset"}
	typeValid := false
	for _, tokenType := range validTypes {
		if v.Type == tokenType {
			typeValid = true
			break
		}
	}
	if !typeValid {
		return errors.New("invalid token type")
	}

	if v.ExpiresAt.IsZero() {
		return errors.New("expiration time is required")
	}

	return nil
}

// IsExpired checks if the verification token has expired
func (v *VerificationToken) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// IsValid checks if the verification token is valid (not expired and not used)
func (v *VerificationToken) IsValid() bool {
	return !v.IsExpired() && !v.Used
}

// MarkAsUsed marks the verification token as used
func (v *VerificationToken) MarkAsUsed() {
	v.Used = true
}

// GetTimeUntilExpiry returns the time until the verification token expires
func (v *VerificationToken) GetTimeUntilExpiry() time.Duration {
	return v.ExpiresAt.Sub(time.Now())
}

// IsEmailVerification checks if this is an email verification token
func (v *VerificationToken) IsEmailVerification() bool {
	return v.Type == "email_verification"
}

// IsPhoneVerification checks if this is a phone verification token
func (v *VerificationToken) IsPhoneVerification() bool {
	return v.Type == "phone_verification"
}

// IsPasswordReset checks if this is a password reset token
func (v *VerificationToken) IsPasswordReset() bool {
	return v.Type == "password_reset"
}
