package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID            uint   `gorm:"primaryKey"`
	Username      string `gorm:"unique;not null"`
	Email         string `gorm:"unique;not null"`
	PasswordHash  string `gorm:"not null"`
	Phone         string `gorm:"unique"`
	EmailVerified *time.Time
	PhoneVerified *time.Time
	Roles         []Role    `gorm:"many2many:user_roles;"`
	Posts         []Post    `gorm:"foreignKey:UserID"`
	Comments      []Comment `gorm:"foreignKey:UserID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (User) TableName() string {
	return "people"
}
func (m *User) BeforeCreate(*gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *User) BeforeUpdate(*gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

func (user *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}

func (user *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	return nil
}
