package models

import "time"

type Role struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"unique;not null"`
	Users     []User `gorm:"many2many:user_roles;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
