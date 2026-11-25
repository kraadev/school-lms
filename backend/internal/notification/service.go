package notification

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/models"
)

type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

type CreateNotificationRequest struct {
	UserID        uuid.UUID
	Type          string
	Title         string
	Message       string
	ReferenceType *string
	ReferenceID   *uuid.UUID
}

func (s *Service) Create(
	req CreateNotificationRequest,
) (*models.Notification, error) {

	notification := models.Notification{
		UserID:        req.UserID,
		Type:          req.Type,
		Title:         req.Title,
		Message:       req.Message,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		IsRead:        false,
	}

	if err := s.DB.Create(&notification).Error; err != nil {
		return nil, err
	}

	return &notification, nil
}

func (s *Service) NotifyClass(classID uuid.UUID, notifType, title, message string, refType *string, refID *uuid.UUID) error {
	var studentUserIDs []uuid.UUID
	err := s.DB.Table("students").
		Joins("JOIN class_students ON class_students.student_id = students.id").
		Where("class_students.class_id = ?", classID).
		Pluck("students.user_id", &studentUserIDs).Error
	if err != nil {
		return err
	}

	for _, uid := range studentUserIDs {
		_ = s.DB.Create(&models.Notification{
			UserID:        uid,
			Type:          notifType,
			Title:         title,
			Message:       message,
			ReferenceType: refType,
			ReferenceID:   refID,
			IsRead:        false,
		})
	}
	return nil
}

func (s *Service) NotifyUser(userID uuid.UUID, notifType, title, message string, refType *string, refID *uuid.UUID) error {
	return s.DB.Create(&models.Notification{
		UserID:        userID,
		Type:          notifType,
		Title:         title,
		Message:       message,
		ReferenceType: refType,
		ReferenceID:   refID,
		IsRead:        false,
	}).Error
}

func (s *Service) GetUserNotifications(
	userID string,
) ([]models.Notification, error) {

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	var notifications []models.Notification

	err = s.DB.
		Where("user_id = ?", id).
		Order("created_at DESC").
		Find(&notifications).Error

	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (s *Service) GetUnreadCount(
	userID string,
) (int64, error) {

	id, err := uuid.Parse(userID)
	if err != nil {
		return 0, errors.New("user id tidak valid")
	}

	var count int64

	err = s.DB.
		Model(&models.Notification{}).
		Where(
			"user_id = ? AND is_read = ?",
			id,
			false,
		).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) MarkAsRead(
	userID string,
	notificationID string,
) error {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user id tidak valid")
	}

	notificationUUID, err := uuid.Parse(notificationID)
	if err != nil {
		return errors.New("notification id tidak valid")
	}

	var notification models.Notification

	err = s.DB.
		Where(
			"id = ? AND user_id = ?",
			notificationUUID,
			userUUID,
		).
		First(&notification).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("notifikasi tidak ditemukan")
		}

		return err
	}

	if notification.IsRead {
		return nil
	}

	now := time.Now()

	notification.IsRead = true
	notification.ReadAt = &now

	return s.DB.Save(&notification).Error
}

func (s *Service) MarkAllAsRead(
	userID string,
) error {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user id tidak valid")
	}

	now := time.Now()

	return s.DB.
		Model(&models.Notification{}).
		Where(
			"user_id = ? AND is_read = ?",
			userUUID,
			false,
		).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}
