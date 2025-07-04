package models

import (
	"time"

	"gorm.io/gorm"
)

type Content struct {
	ID        uint   `gorm:"primaryKey"`
	ModelId   uint   `gorm:"not null"`
	ModelType string `gorm:"not null"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Content) TableName() string {
	return "contents"
}

func (m *Content) BeforeCreate(*gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}
func (m *Content) BeforeUpdate(*gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
