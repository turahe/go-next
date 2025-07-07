package models

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	Base
	Username      string     `gorm:"uniqueIndex;not null;size:50" json:"username" validate:"required,min=3,max=50,alphanum"`
	Email         string     `gorm:"uniqueIndex;not null;size:255" json:"email" validate:"required,email"`
	PasswordHash  string     `gorm:"not null;size:255" json:"-"` // Hidden from JSON
	Phone         *string    `gorm:"uniqueIndex;size:20" json:"phone,omitempty" validate:"omitempty,len=10"`
	EmailVerified *time.Time `gorm:"index" json:"email_verified_at,omitempty"`
	PhoneVerified *time.Time `gorm:"index" json:"phone_verified_at,omitempty"`
	IsActive      bool       `gorm:"default:true;index" json:"is_active"`
	LastLoginAt   *time.Time `gorm:"index" json:"last_login_at,omitempty"`

	// Relationships
	Roles    []Role    `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE" json:"roles,omitempty"`
	Posts    []Post    `gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL" json:"posts,omitempty"`
	Comments []Comment `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// BeforeCreate sets timestamps and validates user data
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if err := u.Base.BeforeCreate(tx); err != nil {
		return err
	}

	// Normalize email and username
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Username = strings.ToLower(strings.TrimSpace(u.Username))

	// Set default values
	if u.IsActive == false {
		u.IsActive = true
	}

	return u.validate()
}

// BeforeUpdate validates user data before update
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if err := u.Base.BeforeUpdate(tx); err != nil {
		return err
	}

	// Normalize email and username
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Username = strings.ToLower(strings.TrimSpace(u.Username))

	return u.validate()
}

// validate performs validation on user fields
func (u *User) validate() error {
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}

	if !strings.Contains(u.Email, "@") {
		return errors.New("invalid email format")
	}

	if u.Phone != nil && len(*u.Phone) != 10 {
		return errors.New("phone number must be 10 digits")
	}

	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

// HashPassword hashes the provided password and stores it
func (u *User) HashPassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// MarkEmailVerified marks the user's email as verified
func (u *User) MarkEmailVerified() {
	now := time.Now()
	u.EmailVerified = &now
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

// HasRole checks if the user has a specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (u *User) HasAnyRole(roleNames ...string) bool {
	for _, roleName := range roleNames {
		if u.HasRole(roleName) {
			return true
		}
	}
	return false
}

// IsEmailVerified checks if the user's email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerified != nil
}

// IsPhoneVerified checks if the user's phone is verified
func (u *User) IsPhoneVerified() bool {
	return u.PhoneVerified != nil
}
