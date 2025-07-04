package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Media struct {
	ID              uint       `gorm:"primaryKey;autoIncrement;column:id"`
	UUID            uuid.UUID  `gorm:"type:uuid;not null;column:uuid"`
	Hash            string     `gorm:"size:255;column:hash"`
	Name            string     `gorm:"size:255;not null;column:name"`
	FileName        string     `gorm:"size:255;not null;column:file_name"`
	Disk            string     `gorm:"size:255;not null;column:disk"`
	MimeType        string     `gorm:"size:255;not null;column:mime_type"`
	Size            int        `gorm:"not null;check:size > 0;column:size"`
	RecordLeft      *int64     `gorm:"column:record_left"`
	RecordRight     *int64     `gorm:"column:record_right"`
	RecordDept      *int64     `gorm:"column:record_dept"`
	RecordOrdering  *int64     `gorm:"column:record_ordering"`
	ParentID        *int64     `gorm:"column:parent_id"`
	CustomAttribute string     `gorm:"size:255;column:custom_attribute"`
	CreatedBy       *int64     `gorm:"column:created_by"`
	UpdatedBy       *int64     `gorm:"column:updated_by"`
	DeletedBy       *int64     `gorm:"column:deleted_by"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
	Mediables       []Mediable `gorm:"foreignKey:MediaID"`
}

func (Media) TableName() string {
	return "media"
}

func (m *Media) BeforeCreate(*gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Media) BeforeUpdate(*gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
