package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// BaseModelWithUser extends BaseModel with user tracking fields
type BaseModelWithUser struct {
	BaseModel
	CreatedBy *uuid.UUID `json:"created_by,omitempty" gorm:"type:uuid;index"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty" gorm:"type:uuid;index"`
	DeletedBy *uuid.UUID `json:"deleted_by,omitempty" gorm:"type:uuid;index"`
}

// BaseModelWithOrdering extends BaseModelWithUser with nested set model fields
type BaseModelWithOrdering struct {
	BaseModelWithUser
	RecordLeft     int        `json:"record_left" gorm:"index"`
	RecordRight    int        `json:"record_right" gorm:"index"`
	RecordDept     int        `json:"record_dept" gorm:"index"`
	RecordOrdering int        `json:"record_ordering" gorm:"index"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid;index"`
}

// BeforeCreate hook for BaseModelWithUser
func (b *BaseModelWithUser) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for BaseModelWithUser
func (b *BaseModelWithUser) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// BeforeCreate hook for BaseModelWithOrdering
func (b *BaseModelWithOrdering) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for BaseModelWithOrdering
func (b *BaseModelWithOrdering) BeforeUpdate(tx *gorm.DB) error {
	return nil
}
