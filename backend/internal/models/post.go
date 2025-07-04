package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID         uint   `gorm:"primaryKey"`
	Title      string `gorm:"not null"`
	Content    string `gorm:"type:text;not null"`
	CategoryID uint   `gorm:"not null"`
	CreatedBy  *uint  `gorm:"not null"`
	UpdatedBy  *uint  `gorm:"not null"`
	DeletedBy  *uint  `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	Category   Category  `gorm:"foreignKey:CategoryID"`
	Comments   []Comment `gorm:"foreignKey:PostID"`
	Contents   []Content `gorm:"foreignKey:modelId"`
	Medias     []Media   `gorm:"foreignKey:modelId"`
}

func (Post) TableName() string {
	return "posts"
}

func (m *Post) BeforeCreate(*gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Post) BeforeUpdate(*gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
