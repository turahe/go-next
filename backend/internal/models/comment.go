package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID             uint    `gorm:"primaryKey"`
	Content        string  `gorm:"type:text;not null"`
	UserID         uint    `gorm:"not null"`
	User           User    `gorm:"foreignKey:UserID"`
	PostID         uint    `gorm:"not null"`
	Post           Post    `gorm:"foreignKey:PostID"`
	RecordLeft     *int64  `gorm:"column:record_left"`
	RecordRight    *int64  `gorm:"column:record_right"`
	RecordDept     *int64  `gorm:"column:record_dept"`
	RecordOrdering *int64  `gorm:"column:record_ordering"`
	ParentID       *int64  `gorm:"column:parent_id"`
	Medias         []Media `gorm:"foreignKey:modelId"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Comment) TableName() string {
	return "comments"
}

func (c *Comment) BeforeCreate(*gorm.DB) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Comment) BeforeUpdate(*gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}
