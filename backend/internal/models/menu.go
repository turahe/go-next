package models

import (
	"github.com/google/uuid"
)

type Menu struct {
	BaseModelWithOrdering
	Name        string    `json:"name" gorm:"not null;size:50"`
	Description string    `json:"description" gorm:"size:255"`
	Icon        string    `json:"icon" gorm:"size:50"`
	URL         string    `json:"url" gorm:"size:255"`
	ParentID    uuid.UUID `json:"parent_id" gorm:"type:uuid;index"`
	Children    []Menu    `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Roles       []Role    `json:"roles,omitempty" gorm:"many2many:role_menus;constraint:OnDelete:CASCADE"`
}

func (Menu) TableName() string {
	return "menus"
}
