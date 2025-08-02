package services

import (
	"errors"
	"go-next/internal/models"
	"go-next/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		db: database.DB,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(req *models.NotificationRequest) (*models.Notification, error) {
	notification := &models.Notification{
		UserID:   req.UserID,
		Type:     req.Type,
		Title:    req.Title,
		Message:  req.Message,
		Data:     req.Data,
		Priority: req.Priority,
	}

	if err := s.db.Create(notification).Error; err != nil {
		return nil, err
	}

	return notification, nil
}

// GetUserNotifications retrieves notifications for a specific user
func (s *NotificationService) GetUserNotifications(userID string, limit, offset int) ([]models.Notification, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	var notifications []models.Notification
	query := s.db.Where("user_id = ?", userUUID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&notifications).Error; err != nil {
		return nil, err
	}

	return notifications, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID string) (int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return 0, err
	}

	var count int64
	err = s.db.Model(&models.Notification{}).Where("user_id = ? AND read = ?", userUUID, false).Count(&count).Error
	return count, err
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID string, userID string) error {
	notificationUUID, err := uuid.Parse(notificationID)
	if err != nil {
		return err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationUUID, userUUID).
		Update("read", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("notification not found or unauthorized")
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userUUID, false).
		Update("read", true).Error
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID string, userID string) error {
	notificationUUID, err := uuid.Parse(notificationID)
	if err != nil {
		return err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	result := s.db.Where("id = ? AND user_id = ?", notificationUUID, userUUID).Delete(&models.Notification{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("notification not found or unauthorized")
	}

	return nil
}

// DeleteAllNotifications deletes all notifications for a user
func (s *NotificationService) DeleteAllNotifications(userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return s.db.Where("user_id = ?", userUUID).Delete(&models.Notification{}).Error
}

// CreateSystemNotification creates a notification for all users or specific users
func (s *NotificationService) CreateSystemNotification(notificationType string, title, message, data string, userIDs []string) error {
	var notifications []models.Notification

	for _, userID := range userIDs {
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			continue // Skip invalid user IDs
		}
		notification := models.Notification{
			UserID:   userUUID,
			Type:     notificationType,
			Title:    title,
			Message:  message,
			Data:     data,
			Priority: "normal",
		}
		notifications = append(notifications, notification)
	}

	return s.db.Create(&notifications).Error
}

// GetNotificationByID retrieves a specific notification
func (s *NotificationService) GetNotificationByID(notificationID string, userID string) (*models.Notification, error) {
	notificationUUID, err := uuid.Parse(notificationID)
	if err != nil {
		return nil, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	var notification models.Notification
	err = s.db.Where("id = ? AND user_id = ?", notificationUUID, userUUID).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

var NotificationSvc *NotificationService = NewNotificationService()
