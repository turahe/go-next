package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VerificationTokenType represents the type of verification token
type VerificationTokenType string

const (
	EmailVerification   VerificationTokenType = "email_verification"
	PhoneVerification   VerificationTokenType = "phone_verification"
	PasswordReset       VerificationTokenType = "password_reset"
	TwoFactorAuth       VerificationTokenType = "two_factor_auth"
	AccountDeactivation VerificationTokenType = "account_deactivation"
)

// VerificationToken represents a verification token for various user actions
type VerificationToken struct {
	BaseModel
	UserID    uuid.UUID             `json:"user_id" gorm:"type:uuid;not null;index" validate:"required"`
	Token     string                `json:"token" gorm:"uniqueIndex;not null;size:255" validate:"required,min=1,max=255"`
	Type      VerificationTokenType `json:"type" gorm:"not null;size:50;index" validate:"required,oneof=email_verification phone_verification password_reset two_factor_auth account_deactivation"`
	ExpiresAt time.Time             `json:"expires_at" gorm:"not null;index" validate:"required"`
	Used      bool                  `json:"used" gorm:"default:false;index"`
	IPAddress string                `json:"ip_address" gorm:"size:45"`
	UserAgent string                `json:"user_agent" gorm:"size:500"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for VerificationToken
func (VerificationToken) TableName() string {
	return "verification_tokens"
}

// BeforeCreate hook for VerificationToken
func (v *VerificationToken) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for VerificationToken
func (v *VerificationToken) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// IsExpired checks if the verification token is expired
func (v *VerificationToken) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// IsValid checks if the verification token is valid
func (v *VerificationToken) IsValid() bool {
	return !v.Used && !v.IsExpired()
}

// MarkAsUsed marks the verification token as used
func (v *VerificationToken) MarkAsUsed() {
	v.Used = true
}

// IsEmailVerification checks if this is an email verification token
func (v *VerificationToken) IsEmailVerification() bool {
	return v.Type == EmailVerification
}

// IsPhoneVerification checks if this is a phone verification token
func (v *VerificationToken) IsPhoneVerification() bool {
	return v.Type == PhoneVerification
}

// IsPasswordReset checks if this is a password reset token
func (v *VerificationToken) IsPasswordReset() bool {
	return v.Type == PasswordReset
}
