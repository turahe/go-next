package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationUser represents the many-to-many relationship between organizations and users
type OrganizationUser struct {
	BaseModel
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Role           string    `json:"role" gorm:"size:50;default:'member'"` // Role within the organization
	IsActive       bool      `json:"is_active" gorm:"default:true"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	User         User         `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for OrganizationUser
func (OrganizationUser) TableName() string {
	return "organization_users"
}

// BeforeCreate hook for OrganizationUser
func (ou *OrganizationUser) BeforeCreate(tx *gorm.DB) error {
	if ou.ID == uuid.Nil {
		ou.ID = uuid.New()
	}
	return nil
}
