package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Token represents an API token
type Token struct {
	BaseModel
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index" validate:"required"`
	Token      string     `json:"token" gorm:"uniqueIndex;not null;size:255" validate:"required,min=1,max=255"`
	Type       string     `json:"type" gorm:"not null;size:20" validate:"required,oneof=access refresh"`
	ExpiredAt  *time.Time `json:"expired_at,omitempty" gorm:"index"`
	IsActive   bool       `json:"is_active" gorm:"default:true;index"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	IPAddress  string     `json:"ip_address" gorm:"size:45"`
	UserAgent  string     `json:"user_agent" gorm:"size:500"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for Token
func (Token) TableName() string {
	return "tokens"
}

// BeforeCreate hook for Token
func (t *Token) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for Token
func (t *Token) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// IsExpired checks if the token is expired
func (t *Token) IsExpired() bool {
	if t.ExpiredAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiredAt)
}

// IsValid checks if the token is valid
func (t *Token) IsValid() bool {
	return t.IsActive && !t.IsExpired()
}

// UpdateLastUsed updates the last used timestamp
func (t *Token) UpdateLastUsed() {
	now := time.Now()
	t.LastUsedAt = &now
}

// JWTKey represents a JWT signing key
type JWTKey struct {
	BaseModel
	KeyID     string `json:"key_id" gorm:"uniqueIndex;not null;size:50" validate:"required,min=1,max=50"`
	Algorithm string `json:"algorithm" gorm:"not null;size:20" validate:"required,oneof=HS256 HS384 HS512 RS256 RS384 RS512"`
	Key       string `json:"key" gorm:"type:text;not null" validate:"required,min=1"`
	IsActive  bool   `json:"is_active" gorm:"default:true;index"`
}

// TableName specifies the table name for JWTKey
func (JWTKey) TableName() string {
	return "jwt_keys"
}

// BeforeCreate hook for JWTKey
func (j *JWTKey) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for JWTKey
func (j *JWTKey) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// IsValid checks if the JWT key is valid
func (j *JWTKey) IsValid() bool {
	return j.IsActive
}
