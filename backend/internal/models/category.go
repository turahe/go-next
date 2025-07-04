package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID             uint    `gorm:"primaryKey"`
	Name           string  `gorm:"unique;not null"`
	Description    string  `gorm:"type:text"`
	RecordLeft     *int64  `gorm:"column:record_left"`
	RecordRight    *int64  `gorm:"column:record_right"`
	RecordDept     *int64  `gorm:"column:record_dept"`
	RecordOrdering *int64  `gorm:"column:record_ordering"`
	ParentID       *int64  `gorm:"column:parent_id"`
	Posts          []Post  `gorm:"foreignKey:CategoryID"`
	Medias         []Media `gorm:"foreignKey:modelId"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

func (Category) TableName() string {
	return "categories"
}

func (m *Category) BeforeCreate(*gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Category) BeforeUpdate(*gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
