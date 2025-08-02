package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	BaseModel
	Username      string     `json:"username" gorm:"uniqueIndex;not null;size:50;check:username ~ '^[a-zA-Z0-9_]+$'" validate:"required,min=3,max=50,alphanum"`
	Email         string     `json:"email" gorm:"uniqueIndex;not null;size:255" validate:"required,email"`
	PasswordHash  string     `json:"-" gorm:"not null;size:255"`
	Phone         string     `json:"phone" gorm:"uniqueIndex;size:20" validate:"omitempty,len=10"`
	EmailVerified *time.Time `json:"email_verified,omitempty" gorm:"index"`
	PhoneVerified *time.Time `json:"phone_verified,omitempty" gorm:"index"`
	IsActive      bool       `json:"is_active" gorm:"default:true;index"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	Roles         []Role     `json:"roles,omitempty" gorm:"many2many:user_roles;constraint:OnDelete:CASCADE"`
	Posts         []Post     `json:"posts,omitempty" gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL"`
	Comments      []Comment  `json:"comments,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook for User
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for User
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// CheckPassword compares the provided password with the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// HashPassword hashes the provided password and stores it
func (u *User) HashPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// IsEmailVerified checks if the user's email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerified != nil
}

// MarkEmailVerified marks the user's email as verified
func (u *User) MarkEmailVerified() {
	now := time.Now()
	u.EmailVerified = &now
}

// IsPhoneVerified checks if the user's phone is verified
func (u *User) IsPhoneVerified() bool {
	return u.PhoneVerified != nil
}

// MarkPhoneVerified marks the user's phone as verified
func (u *User) MarkPhoneVerified() {
	now := time.Now()
	u.PhoneVerified = &now
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// GetIsActive returns the active status
func (u *User) GetIsActive() bool {
	return u.IsActive
}

// Activate activates the user
func (u *User) Activate() {
	u.IsActive = true
}

// Deactivate deactivates the user
func (u *User) Deactivate() {
	u.IsActive = false
}
