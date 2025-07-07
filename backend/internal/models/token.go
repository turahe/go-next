package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Token represents a stored token in the database
type Token struct {
	Base
	Token        string    `gorm:"not null;size:500;index" json:"token" validate:"required"`
	UserID       uint      `gorm:"not null;index" json:"user_id" validate:"required"`
	ClientSecret string    `gorm:"not null;size:255" json:"client_secret" validate:"required"`
	RefreshToken string    `gorm:"not null;size:500;index" json:"refresh_token" validate:"required"`
	ExpiredAt    time.Time `gorm:"not null;index" json:"expired_at" validate:"required"`
	IsRevoked    bool      `gorm:"default:false;index" json:"is_revoked"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName specifies the table name for Token
func (Token) TableName() string {
	return "tokens"
}

// BeforeCreate validates token data before creation
func (t *Token) BeforeCreate(tx *gorm.DB) error {
	if err := t.Base.BeforeCreate(tx); err != nil {
		return err
	}
	return t.validate()
}

// BeforeUpdate validates token data before update
func (t *Token) BeforeUpdate(tx *gorm.DB) error {
	if err := t.Base.BeforeUpdate(tx); err != nil {
		return err
	}
	return t.validate()
}

// validate performs validation on token fields
func (t *Token) validate() error {
	if t.Token == "" {
		return errors.New("token cannot be empty")
	}

	if t.UserID == 0 {
		return errors.New("user ID is required")
	}

	if t.ClientSecret == "" {
		return errors.New("client secret is required")
	}

	if t.RefreshToken == "" {
		return errors.New("refresh token is required")
	}

	if t.ExpiredAt.IsZero() {
		return errors.New("expiration time is required")
	}

	return nil
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiredAt)
}

// IsValid checks if the token is valid (not expired and not revoked)
func (t *Token) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked
}

// Revoke marks the token as revoked
func (t *Token) Revoke() {
	t.IsRevoked = true
}

// GetTimeUntilExpiry returns the time until the token expires
func (t *Token) GetTimeUntilExpiry() time.Duration {
	return t.ExpiredAt.Sub(time.Now())
}

// JWTKey stores per-user/client JWT signing keys and expiration
type JWTKey struct {
	Base
	UserID          uint   `gorm:"not null;index" json:"user_id" validate:"required"`
	ClientKey       string `gorm:"not null;uniqueIndex;size:255" json:"client_key" validate:"required"`
	SecretKey       string `gorm:"not null;size:255" json:"secret_key" validate:"required"`
	TokenExpiration int    `gorm:"not null;default:3600" json:"token_expiration" validate:"required,gt=0"` // in seconds
	IsActive        bool   `gorm:"default:true;index" json:"is_active"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName specifies the table name for JWTKey
func (JWTKey) TableName() string {
	return "jwt_keys"
}

// BeforeCreate validates JWT key data before creation
func (j *JWTKey) BeforeCreate(tx *gorm.DB) error {
	if err := j.Base.BeforeCreate(tx); err != nil {
		return err
	}

	// Set default values
	if j.TokenExpiration == 0 {
		j.TokenExpiration = 3600
	}
	if j.IsActive == false {
		j.IsActive = true
	}

	return j.validate()
}

// BeforeUpdate validates JWT key data before update
func (j *JWTKey) BeforeUpdate(tx *gorm.DB) error {
	if err := j.Base.BeforeUpdate(tx); err != nil {
		return err
	}
	return j.validate()
}

// validate performs validation on JWT key fields
func (j *JWTKey) validate() error {
	if j.UserID == 0 {
		return errors.New("user ID is required")
	}

	if j.ClientKey == "" {
		return errors.New("client key is required")
	}

	if j.SecretKey == "" {
		return errors.New("secret key is required")
	}

	if j.TokenExpiration <= 0 {
		return errors.New("token expiration must be greater than 0")
	}

	return nil
}

// GetExpirationDuration returns the token expiration as a duration
func (j *JWTKey) GetExpirationDuration() time.Duration {
	return time.Duration(j.TokenExpiration) * time.Second
}

// Deactivate marks the JWT key as inactive
func (j *JWTKey) Deactivate() {
	j.IsActive = false
}
