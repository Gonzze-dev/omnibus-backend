package repository

import (
	"context"

	"gorm.io/gorm"

	"tesina/backend/internal/models"
)

type NotificationRepository interface {
	List(ctx context.Context) ([]models.Notification, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *notificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) List(ctx context.Context) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).Order("date desc").Find(&notifications).Error
	return notifications, err
}
