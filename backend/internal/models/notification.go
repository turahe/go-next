package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeInfo    NotificationType = "info"
)

// Notification represents a user notification
type Notification struct {
	BaseModel
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index" validate:"required"`
	Type     string    `json:"type" gorm:"not null;size:50;index" validate:"required,min=1,max=50"`
	Title    string    `json:"title" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	Message  string    `json:"message" gorm:"type:text;not null" validate:"required,min=1"`
	Data     string    `json:"data" gorm:"type:json"`
	Read     bool      `json:"read" gorm:"default:false;index"`
	Priority string    `json:"priority" gorm:"default:'normal';size:20;index" validate:"oneof=low normal high urgent"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}

// BeforeCreate hook for Notification
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// BeforeUpdate hook for Notification
func (n *Notification) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	n.Read = true
}

// MarkAsUnread marks the notification as unread
func (n *Notification) MarkAsUnread() {
	n.Read = false
}

// IsRead checks if the notification is read
func (n *Notification) IsRead() bool {
	return n.Read
}

// IsHighPriority checks if the notification is high priority
func (n *Notification) IsHighPriority() bool {
	return n.Priority == "high"
}

// IsUrgent checks if the notification is urgent
func (n *Notification) IsUrgent() bool {
	return n.Priority == "urgent"
}

// NotificationRequest represents a notification creation request
type NotificationRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	Type     string    `json:"type" validate:"required,min=1,max=50"`
	Title    string    `json:"title" validate:"required,min=1,max=255"`
	Message  string    `json:"message" validate:"required,min=1"`
	Data     string    `json:"data"`
	Priority string    `json:"priority" validate:"omitempty,oneof=low normal high urgent"`
}

// NotificationResponse represents a notification response
type NotificationResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Data      string    `json:"data,omitempty"`
	Read      bool      `json:"read"`
	Priority  string    `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}
